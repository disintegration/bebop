package api

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/disintegration/bebop/avatar"
	"github.com/disintegration/bebop/jwt"
	"github.com/disintegration/bebop/store"
	"github.com/disintegration/bebop/store/mock"
)

func TestHandleMe(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	jwtService, err := jwt.NewService(strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	token1, err := jwtService.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	token100, err := jwtService.Create(100)
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
						return &store.User{
							ID:          1,
							Name:        "TestUser1",
							CreatedAt:   testTime,
							AuthService: "TestAuthService1",
							AuthID:      "TestAuthID1",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar1",
						}, nil
					}
					return nil, store.ErrNotFound
				},
			},
		},
		JWTService: jwtService,
		AvatarService: &avatar.MockService{
			OnURL: func(user *store.User) string {
				return "https://example.com/avatars/" + user.Avatar
			},
		},
	})

	tests := []struct {
		desc     string
		token    string
		wantCode int
		wantBody string
	}{
		{
			desc:     "no token",
			token:    "",
			wantCode: http.StatusOK,
			wantBody: `{"authenticated":false}`,
		},
		{
			desc:     "bad token",
			token:    "bad token",
			wantCode: http.StatusOK,
			wantBody: `{"authenticated":false}`,
		},
		{
			desc:     "unknown user",
			token:    token100,
			wantCode: http.StatusOK,
			wantBody: `{"authenticated":false}`,
		},
		{
			desc:     "known user",
			token:    token1,
			wantCode: http.StatusOK,
			wantBody: `{"authenticated":true,"user":{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","authService":"TestAuthService1","blocked":false,"admin":false,"avatar":"https://example.com/avatars/Avatar1"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/me", nil)
		if err != nil {
			t.Fatal(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}

		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		if tc.wantCode != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.wantCode, w.Code)
		}

		if tc.wantBody != w.Body.String() {
			t.Fatalf("test %q: want response body %q got %q", tc.desc, tc.wantBody, w.Body.String())
		}
	}
}

func TestHandleGetUsers(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	jwtService, err := jwt.NewService(strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	token1, err := jwtService.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	token2, err := jwtService.Create(2)
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
						return &store.User{
							ID:          1,
							Name:        "TestUser1",
							CreatedAt:   testTime,
							AuthService: "TestAuthService1",
							AuthID:      "TestAuthID1",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar1",
						}, nil
					case 2:
						return &store.User{
							ID:          2,
							Name:        "TestUser2",
							CreatedAt:   testTime,
							AuthService: "TestAuthService2",
							AuthID:      "TestAuthID2",
							Blocked:     false,
							Admin:       true,
							Avatar:      "Avatar2",
						}, nil
					}
					return nil, store.ErrNotFound
				},
				OnGetMany: func(ids []int64) (map[int64]*store.User, error) {
					users := make(map[int64]*store.User)
					for _, id := range ids {
						switch id {
						case 1:
							users[id] = &store.User{
								ID:          1,
								Name:        "TestUser1",
								CreatedAt:   testTime,
								AuthService: "TestAuthService1",
								AuthID:      "TestAuthID1",
								Blocked:     false,
								Admin:       false,
								Avatar:      "Avatar1",
							}
						case 2:
							users[id] = &store.User{
								ID:          2,
								Name:        "TestUser2",
								CreatedAt:   testTime,
								AuthService: "TestAuthService2",
								AuthID:      "TestAuthID2",
								Blocked:     false,
								Admin:       true,
								Avatar:      "Avatar2",
							}
						default:
							return nil, store.ErrNotFound
						}
					}
					return users, nil
				},
			},
		},
		JWTService: jwtService,
		AvatarService: &avatar.MockService{
			OnURL: func(user *store.User) string {
				return "https://example.com/avatars/" + user.Avatar
			},
		},
	})

	tests := []struct {
		desc     string
		ids      string
		token    string
		wantCode int
		wantBody string
	}{
		{
			desc:     "no token, single user",
			ids:      "1",
			token:    "",
			wantCode: http.StatusOK,
			wantBody: `{"users":[{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","avatar":"https://example.com/avatars/Avatar1"}]}`,
		},
		{
			desc:     "no token, two users",
			ids:      "1,2",
			token:    "",
			wantCode: http.StatusOK,
			wantBody: `{"users":[{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","avatar":"https://example.com/avatars/Avatar1"},{"id":2,"name":"TestUser2","createdAt":"2001-02-03T04:05:06Z","avatar":"https://example.com/avatars/Avatar2"}]}`,
		},
		{
			desc:     "non-admin token",
			ids:      "1,2",
			token:    token1,
			wantCode: http.StatusOK,
			wantBody: `{"users":[{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","avatar":"https://example.com/avatars/Avatar1"},{"id":2,"name":"TestUser2","createdAt":"2001-02-03T04:05:06Z","avatar":"https://example.com/avatars/Avatar2"}]}`,
		},
		{
			desc:     "admin token",
			ids:      "1,2",
			token:    token2,
			wantCode: http.StatusOK,
			wantBody: `{"users":[{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","authService":"TestAuthService1","blocked":false,"admin":false,"avatar":"https://example.com/avatars/Avatar1"},{"id":2,"name":"TestUser2","createdAt":"2001-02-03T04:05:06Z","authService":"TestAuthService2","blocked":false,"admin":true,"avatar":"https://example.com/avatars/Avatar2"}]}`,
		},
		{
			desc:     "bad user id",
			ids:      "1,BAD_ID",
			token:    "",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Parameter 'ids' contains invalid id"}}`,
		},
		{
			desc:     "unknown user",
			ids:      "1,100",
			token:    "",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"User(s) not found"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/users?ids="+tc.ids, nil)
		if err != nil {
			t.Fatal(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}

		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		if tc.wantCode != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.wantCode, w.Code)
		}

		if tc.wantBody != w.Body.String() {
			t.Fatalf("test %q: want response body %q got %q", tc.desc, tc.wantBody, w.Body.String())
		}
	}
}

func TestHandleGetUser(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	jwtService, err := jwt.NewService(strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	token1, err := jwtService.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	token2, err := jwtService.Create(2)
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
						return &store.User{
							ID:          1,
							Name:        "TestUser1",
							CreatedAt:   testTime,
							AuthService: "TestAuthService1",
							AuthID:      "TestAuthID1",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar1",
						}, nil
					case 2:
						return &store.User{
							ID:          2,
							Name:        "TestUser2",
							CreatedAt:   testTime,
							AuthService: "TestAuthService2",
							AuthID:      "TestAuthID2",
							Blocked:     false,
							Admin:       true,
							Avatar:      "Avatar2",
						}, nil
					}
					return nil, store.ErrNotFound
				},
			},
		},
		JWTService: jwtService,
		AvatarService: &avatar.MockService{
			OnURL: func(user *store.User) string {
				return "https://example.com/avatars/" + user.Avatar
			},
		},
	})

	tests := []struct {
		desc     string
		id       string
		token    string
		wantCode int
		wantBody string
	}{
		{
			desc:     "no token",
			id:       "1",
			token:    "",
			wantCode: http.StatusOK,
			wantBody: `{"user":{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","avatar":"https://example.com/avatars/Avatar1"}}`,
		},
		{
			desc:     "non-admin token",
			id:       "1",
			token:    token1,
			wantCode: http.StatusOK,
			wantBody: `{"user":{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","avatar":"https://example.com/avatars/Avatar1"}}`,
		},
		{
			desc:     "admin token",
			id:       "1",
			token:    token2,
			wantCode: http.StatusOK,
			wantBody: `{"user":{"id":1,"name":"TestUser1","createdAt":"2001-02-03T04:05:06Z","authService":"TestAuthService1","blocked":false,"admin":false,"avatar":"https://example.com/avatars/Avatar1"}}`,
		},
		{
			desc:     "bad user id",
			id:       "BAD_ID",
			token:    "",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid user ID"}}`,
		},
		{
			desc:     "unknown user",
			id:       "100",
			token:    "",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"User not found"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/users/"+tc.id, nil)
		if err != nil {
			t.Fatal(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}

		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		if tc.wantCode != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.wantCode, w.Code)
		}

		if tc.wantBody != w.Body.String() {
			t.Fatalf("test %q: want response body %q got %q", tc.desc, tc.wantBody, w.Body.String())
		}
	}
}

