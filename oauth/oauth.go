// Package oauth provides an HTTP handler to handle OAuth2
// redirect and callback requests for the bebop web app.
package oauth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"

	"github.com/disintegration/bebop/jwt"
	"github.com/disintegration/bebop/store"
)

const (
	stateCookie   = "bebop_oauth_state"
	resultCookie  = "bebop_oauth_result"
	clientTimeout = 10 * time.Second
)

// Config is a configuration of an OAuth handler.
type Config struct {
	Logger     *log.Logger
	UserStore  store.UserStore
	JWTService jwt.Service
	MountURL   string
	CookiePath string
}

// Handler handles oauth2 authentication requests.
type Handler struct {
	*Config
	providers map[string]*provider
	router    chi.Router
}

// New creates a new handler based on the given config.
func New(config *Config) *Handler {
	h := &Handler{Config: config}

	h.providers = make(map[string]*provider)

	h.router = chi.NewRouter()
	h.router.Get("/begin/{provider}", h.handleBegin)
	h.router.Get("/end/{provider}", h.handleEnd)

	return h
}

// AddProvider adds a new provider to oauth handler.
func (h *Handler) AddProvider(name, id, secret string) error {
	pc, ok := providerConfigs[name]
	if !ok {
		return fmt.Errorf("oauth: unknown provider: %q", name)
	}

	if id == "" {
		return fmt.Errorf("oauth: empty client id of provider %q", name)
	}

	if secret == "" {
		return fmt.Errorf("oauth: empty client secret of provider %q", name)
	}

	if h.providers == nil {
		h.providers = make(map[string]*provider)
	}

	h.providers[name] = &provider{
		config: &oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			RedirectURL:  h.MountURL + "/end/" + name,
			Endpoint:     pc.endpoint,
			Scopes:       pc.scopes,
		},
		getUser: pc.getUser,
	}

	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) handleBegin(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers[providerName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	state := h.genState()
	redirectURL := provider.config.AuthCodeURL(state)

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookie,
		Value:    state,
		Path:     h.CookiePath,
		HttpOnly: true,
		Secure:   strings.HasPrefix(h.MountURL, "https"),
		MaxAge:   1 * 60 * 60,
	})

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) handleEnd(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers[providerName]
	if !ok {
		http.NotFound(w, r)
		return
	}

	cookie, err := r.Cookie(stateCookie)
	if err != nil {
		h.handleError(w, "failed to get oauth state cookie: %s", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookie,
		Value:    "",
		Path:     h.CookiePath,
		HttpOnly: true,
		MaxAge:   -1,
	})

	state := cookie.Value
	if state == "" {
		h.handleError(w, "empty oauth state cookie")
		return
	}

	queryState := r.URL.Query().Get("state")
	if queryState != state {
		h.handleError(w, "bad state value")
		return
	}

	queryCode := r.URL.Query().Get("code")
	if queryCode == "" {
		h.handleError(w, "empty code value")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), clientTimeout)
	defer cancel()

	token, err := provider.config.Exchange(ctx, queryCode)
	if err != nil {
		h.handleError(w, "exchange failed: %s", err)
		return
	}
	if !token.Valid() {
		h.handleError(w, "invalid token")
		return
	}

	u, err := provider.getUser(provider.config.Client(ctx, token))
	if err != nil {
		h.handleError(w, "get provider user: %s", err)
		return
	}

	if u.id == "" {
		h.handleError(w, "provider user id is empty")
		return
	}

	var authToken string

	user, err := h.UserStore.GetByAuth(providerName, u.id)
	switch err {
	case nil:
		if user.Blocked {
			h.renderOAuthResult(w, "error:UserBlocked")
			return
		}

		authToken, err = h.JWTService.Create(user.ID)
		if err != nil {
			h.handleError(w, "failed to create auth token: %s", err)
			return
		}

	case store.ErrNotFound:
		userID, err := h.UserStore.New(providerName, u.id)
		if err != nil {
			h.handleError(w, "failed to create user: %s", err)
			return
		}

		authToken, err = h.JWTService.Create(userID)
		if err != nil {
			h.handleError(w, "failed to create auth token: %s", err)
			return
		}

	default:
		h.handleError(w, "failed to get user by auth: %s", err)
		return
	}

	h.renderOAuthResult(w, "success:"+authToken)
}

func (h *Handler) renderOAuthResult(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:   resultCookie,
		Value:  message,
		Path:   h.CookiePath,
		Secure: strings.HasPrefix(h.MountURL, "https"),
		MaxAge: 10 * 60,
	})
	fmt.Fprint(w, `<!doctype html><title>OAuth</title><script>try {opener.bebopOAuthEnd()} finally {window.close()}</script>`)
}

func (h *Handler) handleError(w http.ResponseWriter, format string, a ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	callerNameSplit := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcName := callerNameSplit[len(callerNameSplit)-1]
	h.Logger.Printf("ERROR: %s: %s", funcName, fmt.Sprintf(format, a...))
	h.renderOAuthResult(w, "error:Other")
}

func (h *Handler) genState() string {
	return uuid.NewV4().String()
}
