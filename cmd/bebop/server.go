package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"

	"encoding/json"

	"github.com/disintegration/bebop/api"
	"github.com/disintegration/bebop/avatar"
	"github.com/disintegration/bebop/config"
	"github.com/disintegration/bebop/jwt"
	"github.com/disintegration/bebop/oauth"
	"github.com/disintegration/bebop/static"
)

// startServer configures and starts the bebop web server.
func startServer() {
	cfg, err := config.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to load configuration file: %s", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.Fatalf("failed to parse base url: %s", err)
	}

	fileStorage, err := getFileStorage(cfg)
	if err != nil {
		log.Fatalf("failed to init file storage: %s", err)
	}

	store, err := getStore(cfg)
	if err != nil {
		log.Fatalf("failed to init data store: %s", err)
	}

	jwtService, err := jwt.NewService(cfg.JWT.Secret)
	if err != nil {
		log.Fatalf("failed to create jwt service: %s", err)
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

	for providerName, provider := range cfg.OAuth {
		if provider.ClientID != "" && provider.Secret != "" {
			err := oauthHandler.AddProvider(providerName, provider.ClientID, provider.Secret)
			if err != nil {
				log.Fatalf("failed to init oauth provider (%s): %s", providerName, err)
			}
		}
	}

	configHandler, err := newConfigHandler(cfg)
	if err != nil {
		log.Fatalf("failed to create config handler: %s", err)
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

	router.Get("/config", configHandler)
	router.Get("/", static.EmbeddedFile("/frontend/app.html").ServeHTTP)

	log.Printf("starting the server: %s", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, http.StripPrefix(baseURL.Path, router)); err != nil {
		log.Fatalf("listen and serve failed: %v", err)
	}
}

func newConfigHandler(cfg *config.Config) (http.HandlerFunc, error) {
	appConfig := struct {
		Title string   `json:"title"`
		OAuth []string `json:"oauth"`
	}{
		Title: cfg.Title,
		OAuth: []string{},
	}

	for providerName, provider := range cfg.OAuth {
		if provider.ClientID != "" && provider.Secret != "" {
			appConfig.OAuth = append(appConfig.OAuth, providerName)
		}
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
