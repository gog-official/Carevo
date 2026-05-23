package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type SuggestionHandler struct {
	db *sqlx.DB
}

func NewSuggestionHandler(db *sqlx.DB) *SuggestionHandler {
	return &SuggestionHandler{db: db}
}

func (h *SuggestionHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req models.CreateSuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Title == "" || req.Description == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title and description are required"})
		return
	}

	var suggestion models.CareerSuggestion
	err := h.db.QueryRowx(
		`INSERT INTO career_suggestions (user_id, title, description, reason)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, title, description, reason, status, created_at`,
		claims.UserID, req.Title, req.Description, req.Reason,
	).StructScan(&suggestion)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create suggestion"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"suggestion": suggestion})
}

func (h *SuggestionHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var suggestions []models.CareerSuggestion
	if err := h.db.Select(&suggestions, `SELECT id, user_id, title, description, reason, status, created_at FROM career_suggestions WHERE user_id = $1 ORDER BY created_at DESC`, claims.UserID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch suggestions"})
		return
	}

	if suggestions == nil {
		suggestions = []models.CareerSuggestion{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"suggestions": suggestions})
}
