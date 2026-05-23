package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/guruorgoru/carevo/internal/auth"
	"github.com/guruorgoru/carevo/internal/email"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type VerifyHandler struct {
	db         *sqlx.DB
	mailer     *email.Sender
	jwtService *auth.JWTService
}

func NewVerifyHandler(db *sqlx.DB, mailer *email.Sender, jwtService *auth.JWTService) *VerifyHandler {
	return &VerifyHandler{db: db, mailer: mailer, jwtService: jwtService}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type sendCodeRequest struct {
	Email string `json:"email"`
}

func (h *VerifyHandler) SendCode(w http.ResponseWriter, r *http.Request) {
	var req sendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	// Check pending_users first (new signups that haven't verified yet)
	var pendingExists bool
	h.db.Get(&pendingExists, `SELECT EXISTS(SELECT 1 FROM pending_users WHERE email = $1)`, req.Email)
	if pendingExists {
		code := generateCode()
		_, err := h.db.Exec(
			`UPDATE pending_users SET code = $1, expires_at = $2, updated_at = NOW() WHERE email = $3`,
			code, time.Now().Add(10*time.Minute), req.Email,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate code"})
			return
		}
		go h.mailer.SendVerificationCode(req.Email, code)
		writeJSON(w, http.StatusOK, map[string]string{"message": "verification code sent"})
		return
	}

	// Fallback: check users table for unverified users
	var userID int64
	err := h.db.QueryRow(`SELECT id FROM users WHERE email = $1 AND email_verified = false`, req.Email).Scan(&userID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email exists and is unverified, a code has been sent"})
		return
	}

	code := generateCode()
	_, err = h.db.Exec(
		`INSERT INTO verification_codes (user_id, code, expires_at) VALUES ($1, $2, $3)`,
		userID, code, time.Now().Add(10*time.Minute),
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate code"})
		return
	}

	go h.mailer.SendVerificationCode(req.Email, code)
	writeJSON(w, http.StatusOK, map[string]string{"message": "verification code sent"})
}

type verifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (h *VerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and code are required"})
		return
	}

	// Look up in pending_users (new signups)
	var pending struct {
		ID           int64  `db:"id"`
		Email        string `db:"email"`
		PasswordHash string `db:"password_hash"`
		Name         string `db:"name"`
		Code         string `db:"code"`
	}
	err := h.db.Get(&pending,
		`SELECT id, email, password_hash, name, code FROM pending_users WHERE email = $1 AND expires_at > NOW()`,
		req.Email,
	)
	if err != nil {
		// Fallback: check verification_codes table for existing users
		var userID int64
		err = h.db.Get(&userID, `SELECT id FROM users WHERE email = $1 AND email_verified = false`, req.Email)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}

		var codeID int64
		err = h.db.Get(&codeID,
			`SELECT id FROM verification_codes WHERE user_id = $1 AND code = $2 AND expires_at > NOW() ORDER BY created_at DESC LIMIT 1`,
			userID, req.Code,
		)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired code"})
			return
		}

		_, err = h.db.Exec(`UPDATE users SET email_verified = true, updated_at = NOW() WHERE id = $1`, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to verify email"})
			return
		}
		h.db.Exec(`DELETE FROM verification_codes WHERE user_id = $1`, userID)

		var user models.User
		if err := h.db.Get(&user, `SELECT * FROM users WHERE id = $1`, userID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
			return
		}
		h.writeTokenResponse(w, &user)
		return
	}

	if pending.Code != req.Code {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired code"})
		return
	}

	// Move pending user to users table
	var userID int64
	err = h.db.Get(&userID,
		`INSERT INTO users (email, password_hash, name, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, true, NOW(), NOW())
		 RETURNING id`,
		pending.Email, pending.PasswordHash, pending.Name,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	h.db.Exec(`DELETE FROM pending_users WHERE id = $1`, pending.ID)

	var user models.User
	if err := h.db.Get(&user, `SELECT * FROM users WHERE id = $1`, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}
	h.writeTokenResponse(w, &user)
}

func (h *VerifyHandler) writeTokenResponse(w http.ResponseWriter, user *models.User) {
	access, err := h.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
		return
	}

	raw, hashed, err := h.jwtService.GenerateRefreshToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
		return
	}

	h.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		user.ID, hashed, time.Now().Add(h.jwtService.RefreshTokenTTL()),
	)

	tokens := TokenPair{AccessToken: access, RefreshToken: raw}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "email verified",
		"user":    user,
		"tokens":  tokens,
	})
}

func generateCode() string {
	max := big.NewInt(999999)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}
