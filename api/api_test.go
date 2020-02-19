package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store/mock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TODO add more unit tests

func TestLogin(t *testing.T) {
	user := getUser(t, "username", "password", 1)

	mockStore := &mock.Store{
		UserStore: &mock.UserStore{
			OnGetByUsername: func(ctx context.Context, username string) (*store.User, error) {
				if user.Username == username {
					return user, nil
				}

				return nil, store.ErrNotFound
			},
		},
	}

	type req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	handler := NewHandler(mockStore, nil)

	tests := []struct {
		name     string
		req      req
		wantCode int
	}{
		{
			name: "empty request",
			req: req{
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "user not exists",
			req: req{
				Username: "username_not_exists",
				Password: "password",
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "wrong password",
			req: req{
				Username: user.Username,
				Password: "wrong password",
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "success",
			req: req{
				Username: user.Username,
				Password: "password",
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.req)
			assert.NoError(t, err)

			request := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			request.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			_, err = ioutil.ReadAll(w.Body)
			assert.Nil(t, err)

			assert.Equal(t, tc.wantCode, w.Code, "status code")
		})
	}
}

// Test the authentication checking on all the routes that require authentication.
func TestRequireAuthenticateRoutes(t *testing.T) {
	mockStore := &mock.Store{
		TokenStore: &mock.TokenStore{
			OnCreate: nil,
			OnGetUserID: func(ctx context.Context, token string) (*store.Token, error) {
				if token == "exists" {
					return &store.Token{
						UserID:    1,
						UpdatedAt: time.Now(),
					}, nil
				}

				return nil, store.ErrNotFound
			},
		},
	}

	handler := NewHandler(mockStore, nil)

	tests := []struct {
		url    string
		method string
	}{
		{
			url:    "/",
			method: "POST",
		},
		{
			url:    "/1",
			method: "POST",
		},
		{
			url:    "/1",
			method: "GET",
		},
		{
			url:    "/",
			method: "GET",
		},
		{
			url:    "/1",
			method: "DELETE",
		},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			// Test empty token
			{
				request := httptest.NewRequest(tc.method, tc.url, nil)

				w := httptest.NewRecorder()

				handler.ServeHTTP(w, request)

				_, err := ioutil.ReadAll(w.Body)
				assert.Nil(t, err)

				assert.NotEqual(t, http.StatusNotFound, w.Code, "status code")
				assert.Equal(t, http.StatusUnauthorized, w.Code, "status code")
			}

			// Test token user does not exist
			{
				request := httptest.NewRequest(tc.method, tc.url, nil)
				request.Header.Add("Authorization", "Bearer notexists")

				w := httptest.NewRecorder()

				handler.ServeHTTP(w, request)

				_, err := ioutil.ReadAll(w.Body)
				assert.Nil(t, err)

				assert.NotEqual(t, http.StatusNotFound, w.Code, "status code")
				assert.Equal(t, http.StatusUnauthorized, w.Code, "status code")
			}

			// Test success
			{
				request := httptest.NewRequest(tc.method, tc.url, nil)
				request.Header.Add("Authorization", "Bearer exists")

				w := httptest.NewRecorder()

				handler.ServeHTTP(w, request)

				_, err := ioutil.ReadAll(w.Body)
				assert.Nil(t, err)

				assert.NotEqual(t, http.StatusNotFound, w.Code, "status code")
				assert.NotEqual(t, http.StatusUnauthorized, w.Code, "status code")
			}
		})
	}
}

func TestGetMessages(t *testing.T) {
	messages := []*store.Message{{
		ID:              1,
		Content:         "content",
		Sender:          "sender",
		SentDateTime:    time.Now(),
		UpdatedDateTime: time.Now(),
	}}

	mockStore := &mock.Store{
		MessageStore: &mock.MessageStore{
			OnGet: func(ctx context.Context, userID int64) ([]*store.Message, error) {
				if userID == 2 {
					return messages, nil
				}
				return nil, nil
			},
		},
		TokenStore: &mock.TokenStore{
			OnCreate: nil,
			OnGetUserID: func(ctx context.Context, token string) (*store.Token, error) {
				if token == "empty" {
					return &store.Token{
						UserID:    1,
						UpdatedAt: time.Now(),
					}, nil
				}

				if token == "not empty" {
					return &store.Token{
						UserID:    2,
						UpdatedAt: time.Now(),
					}, nil
				}

				return nil, store.ErrNotFound
			},
		},
	}

	type res struct {
		Messages []*store.Message `json:"messages"`
	}

	handler := NewHandler(mockStore, nil)

	tests := []struct {
		name     string
		token    string
		wantCode int
		response interface{}
	}{
		{
			name:     "empty",
			token:    "empty",
			wantCode: http.StatusOK,
			response: res{
				Messages: nil,
			},
		},
		{
			name:     "not empty",
			token:    "not empty",
			wantCode: http.StatusOK,
			response: res{
				Messages: messages,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/", nil)
			request.Header.Add("Content-Type", "application/json")
			if tc.token != "" {
				request.Header.Add("Authorization", "Bearer "+tc.token)
			}

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			actualJSON, err := ioutil.ReadAll(w.Body)
			assert.Nil(t, err)

			assert.Equal(t, tc.wantCode, w.Code, "status code")

			expectedJSON, err := json.Marshal(tc.response)
			assert.NoError(t, err)

			assert.NoError(t, compareJSON(expectedJSON, actualJSON))
		})
	}
}

func compareJSON(expected []byte, response []byte) error {
	if bytes.Equal(bytes.TrimSpace(response), expected) {
		return nil
	}

	return fmt.Errorf("expected response `%s`, found `%s`", string(expected), string(response))
}

func getUser(t *testing.T, username, password string, id int) *store.User {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	return &store.User{
		ID:           1,
		Username:     username,
		PasswordHash: string(passwordHash),
	}
}
