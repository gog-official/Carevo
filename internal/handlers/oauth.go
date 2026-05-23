package handlers

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/guruorgoru/carevo/internal/auth"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type OAuthHandler struct {
	db         *sqlx.DB
	jwtService *auth.JWTService
}

func NewOAuthHandler(db *sqlx.DB, jwtService *auth.JWTService) *OAuthHandler {
	return &OAuthHandler{db: db, jwtService: jwtService}
}

func (h *OAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.OAuthLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Provider == "" || req.ProviderID == "" || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider, provider_id, and email are required"})
		return
	}

	var user models.User
	err := h.db.Get(&user, `SELECT * FROM users WHERE auth_provider = $1 AND auth_provider_id = $2`, req.Provider, req.ProviderID)

	if err != nil {
		err = h.db.Get(&user, `SELECT * FROM users WHERE email = $1`, req.Email)
		if err == nil {
			_, err = h.db.Exec(`UPDATE users SET auth_provider = $1, auth_provider_id = $2, avatar_url = COALESCE(NULLIF($3, ''), avatar_url), updated_at = NOW() WHERE id = $4`,
				req.Provider, req.ProviderID, req.AvatarURL, user.ID)
			if err != nil {
				log.Printf("link oauth account: %v", err)
			}
		}
	}

	if err != nil {
		passwordHash := randomPasswordHash()

		err = h.db.QueryRowx(
			`INSERT INTO users (email, password_hash, name, avatar_url, auth_provider, auth_provider_id)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING id, email, name, bio, avatar_url, auth_provider, is_admin, created_at, updated_at`,
			req.Email, passwordHash, req.Name, req.AvatarURL, req.Provider, req.ProviderID,
		).StructScan(&user)
		if err != nil {
			if isUniqueViolation(err) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
			return
		}
	}

	pair, err := h.issueTokenPairOAuth(user.ID, user.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user, "tokens": pair})
}

func (h *OAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.db.Get(&user, `SELECT id, email, name, bio, avatar_url, auth_provider, is_admin, created_at, updated_at FROM users WHERE id = $1`, claims.UserID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *OAuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var user models.User
	err := h.db.QueryRowx(
		`UPDATE users SET name = COALESCE(NULLIF($1, ''), name),
		                   bio = COALESCE(NULLIF($2, ''), bio),
		                   avatar_url = COALESCE(NULLIF($3, ''), avatar_url),
		                   updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, email, name, bio, avatar_url, auth_provider, is_admin, created_at, updated_at`,
		req.Name, req.Bio, req.AvatarURL, claims.UserID,
	).StructScan(&user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *OAuthHandler) issueTokenPairOAuth(userID int64, email string) (*models.TokenPair, error) {
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

func randomPasswordHash() string {
	b := make([]byte, 32)
	rand.Read(b)
	hash, _ := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	return string(hash)
}