func TestHandleSetUserName(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	jwtService, err := jwt.NewService(strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	token1, err := jwtService.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	token2, err := jwtService.Create(2)
	if err != nil {
		t.Fatal(err)
	}

	var userID int64
	var userName string

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			UserStore: &mock.UserStore{
				OnGet: func(id int64) (*store.User, error) {
					switch id {
					case 1:
						return &store.User{
							ID:          1,
							Name:        "TestUser1",
							CreatedAt:   testTime,
							AuthService: "TestAuthService1",
							AuthID:      "TestAuthID1",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar1",
						}, nil
					case 2:
						return &store.User{
							ID:          2,
							Name:        "TestUser2",
							CreatedAt:   testTime,
							AuthService: "TestAuthService2",
							AuthID:      "TestAuthID2",
							Blocked:     false,
							Admin:       true,
							Avatar:      "Avatar2",
						}, nil
					}
					return nil, store.ErrNotFound
				},
				OnGetByName: func(name string) (*store.User, error) {
					if name == "Unavailavle" {
						return &store.User{}, nil
					}
					return nil, store.ErrNotFound
				},
				OnSetName: func(id int64, name string) error {
					if id != 1 && id != 2 {
						return store.ErrNotFound
					}
					if name == "TestError" {
						return errors.New("TestError")
					}
					userID = id
					userName = name
					return nil
				},
			},
		},
		JWTService: jwtService,
	})

	tests := []struct {
		desc     string
		id       string
		token    string
		body     string
		wantCode int
		wantBody string
		wantID   int64
		wantName string
	}{
		{
			desc:     "no token",
			id:       "1",
			token:    "",
			body:     `{"name":"Alice"}`,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "non-admin token, different user",
			id:       "2",
			token:    token1,
			body:     `{"name":"Alice"}`,
			wantCode: http.StatusForbidden,
			wantBody: `{"error":{"code":"Forbidden","message":"Access denied"}}`,
		},
		{
			desc:     "non-admin token, same user",
			id:       "1",
			token:    token1,
			body:     `{"name":"Alice"}`,
			wantCode: http.StatusOK,
			wantBody: `{}`,
			wantID:   1,
			wantName: "Alice",
		},
		{
			desc:     "admin token, different user",
			id:       "1",
			token:    token2,
			body:     `{"name":"Alice"}`,
			wantCode: http.StatusOK,
			wantBody: `{}`,
			wantID:   1,
			wantName: "Alice",
		},
		{
			desc:     "admin token, unknown user",
			id:       "100",
			token:    token2,
			body:     `{"name":"Alice"}`,
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"User not found"}}`,
		},
		{
			desc:     "admin token, bad user id",
			id:       "BAD_ID",
			token:    token2,
			body:     `{"name":"Alice"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid user ID"}}`,
		},
		{
			desc:     "admin token, unavailable user name",
			id:       "1",
			token:    token2,
			body:     `{"name":"Unavailavle"}`,
			wantCode: http.StatusConflict,
			wantBody: `{"error":{"code":"UnavailableUserName","message":"Username is already taken"}}`,
		},
		{
			desc:     "admin token, bad user name",
			id:       "1",
			token:    token2,
			body:     `{"name":"bad user name"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"InvalidUserName","message":"Invalid user name"}}`,
		},
		{
			desc:     "admin token, bad body",
			id:       "1",
			token:    token2,
			body:     `{bad body can't parse}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid request body"}}`,
		},
		{
			desc:     "store error",
			id:       "1",
			token:    token2,
			body:     `{"name":"TestError"}`,
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":{"code":"ServerError","message":"Server error"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("PUT", "/users/"+tc.id+"/name", ioutil.NopCloser(strings.NewReader(tc.body)))
		if err != nil {
			t.Fatal(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}

		userID = 0
		userName = ""

		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		if tc.wantCode != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.wantCode, w.Code)
		}

		if tc.wantBody != w.Body.String() {
			t.Fatalf("test %q: want response body %q got %q", tc.desc, tc.wantBody, w.Body.String())
		}

		if tc.wantID != userID {
			t.Fatalf("test %q: want userID %d got %d", tc.desc, tc.wantID, userID)
		}

		if tc.wantName != userName {
			t.Fatalf("test %q: want userName %q got %q", tc.desc, tc.wantName, userName)
		}
	}
}

