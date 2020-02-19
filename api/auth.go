package api

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ahmadmuzakkir/go-sample-api-server-structure/store"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Username == "" {
		renderError(w, http.StatusBadRequest, "username is empty")
		return
	}

	if req.Password == "" {
		renderError(w, http.StatusBadRequest, "password is empty")
		return
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))

	user, _ := h.store.User().GetByUsername(r.Context(), username)
	if user == nil {
		renderError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		renderError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	// Generate token by combining a random 16 bytes and user ID. Then, the token encoded using base64.

	var token [24]byte
	if _, err := rand.Read(token[:16]); err != nil {
		h.logger.Printf("ERROR: %v", errors.WithMessage(err, "login"))

		renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	strconv.AppendInt(token[16:], user.ID, 10)
	hash := sha1.Sum(token[:])
	tokenEncoded := base64.StdEncoding.EncodeToString(hash[:])

	now := time.Now()

	if err := h.store.Token().Create(r.Context(), user.ID, tokenEncoded, now); err != nil {
		h.logger.Printf("ERROR: %v", errors.WithMessage(err, "login"))

		renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := struct {
		Token      string    `json:"token"`
		UserID     int64     `json:"user_id"`
		ExpireTime time.Time `json:"expire_time"`
	}{
		Token:      tokenEncoded,
		UserID:     user.ID,
		ExpireTime: now.Add(tokenExpiry),
	}

	render(w, http.StatusOK, &res)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Username == "" {
		renderError(w, http.StatusBadRequest, "username is empty")
		return
	}

	if req.Password == "" {
		renderError(w, http.StatusBadRequest, "password is empty")
		return
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		renderError(w, http.StatusBadRequest, "bad password")
		return
	}

	err = h.store.User().Create(r.Context(), username, string(passwordHash))
	if err != nil {
		if err == store.ErrDuplicate {
			renderError(w, http.StatusBadRequest, "username already exists")
			return
		}

		h.logger.Printf("ERROR: %v", errors.WithMessage(err, "register"))

		renderError(w, http.StatusInternalServerError, "bad password")
		return
	}

	user, err := h.store.User().GetByUsername(r.Context(), username)
	if err != nil {
		h.logger.Printf("ERROR: %v", errors.WithMessage(err, "register"))

		renderError(w, http.StatusInternalServerError, "bad password")
		return
	}

	res := struct {
		UserID int64 `json:"user_id"`
	}{
		UserID: user.ID,
	}

	render(w, http.StatusCreated, &res)
}
