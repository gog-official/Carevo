package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/guruorgoru/carevo/internal/email"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type AdminHandler struct {
	db     *sqlx.DB
	mailer *email.Sender
}

func NewAdminHandler(db *sqlx.DB, mailer *email.Sender) *AdminHandler {
	return &AdminHandler{db: db, mailer: mailer}
}

func (h *AdminHandler) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.UserFromCtx(r.Context())
		if claims == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		var isAdmin bool
		err := h.db.Get(&isAdmin, `SELECT is_admin FROM users WHERE id = $1`, claims.UserID)
		if err != nil || !isAdmin {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *AdminHandler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	query := `SELECT cs.id, cs.user_id, cs.title, cs.description, cs.reason, cs.status, cs.created_at,
		u.name as user_name, u.email as user_email
		FROM career_suggestions cs
		JOIN users u ON u.id = cs.user_id`

	var args []any
	if status != "" {
		query += ` WHERE cs.status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY cs.created_at DESC`

	var suggestions []models.CareerSuggestion
	if err := h.db.Select(&suggestions, query, args...); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch suggestions"})
		return
	}

	if suggestions == nil {
		suggestions = []models.CareerSuggestion{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"suggestions": suggestions})
}

func (h *AdminHandler) UpdateSuggestionStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing suggestion id"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Status != "approved" && req.Status != "rejected" && req.Status != "pending" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be approved, rejected, or pending"})
		return
	}

	var suggestion struct {
		Title    string `db:"title"`
		UserID   int64  `db:"user_id"`
		UserEmail string `db:"user_email"`
	}
	err := h.db.Get(&suggestion,
		`SELECT cs.title, cs.user_id, u.email as user_email FROM career_suggestions cs JOIN users u ON u.id = cs.user_id WHERE cs.id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "suggestion not found"})
		return
	}

	_, err = h.db.Exec(`UPDATE career_suggestions SET status = $1 WHERE id = $2`, req.Status, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update suggestion"})
		return
	}

	if h.mailer != nil {
		if req.Status == "approved" {
			h.mailer.SendSuggestionApproved(suggestion.UserEmail, suggestion.Title)
		} else if req.Status == "rejected" {
			h.mailer.SendSuggestionRejected(suggestion.UserEmail, suggestion.Title)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": req.Status})
}

func (h *AdminHandler) DeleteSuggestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing suggestion id"})
		return
	}

	_, err := h.db.Exec(`DELETE FROM career_suggestions WHERE id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete suggestion"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *AdminHandler) PromoteToAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.db.Exec(`UPDATE users SET is_admin = true WHERE email = $1`, req.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to promote user"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "promoted"})
}
