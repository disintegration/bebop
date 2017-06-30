// Package api provides an HTTP handler that handles the bebop REST API requests.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-chi/chi"

	"github.com/disintegration/bebop/avatar"
	"github.com/disintegration/bebop/jwt"
	"github.com/disintegration/bebop/store"
)

// Config is an API handler configuration.
type Config struct {
	Logger        *log.Logger
	Store         store.Store
	JWTService    jwt.Service
	AvatarService avatar.Service
}

// Handler handles API requests.
type Handler struct {
	*Config
	router chi.Router
}

// New creates a new handler based on the given config.
func New(config *Config) *Handler {
	h := &Handler{Config: config}

	h.router = chi.NewRouter()

	h.router.Get("/me", h.handleMe)

	h.router.Get("/users", h.handleGetUsers)
	h.router.Get("/users/{id}", h.handleGetUser)
	h.router.Put("/users/{id}/name", h.handleSetUserName)
	h.router.Put("/users/{id}/avatar", h.handleSetUserAvatar)
	h.router.Put("/users/{id}/blocked", h.handleSetUserBlocked)

	h.router.Get("/topics", h.handleGetTopics)
	h.router.Post("/topics", h.handleNewTopic)
	h.router.Get("/topics/{id}", h.handleGetTopic)
	h.router.Delete("/topics/{id}", h.handleDeleteTopic)

	h.router.Get("/comments", h.handleGetComments)
	h.router.Post("/comments", h.handleNewComment)
	h.router.Get("/comments/{id}", h.handleGetComment)
	h.router.Delete("/comments/{id}", h.handleDeleteComment)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) urlParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func (h *Handler) currentUser(r *http.Request) *store.User {
	// Check the "Authorization" header for auth token.
	// The token is prepended with the "Bearer" authentication scheme.
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 7 || strings.ToUpper(authHeader[:6]) != "BEARER" {
		return nil
	}
	token := authHeader[7:]

	userID, _, err := h.JWTService.Verify(token)
	if err != nil {
		return nil
	}

	user, err := h.Store.Users().Get(userID)
	if err != nil {
		if err != store.ErrNotFound {
			h.logError("get user: %s", err)
		}
		return nil
	}

	if user.Blocked {
		return nil
	}

	return user
}

func (h *Handler) parseRequest(r *http.Request, data interface{}) error {
	const maxRequestLen = 16 * 1024 * 1024
	lr := io.LimitReader(r.Body, maxRequestLen)
	return json.NewDecoder(lr).Decode(data)
}

func (h *Handler) render(w http.ResponseWriter, status int, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logError("marshal json: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
}

func (h *Handler) renderError(w http.ResponseWriter, status int, code, message string) {
	response := struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	response.Error.Code = code
	response.Error.Message = message
	h.render(w, status, response)
}

func (h *Handler) logError(format string, a ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	callerNameSplit := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	funcName := callerNameSplit[len(callerNameSplit)-1]
	h.Logger.Printf("ERROR: %s: %s", funcName, fmt.Sprintf(format, a...))
}
