package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/guruorgoru/carevo/internal/cache"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type FeatureHandler struct {
	db    *sqlx.DB
	cache *cache.Store
}

func NewFeatureHandler(db *sqlx.DB, cacheStore *cache.Store) *FeatureHandler {
	return &FeatureHandler{db: db, cache: cacheStore}
}

// @Summary      Bookmark a career
// @Tags         bookmarks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.BookmarkRequest  true  "Career ID"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /users/me/bookmarks [post]
func (h *FeatureHandler) BookmarkCareer(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req models.BookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var exists bool
	h.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM careers WHERE id = $1)`, req.CareerID)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "career not found"})
		return
	}

	_, err := h.db.Exec(
		`INSERT INTO bookmarks (user_id, career_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		claims.UserID, req.CareerID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to bookmark career"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "bookmarked"})
}

// @Summary      List bookmarked careers
// @Tags         bookmarks
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Router       /users/me/bookmarks [get]
func (h *FeatureHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var bookmarks []models.Bookmark
	err := h.db.Select(&bookmarks, `
		SELECT b.user_id, b.career_id, b.created_at, c.title, c.slug, c.summary
		FROM bookmarks b
		JOIN careers c ON c.id = b.career_id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC`, claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if bookmarks == nil {
		bookmarks = []models.Bookmark{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"bookmarks": bookmarks})
}

// @Summary      Start a 30-day challenge
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.CreateChallengeRequest  true  "Career ID"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /challenges [post]
func (h *FeatureHandler) StartChallenge(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req models.CreateChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var exists bool
	h.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM careers WHERE id = $1)`, req.CareerID)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "career not found"})
		return
	}

	var active int
	h.db.Get(&active, `SELECT COUNT(*) FROM challenges WHERE user_id = $1 AND is_active = TRUE`, claims.UserID)
	if active > 0 {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "you already have an active challenge"})
		return
	}

	var challenge models.Challenge
	err := h.db.QueryRowx(
		`INSERT INTO challenges (user_id, career_id) VALUES ($1, $2)
		 RETURNING id, user_id, career_id, started_at, current_streak, longest_streak, is_active`,
		claims.UserID, req.CareerID,
	).StructScan(&challenge)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start challenge"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"challenge": challenge})
}

// @Summary      Get active challenge
// @Tags         challenges
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Router       /challenges/active [get]
func (h *FeatureHandler) ActiveChallenge(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var challenge models.Challenge
	err := h.db.Get(&challenge, `
		SELECT ch.*, c.title as career_title, c.slug as career_slug
		FROM challenges ch
		JOIN careers c ON c.id = ch.career_id
		WHERE ch.user_id = $1 AND ch.is_active = TRUE
		ORDER BY ch.started_at DESC
		LIMIT 1`, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no active challenge"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"challenge": challenge})
}

// @Summary      Daily check-in for a challenge
// @Tags         challenges
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Challenge ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /challenges/{id}/checkin [post]
func (h *FeatureHandler) DailyCheckin(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	challengeID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
		return
	}

	var challenge models.Challenge
	err = h.db.Get(&challenge, `SELECT * FROM challenges WHERE id = $1`, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "challenge not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if challenge.UserID != claims.UserID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "not your challenge"})
		return
	}

	if !challenge.IsActive {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "challenge is already completed"})
		return
	}

	today := time.Now().Truncate(24 * time.Hour)

	var alreadyCheckedIn bool
	h.db.Get(&alreadyCheckedIn, `SELECT EXISTS(SELECT 1 FROM challenge_checkins WHERE challenge_id = $1 AND checkin_date = $2)`, challengeID, today)
	if alreadyCheckedIn {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already checked in today"})
		return
	}

	_, err = h.db.Exec(`INSERT INTO challenge_checkins (challenge_id, checkin_date) VALUES ($1, $2)`, challengeID, today)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to check in"})
		return
	}

	yesterday := today.Add(-24 * time.Hour)
	var lastCheckin sql.NullTime
	h.db.Get(&lastCheckin, `SELECT MAX(checkin_date) FROM challenge_checkins WHERE challenge_id = $1 AND checkin_date < $2`, challengeID, today)

	var newStreak int
	if lastCheckin.Valid && lastCheckin.Time.Equal(yesterday) {
		newStreak = challenge.CurrentStreak + 1
	} else {
		newStreak = 1
	}

	newLongest := challenge.LongestStreak
	if newStreak > newLongest {
		newLongest = newStreak
	}

	isActive := newStreak < 30

	_, err = h.db.Exec(
		`UPDATE challenges SET current_streak = $1, longest_streak = $2, is_active = $3 WHERE id = $4`,
		newStreak, newLongest, isActive, challengeID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update challenge"})
		return
	}

	var checkinCount int
	h.db.Get(&checkinCount, `SELECT COUNT(*) FROM challenge_checkins WHERE challenge_id = $1`, challengeID)

	writeJSON(w, http.StatusOK, map[string]any{
		"checked_in":       true,
		"current_streak":   newStreak,
		"longest_streak":   newLongest,
		"days_completed":   checkinCount,
		"challenge_active": isActive,
	})
}

// @Summary      Get challenge progress
// @Tags         challenges
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Challenge ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /challenges/{id}/progress [get]
func (h *FeatureHandler) ChallengeProgress(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	challengeID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid challenge id"})
		return
	}

	var challenge models.Challenge
	err = h.db.Get(&challenge, `SELECT * FROM challenges WHERE id = $1`, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "challenge not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if challenge.UserID != claims.UserID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "not your challenge"})
		return
	}

	var daysCompleted int
	h.db.Get(&daysCompleted, `SELECT COUNT(*) FROM challenge_checkins WHERE challenge_id = $1`, challengeID)

	pct := (daysCompleted * 100) / 30
	remaining := 30 - daysCompleted
	if remaining < 0 {
		remaining = 0
	}

	progress := models.ChallengeProgress{
		ChallengeID:     challengeID,
		DaysCompleted:   daysCompleted,
		CurrentStreak:   challenge.CurrentStreak,
		LongestStreak:   challenge.LongestStreak,
		PercentComplete: pct,
		DaysRemaining:   remaining,
	}

	writeJSON(w, http.StatusOK, map[string]any{"progress": progress})
}

// @Summary      Submit a project idea
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  int                                 true  "Career ID"
// @Param        body  body  models.CreateProjectIdeaRequest     true  "Project idea"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /careers/{id}/projects [post]
func (h *FeatureHandler) CreateProjectIdea(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	careerID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid career id"})
		return
	}

	var exists bool
	h.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM careers WHERE id = $1)`, careerID)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "career not found"})
		return
	}

	var req models.CreateProjectIdeaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	var idea models.ProjectIdea
	err = h.db.QueryRowx(
		`INSERT INTO project_ideas (career_id, user_id, title, description) VALUES ($1, $2, $3, $4)
		 RETURNING id, career_id, user_id, title, description, created_at`,
		careerID, claims.UserID, req.Title, req.Description,
	).StructScan(&idea)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create project idea"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"project_idea": idea})
}

// @Summary      Browse project ideas for a career
// @Tags         projects
// @Produce      json
// @Param        id   path  int  true  "Career ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Router       /careers/{id}/projects [get]
func (h *FeatureHandler) ListProjectIdeas(w http.ResponseWriter, r *http.Request) {
	careerID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid career id"})
		return
	}

	var exists bool
	h.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM careers WHERE id = $1)`, careerID)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "career not found"})
		return
	}

	var ideas []models.ProjectIdea
	err = h.db.Select(&ideas, `SELECT * FROM project_ideas WHERE career_id = $1 ORDER BY created_at DESC`, careerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if ideas == nil {
		ideas = []models.ProjectIdea{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"project_ideas": ideas})
}
