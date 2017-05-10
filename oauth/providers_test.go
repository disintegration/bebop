package oauth

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		desc           string
		fn             func(*http.Client) (*user, error)
		responseErr    error
		responseStatus int
		responseBody   string
		wantErr        bool
		wantUser       *user
	}{
		{
			desc:           "google",
			fn:             getGoogleUser,
			responseErr:    nil,
			responseStatus: http.StatusOK,
			responseBody:   `{"id":"123456789012345678901","name":"Google Username","key":"value"}`,
			wantErr:        false,
			wantUser:       &user{id: "123456789012345678901", name: "Google Username"},
		},
		{
			desc:           "facebook",
			fn:             getFacebookUser,
			responseErr:    nil,
			responseStatus: http.StatusOK,
			responseBody:   `{"id":"123456789012345","name":"Facebook Username","key":"value"}`,
			wantErr:        false,
			wantUser:       &user{id: "123456789012345", name: "Facebook Username"},
		},
		{
			desc:           "github",
			fn:             getGithubUser,
			responseErr:    nil,
			responseStatus: http.StatusOK,
			responseBody:   `{"id":1234567,"name":"Github Username","key":"value"}`,
			wantErr:        false,
			wantUser:       &user{id: "1234567", name: "Github Username"},
		},
		{
			desc:           "bad status",
			fn:             getGoogleUser,
			responseErr:    nil,
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"id":"123456789012345678901","name":"Google Username","key":"value"}`,
			wantErr:        true,
			wantUser:       nil,
		},
		{
			desc:           "invalid body",
			fn:             getGoogleUser,
			responseErr:    nil,
			responseStatus: http.StatusOK,
			responseBody:   `{invalid body}`,
			wantErr:        true,
			wantUser:       nil,
		},
		{
			desc:           "round trip error",
			fn:             getGoogleUser,
			responseErr:    errors.New("test error"),
			responseStatus: http.StatusOK,
			responseBody:   `{"id":"123456789012345678901","name":"Google Username","key":"value"}`,
			wantErr:        true,
			wantUser:       nil,
		},
	}

	for _, tc := range tests {
		c := &http.Client{
			Transport: &testTransport{
				responseErr:    tc.responseErr,
				responseStatus: tc.responseStatus,
				responseBody:   tc.responseBody,
			},
		}
		gotUser, gotErr := tc.fn(c)
		if tc.wantErr != (gotErr != nil) {
			t.Fatalf("test %q: wantErr=%v, got error: %v", tc.desc, tc.wantErr, gotErr)
		}
		if !reflect.DeepEqual(tc.wantUser, gotUser) {
			t.Fatalf("test %q: want user %v, got %v", tc.desc, tc.wantUser, gotUser)
		}
	}
}

type testTransport struct {
	responseErr    error
	responseStatus int
	responseBody   string
}

func (t *testTransport) RoundTrip(*http.Request) (*http.Response, error) {
	if t.responseErr != nil {
		return nil, t.responseErr
	}
	return &http.Response{
		StatusCode: t.responseStatus,
		Body:       ioutil.NopCloser(strings.NewReader(t.responseBody)),
	}, nil
}
