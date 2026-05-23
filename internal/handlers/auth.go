package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/guruorgoru/carevo/internal/auth"
	"github.com/guruorgoru/carevo/internal/email"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db         *sqlx.DB
	jwtService *auth.JWTService
	mailer     *email.Sender
}

func NewAuthHandler(db *sqlx.DB, jwtService *auth.JWTService, mailer *email.Sender) *AuthHandler {
	return &AuthHandler{db: db, jwtService: jwtService, mailer: mailer}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	// Check if user already exists (verified)
	var exists bool
	h.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, req.Email)
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	code := generateCode()

	// Upsert into pending_users (in case they already have a pending but unverified registration)
	_, err = h.db.Exec(
		`INSERT INTO pending_users (email, password_hash, name, code, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (email) DO UPDATE SET
		   password_hash = EXCLUDED.password_hash,
		   name = EXCLUDED.name,
		   code = EXCLUDED.code,
		   expires_at = EXCLUDED.expires_at,
		   updated_at = NOW()`,
		req.Email, string(hash), req.Name, code, time.Now().Add(10*time.Minute),
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create registration"})
		return
	}

	go h.mailer.SendVerificationCode(req.Email, code)

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":             "verification code sent to your email",
		"email":               req.Email,
		"requires_verification": true,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var user models.User
	err := h.db.Get(&user, `SELECT * FROM users WHERE email = $1`, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	if !user.EmailVerified {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error":               "please verify your email first",
			"requires_verification": true,
			"email":               user.Email,
		})
		return
	}

	pair, err := h.issueTokenPair(user.ID, user.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user, "tokens": pair})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
		return
	}

	tokenHash := auth.HashToken(req.RefreshToken)

	var stored models.RefreshToken
	err := h.db.Get(&stored, `SELECT * FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if time.Now().After(stored.ExpiresAt) {
		h.db.Exec(`DELETE FROM refresh_tokens WHERE id = $1`, stored.ID)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "refresh token expired"})
		return
	}

	_, err = h.db.Exec(`DELETE FROM refresh_tokens WHERE id = $1`, stored.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to rotate token"})
		return
	}

	var user models.User
	if err := h.db.Get(&user, `SELECT * FROM users WHERE id = $1`, stored.UserID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "user not found"})
		return
	}

	pair, err := h.issueTokenPair(user.ID, user.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"tokens": pair})
}

func (h *AuthHandler) issueTokenPair(userID int64, email string) (*models.TokenPair, error) {
	access, err := h.jwtService.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	raw, hashed, err := h.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	_, err = h.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hashed, time.Now().Add(h.jwtService.RefreshTokenTTL()),
	)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{AccessToken: access, RefreshToken: raw}, nil
}
