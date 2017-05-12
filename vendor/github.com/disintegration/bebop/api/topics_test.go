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

func TestHandleGetTopics(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			TopicStore: &mock.TopicStore{
				OnGetLatest: func(offset, limit int) ([]*store.Topic, int, error) {
					if offset == 0 {
						return []*store.Topic{
							{
								ID:            1,
								AuthorID:      1,
								Title:         "Topic1",
								CreatedAt:     testTime,
								LastCommentAt: testTime,
								CommentCount:  10,
							},
							{
								ID:            2,
								AuthorID:      2,
								Title:         "Topic2",
								CreatedAt:     testTime,
								LastCommentAt: testTime,
								CommentCount:  20,
							},
						}, 2, nil
					}
					if offset == 100 {
						return []*store.Topic{}, 2, nil
					}
					t.Fatalf("OnGetByTopic: unexpected params (unknown test)")
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
			wantBody: `{"topics":[{"id":1,"authorId":1,"title":"Topic1","createdAt":"2001-02-03T04:05:06Z","lastCommentAt":"2001-02-03T04:05:06Z","commentCount":10},{"id":2,"authorId":2,"title":"Topic2","createdAt":"2001-02-03T04:05:06Z","lastCommentAt":"2001-02-03T04:05:06Z","commentCount":20}],"count":2}`,
		},
		{
			desc:     "offset 0",
			offset:   "0",
			limit:    "100",
			wantCode: http.StatusOK,
			wantBody: `{"topics":[{"id":1,"authorId":1,"title":"Topic1","createdAt":"2001-02-03T04:05:06Z","lastCommentAt":"2001-02-03T04:05:06Z","commentCount":10},{"id":2,"authorId":2,"title":"Topic2","createdAt":"2001-02-03T04:05:06Z","lastCommentAt":"2001-02-03T04:05:06Z","commentCount":20}],"count":2}`,
		},
		{
			desc:     "offset 100",
			offset:   "100",
			limit:    "100",
			wantCode: http.StatusOK,
			wantBody: `{"topics":[],"count":2}`,
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
		url := "/topics?"
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

func TestHandleNewTopic(t *testing.T) {
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
					}
					return nil, store.ErrNotFound
				},
			},
			TopicStore: &mock.TopicStore{
				OnNew: func(authorID int64, title string) (int64, error) {
					if authorID != 1 || title != "Topic1" {
						t.Fatalf("TopicStore.OnNew: unexpected params: %d, %q", authorID, title)
					}
					return 11, nil
				},
			},
			CommentStore: &mock.CommentStore{
				OnNew: func(topicID int64, authorID int64, content string) (int64, error) {
					if topicID != 11 || authorID != 1 || content != "Comment1" {
						t.Fatalf("CommentStore.OnNew: unexpected params: %d, %d, %q", topicID, authorID, content)
					}
					return 12, nil
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
			body:     `{"title":"Topic1","content":"Comment1"}`,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "blocked user token",
			body:     `{"title":"Topic1","content":"Comment1"}`,
			token:    token2,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "unactivated user token",
			body:     `{"title":"Topic1","content":"Comment1"}`,
			token:    token3,
			wantCode: http.StatusForbidden,
			wantBody: `{"error":{"code":"Forbidden","message":"User name is empty"}}`,
		},
		{
			desc:     "not found user token",
			body:     `{"title":"Topic1","content":"Comment1"}`,
			token:    token100,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "good token",
			body:     `{"title":"Topic1","content":"Comment1"}`,
			token:    token1,
			wantCode: http.StatusCreated,
			wantBody: `{"id":11,"commentId":12}`,
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
			body:     `{"title":"","content":"Comment1"}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid topic title"}}`,
		},
		{
			desc:     "large title",
			body:     `{"title":"` + strings.Repeat("X", 101) + `","content":"Comment1"}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid topic title"}}`,
		},
		{
			desc:     "empty content",
			body:     `{"title":"Topic1","content":""}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid comment content"}}`,
		},
		{
			desc:     "large content",
			body:     `{"title":"Topic1","content":"` + strings.Repeat("X", 10001) + `"}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid comment content"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("POST", "/topics", ioutil.NopCloser(strings.NewReader(tc.body)))
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

func TestHandleGetTopic(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			TopicStore: &mock.TopicStore{
				OnGet: func(id int64) (*store.Topic, error) {
					switch id {
					case 1:
						return &store.Topic{
							ID:            1,
							AuthorID:      1,
							Title:         "Topic1",
							CreatedAt:     testTime,
							LastCommentAt: testTime,
							CommentCount:  10,
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
			wantBody: `{"topic":{"id":1,"authorId":1,"title":"Topic1","createdAt":"2001-02-03T04:05:06Z","lastCommentAt":"2001-02-03T04:05:06Z","commentCount":10}}`,
		},
		{
			desc:     "not found",
			id:       "100",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"Topic not found"}}`,
		},
		{
			desc:     "bad id",
			id:       "BAD_ID",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid topic ID"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/topics/"+tc.id, nil)
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

func TestDeleteTopic(t *testing.T) {
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
			TopicStore: &mock.TopicStore{
				OnGet: func(id int64) (*store.Topic, error) {
					switch id {
					case 1:
						return &store.Topic{
							ID:            1,
							AuthorID:      1,
							Title:         "Topic1",
							CreatedAt:     testTime,
							LastCommentAt: testTime,
							CommentCount:  10,
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
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid topic ID"}}`,
		},
		{
			desc:     "admin token, unknown id",
			token:    token2,
			id:       "100",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"Topic not found"}}`,
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
		req, err := http.NewRequest("DELETE", "/topics/"+tc.id, nil)
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
