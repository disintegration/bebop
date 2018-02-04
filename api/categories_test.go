package api

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/disintegration/bebop/jwt"
	"github.com/disintegration/bebop/store"
	"github.com/disintegration/bebop/store/mock"
)

func TestHandleGetCategories(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			CategoryStore: &mock.CategoryStore{
				OnGetLatest: func(offset, limit int) ([]*store.Category, int, error) {
					if offset == 0 {
						return []*store.Category{
							{
								ID:          1,
								AuthorID:    1,
								Title:       "Cat1",
								CreatedAt:   testTime,
								LastTopicAt: testTime,
								TopicCount:  10,
							},
							{
								ID:          2,
								AuthorID:    2,
								Title:       "Cat2",
								CreatedAt:   testTime,
								LastTopicAt: testTime,
								TopicCount:  20,
							},
						}, 2, nil
					}
					if offset == 100 {
						return []*store.Category{}, 2, nil
					}
					t.Fatalf("OnGetByCategory: unexpected params (unknown test)")
					return nil, 0, nil
				},
			},
		},
	})

	tests := []struct {
		desc     string
		offset   string
		limit    string
		wantCode int
		wantBody string
	}{
		{
			desc:     "no offset",
			wantCode: http.StatusOK,
			wantBody: `{"categories":[{"id":1,"authorId":1,"title":"Cat1","createdAt":"2001-02-03T04:05:06Z","lastTopicAt":"2001-02-03T04:05:06Z","topicCount":10},{"id":2,"authorId":2,"title":"Cat2","createdAt":"2001-02-03T04:05:06Z","lastTopicAt":"2001-02-03T04:05:06Z","topicCount":20}],"count":2}`,
		},
		{
			desc:     "offset 0",
			offset:   "0",
			limit:    "100",
			wantCode: http.StatusOK,
			wantBody: `{"categories":[{"id":1,"authorId":1,"title":"Cat1","createdAt":"2001-02-03T04:05:06Z","lastTopicAt":"2001-02-03T04:05:06Z","topicCount":10},{"id":2,"authorId":2,"title":"Cat2","createdAt":"2001-02-03T04:05:06Z","lastTopicAt":"2001-02-03T04:05:06Z","topicCount":20}],"count":2}`,
		},
		{
			desc:     "offset 100",
			offset:   "100",
			limit:    "100",
			wantCode: http.StatusOK,
			wantBody: `{"categories":[],"count":2}`,
		},
		{
			desc:     "bad offset",
			offset:   "BAD_OFFSET",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid offset"}}`,
		},
		{
			desc:     "bad limit",
			limit:    "BAD_LIMIT",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid limit"}}`,
		},
	}

	for _, tc := range tests {
		url := "/categories?"
		if tc.offset != "" {
			url += "&offset=" + tc.offset
		}
		if tc.limit != "" {
			url += "&limit=" + tc.limit
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatal(err)
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

func TestHandleNewCategory(t *testing.T) {
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
	token3, err := jwtService.Create(3)
	if err != nil {
		t.Fatal(err)
	}
	token4, err := jwtService.Create(4)
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
							Admin:       true,
							Avatar:      "Avatar1",
						}, nil
					case 2:
						return &store.User{
							ID:          2,
							Name:        "TestUser2",
							CreatedAt:   testTime,
							AuthService: "TestAuthService2",
							AuthID:      "TestAuthID2",
							Blocked:     true,
							Admin:       false,
							Avatar:      "Avatar2",
						}, nil
					case 3:
						return &store.User{
							ID:          3,
							Name:        "",
							CreatedAt:   testTime,
							AuthService: "TestAuthService3",
							AuthID:      "TestAuthID3",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar3",
						}, nil
					case 4:
						return &store.User{
							ID:          1,
							Name:        "TestUser4",
							CreatedAt:   testTime,
							AuthService: "TestAuthService4",
							AuthID:      "TestAuthID4",
							Blocked:     false,
							Admin:       false,
							Avatar:      "Avatar4",
						}, nil
					}
					return nil, store.ErrNotFound
				},
			},
			CategoryStore: &mock.CategoryStore{
				OnNew: func(authorID int64, title string) (int64, error) {
					if authorID != 1 || title != "Cat1" {
						t.Fatalf("CategoryStore.OnNew: unexpected params: %d, %q", authorID, title)
					}
					return 1, nil
				},
			},
			TopicStore: &mock.TopicStore{
				OnNew: func(category, authorID int64, title string) (int64, error) {
					if category != 1 || authorID != 1 || title != "Topic1" {
						t.Fatalf("TopicStore.OnNew: unexpected params: %d, %d, %q", category, authorID, title)
					}
					return 11, nil
				},
			},
		},
		JWTService: jwtService,
	})

	tests := []struct {
		desc     string
		token    string
		body     string
		wantCode int
		wantBody string
	}{
		{
			desc:     "no token",
			token:    "",
			body:     `{"title":"Cat1"}`,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "blocked user token",
			body:     `{"title":"Cat1"}`,
			token:    token2,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "unactivated user token",
			body:     `{"title":"Cat1"}`,
			token:    token3,
			wantCode: http.StatusForbidden,
			wantBody: `{"error":{"code":"Forbidden","message":"User name is empty"}}`,
		},
		{
			desc:     "not an admin",
			body:     `{"title":"Cat1"}`,
			token:    token4,
			wantCode: http.StatusForbidden,
			wantBody: `{"error":{"code":"Forbidden","message":"Access denied"}}`,
		},
		{
			desc:     "not found user token",
			body:     `{"title":"Cat1"}`,
			token:    token100,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "good token",
			body:     `{"title":"Cat1"}`,
			token:    token1,
			wantCode: http.StatusCreated,
			wantBody: `{"id":1}`,
		},
		{
			desc:     "bad request body",
			body:     `{bad request body}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid request body"}}`,
		},
		{
			desc:     "empty title",
			body:     `{"title":""}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid category title"}}`,
		},
		{
			desc:     "large title",
			body:     `{"title":"` + strings.Repeat("X", 101) + `"}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid category title"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("POST", "/categories", ioutil.NopCloser(strings.NewReader(tc.body)))
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

func TestHandleGetCategory(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			CategoryStore: &mock.CategoryStore{
				OnGet: func(id int64) (*store.Category, error) {
					switch id {
					case 1:
						return &store.Category{
							ID:          1,
							AuthorID:    1,
							Title:       "Cat1",
							CreatedAt:   testTime,
							LastTopicAt: testTime,
							TopicCount:  10,
						}, nil
					}
					return nil, store.ErrNotFound
				},
			},
		},
	})

	tests := []struct {
		desc     string
		id       string
		wantCode int
		wantBody string
	}{
		{
			desc:     "found",
			id:       "1",
			wantCode: http.StatusOK,
			wantBody: `{"category":{"id":1,"authorId":1,"title":"Cat1","createdAt":"2001-02-03T04:05:06Z","lastTopicAt":"2001-02-03T04:05:06Z","topicCount":10}}`,
		},
		{
			desc:     "not found",
			id:       "100",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"Category not found"}}`,
		},
		{
			desc:     "bad id",
			id:       "BAD_ID",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid category ID"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/categories/"+tc.id, nil)
		if err != nil {
			t.Fatal(err)
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

func TestDeleteCategory(t *testing.T) {
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
			CategoryStore: &mock.CategoryStore{
				OnGet: func(id int64) (*store.Category, error) {
					switch id {
					case 1:
						return &store.Category{
							ID:          1,
							AuthorID:    1,
							Title:       "Cat1",
							CreatedAt:   testTime,
							LastTopicAt: testTime,
							TopicCount:  10,
						}, nil
					}
					return nil, store.ErrNotFound
				},
				OnDelete: func(id int64) error {
					if id == 1 {
						return nil
					}
					return store.ErrNotFound
				},
			},
		},
		JWTService: jwtService,
	})

	tests := []struct {
		desc     string
		token    string
		id       string
		wantCode int
		wantBody string
	}{
		{
			desc:     "no token",
			token:    "",
			id:       "1",
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "non-admin token",
			token:    token1,
			id:       "1",
			wantCode: http.StatusForbidden,
			wantBody: `{"error":{"code":"Forbidden","message":"Access denied"}}`,
		},
		{
			desc:     "admin token, bad id",
			token:    token2,
			id:       "BAD_ID",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid category ID"}}`,
		},
		{
			desc:     "admin token, unknown id",
			token:    token2,
			id:       "100",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"Category not found"}}`,
		},
		{
			desc:     "admin token, good id",
			token:    token2,
			id:       "1",
			wantCode: http.StatusOK,
			wantBody: `{}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("DELETE", "/categories/"+tc.id, nil)
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
