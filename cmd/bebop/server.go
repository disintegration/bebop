package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/disintegration/bebop/api"
	"github.com/disintegration/bebop/avatar"
	"github.com/disintegration/bebop/config"
	"github.com/disintegration/bebop/jwt"
	"github.com/disintegration/bebop/oauth"
	"github.com/disintegration/bebop/static"
)

// startServer configures and starts the bebop web server.
func startServer() {
	cfg, err := getConfig()
	if err != nil {
		logger.Fatalf("failed to load configuration: %s", err)
	}

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		logger.Fatalf("failed to parse base url: %s", err)
	}

	fileStorage, err := getFileStorage(cfg)
	if err != nil {
		logger.Fatalf("failed to init file storage: %s", err)
	}

	store, err := getStore(cfg)
	if err != nil {
		logger.Fatalf("failed to init data store: %s", err)
	}

	jwtService, err := jwt.NewService(cfg.JWT.Secret)
	if err != nil {
		logger.Fatalf("failed to create jwt service: %s", err)
	}

	avatarService := avatar.NewService(store.Users(), fileStorage, logger)

	apiHandler := api.New(&api.Config{
		Logger:        logger,
		Store:         store,
		JWTService:    jwtService,
		AvatarService: avatarService,
	})

	oauthHandler := oauth.New(&oauth.Config{
		Logger:     logger,
		UserStore:  store.Users(),
		JWTService: jwtService,
		MountURL:   baseURL.String() + "/oauth",
		CookiePath: baseURL.Path + "/",
	})

	oauthProviders, err := initOAuthProviders(cfg, oauthHandler)
	if err != nil {
		logger.Fatalf("failed to init oauth providers: %s", err)
	}

	configHandler, err := newConfigHandler(cfg.Title, oauthProviders)
	if err != nil {
		logger.Fatalf("failed to create config handler: %s", err)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger}))
	router.Use(middleware.Recoverer)

	router.Mount("/api/v1", apiHandler)
	router.Mount("/oauth", oauthHandler)

	router.Mount("/static/-", static.Embedded("/static/-"))

	if cfg.FileStorage.Type == "local" {
		router.Mount("/static", static.Dir("/static", cfg.FileStorage.Local.Dir))
	}

	router.Get("/config.json", configHandler)
	router.Get("/", static.EmbeddedFile("/frontend/app.html").ServeHTTP)

	logger.Printf("starting the server: %s", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, http.StripPrefix(baseURL.Path, router)); err != nil {
		logger.Fatalf("listen and serve failed: %v", err)
	}
}

func initOAuthProviders(cfg *config.Config, h *oauth.Handler) ([]string, error) {
	var providers []string

	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.Secret != "" {
		err := h.AddProvider("google", cfg.OAuth.Google.ClientID, cfg.OAuth.Google.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed to init google oauth provider: %s", err)
		}
		providers = append(providers, "google")
	}

	if cfg.OAuth.Facebook.ClientID != "" && cfg.OAuth.Facebook.Secret != "" {
		err := h.AddProvider("facebook", cfg.OAuth.Facebook.ClientID, cfg.OAuth.Facebook.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed to init facebook oauth provider: %s", err)
		}
		providers = append(providers, "facebook")
	}

	if cfg.OAuth.Github.ClientID != "" && cfg.OAuth.Github.Secret != "" {
		err := h.AddProvider("github", cfg.OAuth.Github.ClientID, cfg.OAuth.Github.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed to init github oauth provider: %s", err)
		}
		providers = append(providers, "github")
	}

	return providers, nil
}

func newConfigHandler(title string, oauthProviders []string) (http.HandlerFunc, error) {
	appConfig := struct {
		Title string   `json:"title"`
		OAuth []string `json:"oauth"`
	}{
		Title: title,
		OAuth: oauthProviders,
	}

	sort.Strings(appConfig.OAuth)

	jsonData, err := json.Marshal(appConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal app config json: %s", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	})

	return handler, nil
}
