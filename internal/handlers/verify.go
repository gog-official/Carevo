package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/guruorgoru/carevo/internal/email"
	"github.com/jmoiron/sqlx"
)

type VerifyHandler struct {
	db     *sqlx.DB
	mailer *email.Sender
}

func NewVerifyHandler(db *sqlx.DB, mailer *email.Sender) *VerifyHandler {
	return &VerifyHandler{db: db, mailer: mailer}
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

	var userID int64
	var emailAddr string
	err := h.db.QueryRow(`SELECT id, email FROM users WHERE email = $1 AND email_verified = false`, req.Email).Scan(&userID, &emailAddr)
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

	if err := h.mailer.SendVerificationCode(emailAddr, code); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send email"})
		return
	}

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

	var userID int64
	err := h.db.Get(&userID, `SELECT id FROM users WHERE email = $1 AND email_verified = false`, req.Email)
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

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}

func generateCode() string {
	max := big.NewInt(999999)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}
