package api

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/disintegration/bebop/jwt"
	"github.com/disintegration/bebop/store"
	"github.com/disintegration/bebop/store/mock"
)

func TestCurrentUser(t *testing.T) {
	jwtService, err := jwt.NewService(strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	user1 := &store.User{ID: 1, Name: "User1"}
	token1, err := jwtService.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	user2 := &store.User{ID: 2, Name: "User2", Blocked: true}
	token2, err := jwtService.Create(2)
	if err != nil {
		t.Fatal(err)
	}
	token3, err := jwtService.Create(2)
	if err != nil {
		t.Fatal(err)
	}

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			UserStore: &mock.UserStore{
				OnGet: func(id int64) (*store.User, error) {
					switch id {
					case 1:
						return user1, nil
					case 2:
						return user2, nil
					}
					return nil, store.ErrNotFound
				},
			},
		},
		JWTService: jwtService,
	})

	tests := []struct {
		desc     string
		token    string
		wantUser *store.User
	}{
		{
			desc:     "no token",
			token:    "",
			wantUser: nil,
		},
		{
			desc:     "bad token",
			token:    "BAD_TOKEN",
			wantUser: nil,
		},
		{
			desc:     "good token, blocked user",
			token:    token2,
			wantUser: nil,
		},
		{
			desc:     "good token, user not found",
			token:    token3,
			wantUser: nil,
		},
		{
			desc:     "good token, good user",
			token:    token1,
			wantUser: user1,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		gotUser := apiHandler.currentUser(req)
		if !reflect.DeepEqual(gotUser, tc.wantUser) {
			t.Fatalf("test %q: want user %v got %v", tc.desc, tc.wantUser, gotUser)
		}
	}
}

func TestRenderError(t *testing.T) {
	tests := []struct {
		desc          string
		status        int
		code, message string
		wantBody      string
	}{
		{
			desc:     "bad request",
			status:   http.StatusBadRequest,
			code:     "BadRequest",
			message:  "Invalid ID",
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid ID"}}`,
		},
		{
			desc:     "not found",
			status:   http.StatusNotFound,
			code:     "NotFound",
			message:  "User not found",
			wantBody: `{"error":{"code":"NotFound","message":"User not found"}}`,
		},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		h := New(&Config{})
		h.renderError(w, tc.status, tc.code, tc.message)
		if tc.status != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.status, w.Code)
		}
		if tc.wantBody != w.Body.String() {
			t.Fatalf("test %q: want response body %q got %q", tc.desc, tc.wantBody, w.Body.String())
		}
	}
}