func TestHandleSetUserAvatar(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	jwtService, err := jwt.NewService(strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	token1, err := jwtService.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	token2, err := jwtService.Create(2)
	if err != nil {
		t.Fatal(err)
	}

	var userID int64
	var userAvatarData string

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			UserStore: &mock.UserStore{
				OnGet: func(id int64) (*store.User, error) {
					switch id {
					case 1:
						return &store.User{
							ID:          1,
							Name:        "TestUser1",
							CreatedAt:   testTime,
							AuthService: "TestAuthService1",
							AuthID:      "TestAuthID1",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar1",
						}, nil
					case 2:
						return &store.User{
							ID:          2,
							Name:        "TestUser2",
							CreatedAt:   testTime,
							AuthService: "TestAuthService2",
							AuthID:      "TestAuthID2",
							Blocked:     false,
							Admin:       true,
							Avatar:      "Avatar2",
						}, nil
					}
					return nil, store.ErrNotFound
				},
			},
		},
		JWTService: jwtService,
		AvatarService: &avatar.MockService{
			OnSave: func(user *store.User, imageData []byte) error {
				switch string(imageData) {
				case "TestDecodeError":
					return avatar.ErrImageDecode
				case "TestTooSmallError":
					return avatar.ErrImageTooSmall
				case "TestTooLargeError":
					return avatar.ErrImageTooLarge
				}
				userID = user.ID
				userAvatarData = string(imageData)
				return nil
			},
		},
	})

	encode := func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}

	tests := []struct {
		desc           string
		id             string
		token          string
		body           string
		wantCode       int
		wantBody       string
		wantID         int64
		wantAvatarData string
	}{
		{
			desc:     "no token",
			id:       "1",
			token:    "",
			body:     `{"avatar":"` + encode("test-data") + `"}`,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "non-admin token, different user",
			id:       "2",
			token:    token1,
			body:     `{"avatar":"` + encode("test-data") + `"}`,
			wantCode: http.StatusForbidden,
			wantBody: `{"error":{"code":"Forbidden","message":"Access denied"}}`,
		},
		{
			desc:           "non-admin token, same user",
			id:             "1",
			token:          token1,
			body:           `{"avatar":"` + encode("test-data") + `"}`,
			wantCode:       http.StatusOK,
			wantBody:       `{}`,
			wantID:         1,
			wantAvatarData: "test-data",
		},
		{
			desc:           "admin token, different user",
			id:             "1",
			token:          token2,
			body:           `{"avatar":"` + encode("test-data") + `"}`,
			wantCode:       http.StatusOK,
			wantBody:       `{}`,
			wantID:         1,
			wantAvatarData: "test-data",
		},
		{
			desc:     "admin token, unknown user",
			id:       "100",
			token:    token2,
			body:     `{"avatar":"` + encode("test-data") + `"}`,
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"User not found"}}`,
		},
		{
			desc:     "admin token, bad user id",
			id:       "BAD_ID",
			token:    token2,
			body:     `{"avatar":"` + encode("test-data") + `"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid user ID"}}`,
		},
		{
			desc:     "admin token, bad avatar data",
			id:       "1",
			token:    token2,
			body:     `{"avatar":"bad~avatar~data"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid avatar data"}}`,
		},
		{
			desc:     "decode error",
			id:       "1",
			token:    token2,
			body:     `{"avatar":"` + encode("TestDecodeError") + `"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Avatar image decode failed"}}`,
		},
		{
			desc:     "too small error",
			id:       "1",
			token:    token2,
			body:     `{"avatar":"` + encode("TestTooSmallError") + `"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Avatar image too small"}}`,
		},
		{
			desc:     "too large error",
			id:       "1",
			token:    token2,
			body:     `{"avatar":"` + encode("TestTooLargeError") + `"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Avatar image too large"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("PUT", "/users/"+tc.id+"/avatar", ioutil.NopCloser(strings.NewReader(tc.body)))
		if err != nil {
			t.Fatal(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}

		userID = 0
		userAvatarData = ""

		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		if tc.wantCode != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.wantCode, w.Code)
		}

		if tc.wantBody != w.Body.String() {
			t.Fatalf("test %q: want response body %q got %q", tc.desc, tc.wantBody, w.Body.String())
		}

		if tc.wantID != userID {
			t.Fatalf("test %q: want userID %d got %d", tc.desc, tc.wantID, userID)
		}

		if tc.wantAvatarData != userAvatarData {
			t.Fatalf("test %q: want avatar data %q got %q", tc.desc, tc.wantAvatarData, userAvatarData)
		}
	}
}

func TestHandleSetUserBlocked(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}
	jwtService, err := jwt.NewService(strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	token1, err := jwtService.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	token2, err := jwtService.Create(2)
	if err != nil {
		t.Fatal(err)
	}

	var (
		userID      int64
		userBlocked bool
	)

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			UserStore: &mock.UserStore{
				OnGet: func(id int64) (*store.User, error) {
					switch id {
					case 1:
						return &store.User{
							ID:          1,
							Name:        "TestUser1",
							CreatedAt:   testTime,
							AuthService: "TestAuthService1",
							AuthID:      "TestAuthID1",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar1",
						}, nil
					case 2:
						return &store.User{
							ID:          2,
							Name:        "TestUser2",
							CreatedAt:   testTime,
							AuthService: "TestAuthService2",
							AuthID:      "TestAuthID2",
							Blocked:     false,
							Admin:       true,
							Avatar:      "Avatar2",
						}, nil
					}
					return nil, store.ErrNotFound
				},
				OnSetBlocked: func(id int64, blocked bool) error {
					userID = id
					userBlocked = blocked
					return nil
				},
			},
		},
		JWTService: jwtService,
	})

	tests := []struct {
		desc        string
		id          string
		token       string
		body        string
		wantCode    int
		wantBody    string
		wantID      int64
		wantBlocked bool
	}{
		{
			desc:     "no token",
			id:       "1",
			token:    "",
			body:     `{"blocked":true}`,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "non-admin token",
			id:       "2",
			token:    token1,
			body:     `{"blocked":true}`,
			wantCode: http.StatusForbidden,
			wantBody: `{"error":{"code":"Forbidden","message":"Access denied"}}`,
		},
		{
			desc:        "admin token",
			id:          "1",
			token:       token2,
			body:        `{"blocked":true}`,
			wantCode:    http.StatusOK,
			wantBody:    `{}`,
			wantID:      1,
			wantBlocked: true,
		},
		{
			desc:        "admin token, unblock unblocked",
			id:          "1",
			token:       token2,
			body:        `{"blocked":false}`,
			wantCode:    http.StatusOK,
			wantBody:    `{}`,
			wantID:      0,
			wantBlocked: false,
		},
		{
			desc:     "admin token, unknown user",
			id:       "100",
			token:    token2,
			body:     `{"blocked":true}`,
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"User not found"}}`,
		},
		{
			desc:     "admin token, bad user id",
			id:       "BAD_ID",
			token:    token2,
			body:     `{"blocked":true}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid user ID"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("PUT", "/users/"+tc.id+"/blocked", ioutil.NopCloser(strings.NewReader(tc.body)))
		if err != nil {
			t.Fatal(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}

		userID = 0
		userBlocked = false

		w := httptest.NewRecorder()
		apiHandler.ServeHTTP(w, req)

		if tc.wantCode != w.Code {
			t.Fatalf("test %q: want status code %d got %d", tc.desc, tc.wantCode, w.Code)
		}

		if tc.wantBody != w.Body.String() {
			t.Fatalf("test %q: want response body %q got %q", tc.desc, tc.wantBody, w.Body.String())
		}

		if tc.wantID != userID {
			t.Fatalf("test %q: want userID %d got %d", tc.desc, tc.wantID, userID)
		}

		if tc.wantBlocked != userBlocked {
			t.Fatalf("test %q: want avatar data %v got %v", tc.desc, tc.wantBlocked, userBlocked)
		}
	}
}
