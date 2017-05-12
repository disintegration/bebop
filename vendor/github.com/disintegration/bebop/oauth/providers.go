package oauth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type provider struct {
	config  *oauth2.Config
	getUser func(c *http.Client, token string) (*user, error)
}

type providerConfig struct {
	endpoint oauth2.Endpoint
	scopes   []string
	getUser  func(c *http.Client, token string) (*user, error)
}

type user struct {
	id   string
	name string
}

var providerConfigs = map[string]providerConfig{
	"google": {
		endpoint: google.Endpoint,
		scopes:   []string{"profile"},
		getUser:  getGoogleUser,
	},
	"facebook": {
		endpoint: facebook.Endpoint,
		scopes:   []string{"public_profile"},
		getUser:  getFacebookUser,
	},
	"github": {
		endpoint: github.Endpoint,
		scopes:   []string{},
		getUser:  getGithubUser,
	},
}

func getGoogleUser(c *http.Client, token string) (*user, error) {
	url := "https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token)

	u := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{}

	err := getJSON(c, url, &u)
	if err != nil {
		return nil, err
	}

	return &user{id: u.ID, name: u.Name}, nil
}

func getFacebookUser(c *http.Client, token string) (*user, error) {
	url := "https://graph.facebook.com/me?fields=id,name&access_token=" + url.QueryEscape(token)

	u := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{}

	err := getJSON(c, url, &u)
	if err != nil {
		return nil, err
	}

	return &user{id: u.ID, name: u.Name}, nil
}

func getGithubUser(c *http.Client, token string) (*user, error) {
	url := "https://api.github.com/user?access_token=" + url.QueryEscape(token)

	u := struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{}

	err := getJSON(c, url, &u)
	if err != nil {
		return nil, err
	}

	return &user{id: strconv.FormatInt(u.ID, 10), name: u.Name}, nil
}

func getJSON(c *http.Client, url string, v interface{}) error {
	response, err := c.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return json.NewDecoder(response.Body).Decode(v)
}
