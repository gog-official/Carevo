package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type UserHandler struct {
	db *sqlx.DB
}

func NewUserHandler(db *sqlx.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.db.Get(&user, `SELECT id, email, name, bio, avatar_url, created_at, updated_at FROM users WHERE id = $1`, claims.UserID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
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
		 RETURNING id, email, name, bio, avatar_url, created_at, updated_at`,
		req.Name, req.Bio, req.AvatarURL, claims.UserID,
	).StructScan(&user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}
