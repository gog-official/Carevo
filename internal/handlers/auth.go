package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	var user models.User
	err = h.db.QueryRowx(
		`INSERT INTO users (email, password_hash, name, email_verified) VALUES ($1, $2, $3, false)
		 RETURNING id, email, name, bio, avatar_url, auth_provider, is_admin, email_verified, created_at, updated_at`,
		req.Email, string(hash), req.Name,
	).StructScan(&user)
	if err != nil {
		if isUniqueViolation(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	code := generateCode()
	h.db.Exec(
		`INSERT INTO verification_codes (user_id, code, expires_at) VALUES ($1, $2, $3)`,
		user.ID, code, time.Now().Add(10*time.Minute),
	)

	go func() {
		if err := h.mailer.SendVerificationCode(user.Email, code); err != nil {
			fmt.Printf("failed to send verification email to %s: %v\n", user.Email, err)
		}
	}()

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":             "verification code sent to your email",
		"email":               user.Email,
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
