package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// @Summary      Health check
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	dbOK := true
	if err := h.db.Ping(); err != nil {
		dbOK = false
	}

	status := "ok"
	if !dbOK {
		status = "degraded"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   status,
		"database": dbOK,
	})
}
