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

func TestHandleGetComments(t *testing.T) {
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
							CommentCount:  2,
						}, nil
					}
					return nil, store.ErrNotFound
				},
			},
			CommentStore: &mock.CommentStore{
				OnGetByTopic: func(topicID int64, offset, limit int) ([]*store.Comment, int, error) {
					if topicID != 1 {
						return nil, 0, store.ErrNotFound
					}
					if offset == 0 {
						return []*store.Comment{
							{
								ID:        1,
								TopicID:   1,
								AuthorID:  1,
								Content:   "Comment1",
								CreatedAt: testTime,
							},
							{
								ID:        2,
								TopicID:   1,
								AuthorID:  2,
								Content:   "Comment2",
								CreatedAt: testTime,
							},
						}, 2, nil
					}
					if offset == 100 {
						return []*store.Comment{}, 2, nil
					}
					t.Fatalf("OnGetByTopic: unexpected params (unknown test)")
					return nil, 0, nil
				},
			},
		},
	})

	tests := []struct {
		desc     string
		topicID  string
		offset   string
		limit    string
		wantCode int
		wantBody string
	}{
		{
			desc:     "no offset",
			topicID:  "1",
			wantCode: http.StatusOK,
			wantBody: `{"comments":[{"id":1,"topicId":1,"authorId":1,"content":"Comment1","createdAt":"2001-02-03T04:05:06Z"},{"id":2,"topicId":1,"authorId":2,"content":"Comment2","createdAt":"2001-02-03T04:05:06Z"}],"count":2}`,
		},
		{
			desc:     "offset 0",
			topicID:  "1",
			offset:   "0",
			limit:    "100",
			wantCode: http.StatusOK,
			wantBody: `{"comments":[{"id":1,"topicId":1,"authorId":1,"content":"Comment1","createdAt":"2001-02-03T04:05:06Z"},{"id":2,"topicId":1,"authorId":2,"content":"Comment2","createdAt":"2001-02-03T04:05:06Z"}],"count":2}`,
		},
		{
			desc:     "offset 100",
			topicID:  "1",
			offset:   "100",
			limit:    "100",
			wantCode: http.StatusOK,
			wantBody: `{"comments":[],"count":2}`,
		},
		{
			desc:     "unknown topic",
			topicID:  "2",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"Topic not found"}}`,
		},
		{
			desc:     "bad topic id",
			topicID:  "BAD_TOPIC_ID",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid topic ID"}}`,
		},
		{
			desc:     "bad offset",
			topicID:  "1",
			offset:   "BAD_OFFSET",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid offset"}}`,
		},
		{
			desc:     "bad limit",
			topicID:  "1",
			limit:    "BAD_LIMIT",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid limit"}}`,
		},
	}

	for _, tc := range tests {
		url := "/comments?topic=" + tc.topicID
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

func TestHandleNewComment(t *testing.T) {
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
				OnGet: func(id int64) (*store.Topic, error) {
					switch id {
					case 1:
						return &store.Topic{
							ID:            1,
							AuthorID:      1,
							Title:         "Topic1",
							CreatedAt:     testTime,
							LastCommentAt: testTime,
							CommentCount:  2,
						}, nil
					}
					return nil, store.ErrNotFound
				},
			},
			CommentStore: &mock.CommentStore{
				OnNew: func(topicID int64, authorID int64, content string) (int64, error) {
					if topicID != 1 || authorID != 1 || content != "Comment1" {
						t.Fatalf("OnNew: unexpected params: %d, %d, %q", topicID, authorID, content)
					}
					return 11, nil
				},
				OnGetByTopic: func(topicID int64, offset, limit int) ([]*store.Comment, int, error) {
					return []*store.Comment{}, 10, nil
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
			body:     `{"topic":1,"content":"Comment1"}`,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "blocked user token",
			body:     `{"topic":1,"content":"Comment1"}`,
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
			body:     `{"topic":1,"content":"Comment1"}`,
			token:    token100,
			wantCode: http.StatusUnauthorized,
			wantBody: `{"error":{"code":"Unauthorized","message":"Authentication required"}}`,
		},
		{
			desc:     "good token",
			body:     `{"topic":1,"content":"Comment1"}`,
			token:    token1,
			wantCode: http.StatusCreated,
			wantBody: `{"id":11,"count":10}`,
		},
		{
			desc:     "bad request body",
			body:     `{bad request body}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid request body"}}`,
		},
		{
			desc:     "bad topic id",
			body:     `{"topic":-1,"content":"Comment1"}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid topic ID"}}`,
		},
		{
			desc:     "empty content",
			body:     `{"topic":1,"content":""}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid comment content"}}`,
		},
		{
			desc:     "large content",
			body:     `{"topic":1,"content":"` + strings.Repeat("X", 10001) + `"}`,
			token:    token1,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid comment content"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("POST", "/comments", ioutil.NopCloser(strings.NewReader(tc.body)))
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

func TestHandleGetComment(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2001-02-03T04:05:06Z")
	if err != nil {
		t.Fatal(err)
	}

	apiHandler := New(&Config{
		Logger: log.New(ioutil.Discard, "", 0),
		Store: &mock.Store{
			CommentStore: &mock.CommentStore{
				OnGet: func(id int64) (*store.Comment, error) {
					switch id {
					case 1:
						return &store.Comment{
							ID:        1,
							TopicID:   1,
							AuthorID:  1,
							Content:   "Comment1",
							CreatedAt: testTime,
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
			wantBody: `{"comment":{"id":1,"topicId":1,"authorId":1,"content":"Comment1","createdAt":"2001-02-03T04:05:06Z"}}`,
		},
		{
			desc:     "not found",
			id:       "100",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"Comment not found"}}`,
		},
		{
			desc:     "bad id",
			id:       "BAD_ID",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid comment ID"}}`,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/comments/"+tc.id, nil)
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

func TestDeleteComment(t *testing.T) {
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
			CommentStore: &mock.CommentStore{
				OnGet: func(id int64) (*store.Comment, error) {
					switch id {
					case 1:
						return &store.Comment{
							ID:        1,
							TopicID:   1,
							AuthorID:  1,
							Content:   "Comment1",
							CreatedAt: testTime,
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
			wantBody: `{"error":{"code":"BadRequest","message":"Invalid comment ID"}}`,
		},
		{
			desc:     "admin token, unknown id",
			token:    token2,
			id:       "100",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NotFound","message":"Comment not found"}}`,
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
		req, err := http.NewRequest("DELETE", "/comments/"+tc.id, nil)
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
