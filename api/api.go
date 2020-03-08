package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	"github.com/ahmadmuzakkir/go-sample-api-server-structure/version"
	"github.com/go-chi/chi"
)

const (
	tokenExpiry = 24 * time.Hour
)

type Handler struct {
	logger *log.Logger
	router chi.Router
	store  store.Store
}

func NewHandler(store store.Store, logger *log.Logger) *Handler {
	h := &Handler{
		store:  store,
		logger: logger,
	}

	r := chi.NewRouter()

	// No authentication
	r.Group(func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/register", h.register)
		r.Get("/version/", h.version())
	})

	// Require authentication
	r.Group(func(r chi.Router) {
		r.Use(h.authenticate)

		r.Get("/me", h.getMe())

		r.Get("/", h.getMessages())
		r.Post("/", h.createMessage)

		r.Route("/{id}", func(r chi.Router) {
			r.Use(h.authorizeMessage)

			r.Post("/", h.updateFood())
			r.Delete("/", h.deleteFood)
		})
	})

	h.router = r

	return h
}

func (h *Handler) createMessage(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Content    string  `json:"content"`
		Recipients []int64 `json:"recipients"`
	}

	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			renderError(w, http.StatusBadRequest, "body is empty")
			return
		}

		renderError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Content == "" {
		renderError(w, http.StatusBadRequest, "content is empty")
		return
	}

	if len(req.Recipients) == 0 {
		renderError(w, http.StatusBadRequest, "recipients is empty")
		return
	}

	now := time.Now()

	msg := store.Message{
		Content:         req.Content,
		SenderID:        r.Context().Value("user_id").(int64),
		SentDateTime:    now,
		UpdatedDateTime: now,
	}

	if err := h.store.Message().Create(r.Context(), msg, req.Recipients); err != nil {
		if err == store.ErrDuplicate {
			renderError(w, http.StatusBadRequest, "duplicate productId")
			return
		}

		renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) getMessages() http.HandlerFunc {
	type response struct {
		Messages []*store.Message `json:"messages"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int64)

		messages, err := h.store.Message().Get(r.Context(), userID)
		if err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
		}

		render(w, http.StatusOK, response{
			Messages: messages,
		})
	}
}

func (h *Handler) deleteFood(w http.ResponseWriter, r *http.Request) {
	msg := r.Context().Value("msg").(*store.Message)

	if err := h.store.Message().Delete(r.Context(), msg.ID); err == store.ErrNotFound {
		renderError(w, http.StatusBadRequest, "invalid message id")
		return
	} else if err != nil {
		renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateFood() http.HandlerFunc {
	type request struct {
		Content    string  `json:"content"`
		Recipients []int64 `json:"recipients"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		msg := r.Context().Value("msg").(*store.Message)

		var req request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err == io.EOF {
				renderError(w, http.StatusBadRequest, "body is empty")
				return
			}

			renderError(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.Content == "" {
			renderError(w, http.StatusBadRequest, "content is empty")
			return
		}

		if len(req.Recipients) == 0 {
			renderError(w, http.StatusBadRequest, "recipients is empty")
			return
		}

		msg.Content = req.Content
		msg.UpdatedDateTime = time.Now()

		if err := h.store.Message().Update(r.Context(), *msg, req.Recipients); err == store.ErrNotFound {
			renderError(w, http.StatusBadRequest, "invalid message id")
			return
		} else if err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) authenticate(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if len(bearer) <= 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
			renderError(w, http.StatusUnauthorized, "Missing Authorization Bearer header")
			return
		}

		token, err := h.store.Token().GetUserID(r.Context(), bearer[7:])
		if err == store.ErrNotFound {
			renderError(w, http.StatusUnauthorized, "Invalid token")
			return
		} else if err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if time.Now().After(token.UpdatedAt.Add(tokenExpiry)) {
			renderError(w, http.StatusUnauthorized, "Expired token")
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "user_id", token.UserID))

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (h *Handler) authorizeMessage(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int64)

		var msgID int64

		if val := chi.URLParam(r, "id"); val != "" {
			msgID, _ = strconv.ParseInt(val, 10, 64)
		}

		if msgID == 0 {
			renderError(w, http.StatusBadRequest, "invalid id")
			return
		}

		msg, err := h.store.Message().GetByID(r.Context(), msgID)
		if err == store.ErrNotFound {
			renderError(w, http.StatusBadRequest, "invalid message id")
			return
		} else if err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if msg.SenderID != userID {
			renderError(w, http.StatusForbidden, "not permitted")
			return
		}

		ctx := context.WithValue(r.Context(), "msg", msg)

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(f)
}

func (h *Handler) getMe() http.HandlerFunc {
	type response struct {
		ID       int64  `json:"user_id"`
		Username string `json:"username"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int64)

		// At this point, we can assume the user should exist.
		// So, any error is treated as server error.

		user, err := h.store.User().GetByID(r.Context(), userID)
		if err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
			return
		}

		render(w, http.StatusOK, response{
			ID:       user.ID,
			Username: user.Username,
		})
	}
}

func (h *Handler) version() http.HandlerFunc {
	type response struct {
		BuildTime string `json:"buildTime"`
		Commit    string `json:"commit"`
		Release   string `json:"release"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		render(w, http.StatusOK, response{
			BuildTime: version.BuildTime,
			Commit:    version.Commit,
			Release:   version.Release,
		})
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func render(w http.ResponseWriter, status int, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

func renderError(w http.ResponseWriter, status int, message string) {
	res := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}

	render(w, status, res)
}
