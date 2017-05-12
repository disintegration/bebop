package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"

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

	indexHandler, err := newIndexHandler(cfg)
	if err != nil {
		log.Fatalf("failed to create index handler: %s", err)
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

	router.Get("/", indexHandler)

	log.Printf("starting the server: %s", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, http.StripPrefix(baseURL.Path, router)); err != nil {
		log.Fatalf("listen and serve failed: %v", err)
	}
}

func newIndexHandler(cfg *config.Config) (http.HandlerFunc, error) {
	t, err := template.New("indexTemplate").Parse(indexTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %s", err)
	}

	appConfig := struct {
		Title string   `json:"title"`
		OAuth []string `json:"oauth"`
	}{
		Title: cfg.Title,
		OAuth: []string{},
	}

	for providerName := range cfg.OAuth {
		appConfig.OAuth = append(appConfig.OAuth, providerName)
	}
	sort.Strings(appConfig.OAuth)

	buf := new(bytes.Buffer)
	err = t.Execute(buf, map[string]interface{}{
		"title":     cfg.Title,
		"appConfig": appConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute index template: %s", err)
	}

	indexData := buf.Bytes()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(indexData)
	})

	return handler, nil
}

var indexTemplate = `<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <meta http-equiv="x-ua-compatible" content="ie=edge">
    <title>{{.title}}</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha256-916EbMg70RQy9LHiGkXzG8hSg9EdNy97GazNG/aiY1w=" crossorigin="anonymous" />
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-markdown/2.10.0/css/bootstrap-markdown.min.css" integrity="sha256-umMZCcE/LUcJ3F3V/D6NmvQxdm3OWtRMiMApkNnDIOw=" crossorigin="anonymous" />
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css" integrity="sha256-eZrrJcwDc/3uDhsdt61sL2oOBY362qM3lon1gyExkL0=" crossorigin="anonymous" />
    <link rel="stylesheet" href="static/-/frontend/css/bebop.css">
  </head>
  <body> 
    <div id="app"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.2.1/jquery.min.js" integrity="sha256-hwg4gsxgFZhOsEEamdOYGBf13FyQuiTwlAQgxVSNgt4=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha256-U5ZEeKfGNOja007MMD3YBI0A3OSZOQbeG6z2f2Y0hu8=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/vue/2.2.6/vue.min.js" integrity="sha256-cWZZjnj99rynB+b8FaNGUivxc1kJSRa8ZM/E77cDq0I=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/vue-router/2.4.0/vue-router.min.js" integrity="sha256-fxzMMjPZbIwP33mgE/4GTQ9BTPM7X1PBAHaJ3Kvz6fo=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/vue-resource/1.3.1/vue-resource.min.js" integrity="sha256-vLNsWeWD+1TzgeVJX92ft87XtRoH3UVqKwbfB2nopMY=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/marked/0.3.6/marked.min.js" integrity="sha256-mJAzKDq6kSoKqZKnA6UNLtPaIj8zT2mFnWu/GSouhgQ=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-markdown/2.10.0/js/bootstrap-markdown.min.js" integrity="sha256-vT9X0tmmfKfNTg0U/Iv0rM9mhu8LA0MaDFrzIflHN9A=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/moment.js/2.18.1/moment.min.js" integrity="sha256-1hjUhpc44NwiNg8OwMu2QzJXhD8kcj+sJA3aCQZoUjg=" crossorigin="anonymous"></script>
    <script>var appConfig = {{.appConfig}}</script>
    <script src="static/-/frontend/js/bebop-init.js"></script>
    <script src="static/-/frontend/js/bebop-nav.js"></script>
    <script src="static/-/frontend/js/bebop-username-modal.js"></script>
    <script src="static/-/frontend/js/bebop-topics.js"></script>
    <script src="static/-/frontend/js/bebop-new-topic.js"></script>
    <script src="static/-/frontend/js/bebop-comments.js"></script>
    <script src="static/-/frontend/js/bebop-new-comment.js"></script>
    <script src="static/-/frontend/js/bebop-user.js"></script>
    <script src="static/-/frontend/js/bebop-app.js"></script>
  </body>
</html>
`
