package oauth

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestOAuthBegin(t *testing.T) {
	handler := New(&Config{
		Logger:     log.New(ioutil.Discard, "", 0),
		MountURL:   "https://example.test/forum/oauth",
		CookiePath: "/forum/",
	})
	handler.providers = map[string]*provider{
		"testprovider": {
			config: &oauth2.Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://provider.test/auth",
					TokenURL: "https://provider.test/token",
				},
				RedirectURL: "https://example.test/forum/oauth/end/testprovider",
				Scopes:      []string{"testscope1", "testscope2"},
			},
		},
	}

	tests := []struct {
		desc          string
		provider      string
		wantCode      int
		checkLocation bool
	}{
		{
			desc:          "test provider",
			provider:      "testprovider",
			wantCode:      http.StatusFound,
			checkLocation: true,
		},
		{
			desc:          "unknown provider",
			provider:      "unknownprovider",
			wantCode:      http.StatusNotFound,
			checkLocation: false,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/begin/"+tc.provider, nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if tc.wantCode != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.wantCode, w.Code)
		}

		if !tc.checkLocation {
			continue
		}

		loc := w.Header().Get("Location")
		locURL, err := url.Parse(loc)
		if err != nil {
			t.Fatal(err)
		}

		wantPrefix := handler.providers[tc.provider].config.Endpoint.AuthURL
		if !strings.HasPrefix(loc, wantPrefix) {
			t.Fatalf("test %q: bad location prefix: location: %q, want prefix: %q", tc.desc, loc, wantPrefix)
		}

		state := locURL.Query().Get("state")
		if state == "" {
			t.Fatalf("test %q: empty location state: %q", tc.desc, state)
		}

		redirectURI := locURL.Query().Get("redirect_uri")
		wantRedirectURI := handler.providers[tc.provider].config.RedirectURL
		if redirectURI != wantRedirectURI {
			t.Fatalf("test %q: bad location redirect URI: want %q got %q", tc.desc, wantRedirectURI, redirectURI)
		}
	}
}
