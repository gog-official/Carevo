package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/guruorgoru/carevo/internal/ai"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type AIHandler struct {
	db       *sqlx.DB
	worker   *ai.Worker
	provider *ai.GeminiProvider
}

func NewAIHandler(db *sqlx.DB, worker *ai.Worker, provider *ai.GeminiProvider) *AIHandler {
	return &AIHandler{db: db, worker: worker, provider: provider}
}

// @Summary      Get survey questions
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Router       /ai/survey/questions [get]
func (h *AIHandler) SurveyQuestions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var questions []models.SurveyQuestion
	err := h.db.Select(&questions, `SELECT id, category, question_text, options, sort_order FROM survey_questions ORDER BY sort_order`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if questions == nil {
		questions = []models.SurveyQuestion{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"questions": questions})
}

// @Summary      Submit survey answers, trigger AI analysis
// @Tags         ai
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.SubmitSurveyRequest  true  "Survey answers"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Router       /ai/survey/submit [post]
func (h *AIHandler) SubmitSurvey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req models.SubmitSurveyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.Answers) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one answer is required"})
		return
	}

	var questionCount int
	h.db.Get(&questionCount, `SELECT COUNT(*) FROM survey_questions`)

	tx, err := h.db.Beginx()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	defer tx.Rollback()

	for _, a := range req.Answers {
		if a.QuestionID == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "question_id is required for each answer"})
			return
		}
		ansJSON, _ := json.Marshal(a.Answer)
		_, err := tx.Exec(
			`INSERT INTO survey_responses (user_id, question_id, answer) VALUES ($1, $2, $3)
			 ON CONFLICT (user_id, question_id) DO UPDATE SET answer = $3`,
			claims.UserID, a.QuestionID, ansJSON,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save answer"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save answers"})
		return
	}

	h.worker.Submit(claims.UserID)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "processing",
		"message": "Survey submitted. AI analysis started. Check /ai/results/me for results.",
	})
}

// @Summary      Get latest AI career matches
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]string
// @Router       /ai/results/me [get]
func (h *AIHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var result models.AIResult
	err := h.db.Get(&result, `SELECT id, user_id, recommended_careers, status, error_message, created_at, completed_at FROM ai_results WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no results yet. submit the survey first."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if result.Status == "processing" {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "processing",
			"message": "AI is analyzing your answers. Check back soon.",
		})
		return
	}

	if result.Status == "error" {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":        "error",
			"error_message": result.ErrorMessage,
		})
		return
	}

	var careers []models.ScoredCareer
	if result.RecommendedCareers != nil {
		json.Unmarshal(result.RecommendedCareers, &careers)
	}
	if careers == nil {
		careers = []models.ScoredCareer{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":             "completed",
		"recommended_careers": careers,
		"completed_at":       result.CompletedAt,
	})
}

// @Summary      Get personalized roadmap from top career match
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]string
// @Router       /ai/results/me/roadmap [get]
func (h *AIHandler) GetRoadmap(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var result models.AIResult
	err := h.db.Get(&result, `SELECT id, user_id, recommended_careers, status FROM ai_results WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no results yet. submit the survey first."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if result.Status != "completed" {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  result.Status,
			"message": "Analysis not yet complete.",
		})
		return
	}

	var careers []models.ScoredCareer
	if result.RecommendedCareers != nil {
		json.Unmarshal(result.RecommendedCareers, &careers)
	}

	if len(careers) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no career matches found"})
		return
	}

	topCareer := careers[0]

	type stepWithCareer struct {
		models.RoadmapStep
		CareerTitle string `json:"career_title"`
		CareerSlug  string `json:"career_slug"`
	}

	var career struct {
		Title string `db:"title"`
		Slug  string `db:"slug"`
	}
	h.db.Get(&career, `SELECT title, slug FROM careers WHERE id = $1`, topCareer.CareerID)

	var steps []models.RoadmapStep
	h.db.Select(&steps, `SELECT * FROM roadmap_steps WHERE career_id = $1 ORDER BY step_number`, topCareer.CareerID)
	if steps == nil {
		steps = []models.RoadmapStep{}
	}

	resp := make([]stepWithCareer, len(steps))
	for i, s := range steps {
		resp[i] = stepWithCareer{RoadmapStep: s, CareerTitle: career.Title, CareerSlug: career.Slug}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"career_title":   career.Title,
		"career_slug":    career.Slug,
		"compatibility":  topCareer.Score,
		"roadmap":        resp,
	})
}

// @Summary      Get compatibility score for a specific career
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Career ID"
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]string
// @Router       /ai/careers/{id}/score [get]
func (h *AIHandler) GetCareerScore(w http.ResponseWriter, r *http.Request) {
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

	var score models.CareerScore
	err = h.db.Get(&score, `SELECT * FROM career_scores WHERE user_id = $1 AND career_id = $2`, claims.UserID, careerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "score not found. complete the survey first."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"score": score})
}

// @Summary      Chat with AI mentor (SSE stream)
// @Tags         ai
// @Accept       json
// @Produce      text/event-stream
// @Security     BearerAuth
// @Param        body  body  models.AIMentorRequest  true  "Career ID and message"
// @Success      200  {object}  models.AIMentorChunk
// @Failure      400  {object}  map[string]string
// @Router       /ai/chat [post]
func (h *AIHandler) ChatStream(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req models.AIMentorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}

	var career struct {
		Title       string `db:"title"`
		Description string `db:"description"`
	}
	err := h.db.Get(&career, `SELECT title, description FROM careers WHERE id = $1`, req.CareerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "career not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	systemPrompt := fmt.Sprintf(`You are a friendly career mentor helping someone explore the "%s" career in Nepal.

Career context:
- Title: %s
- Description: %s

Guidelines:
- Be encouraging and practical
- Give Nepal-specific advice (local exams, institutes, job market)
- Keep answers concise (2-3 paragraphs)
- Don't make up specific salary figures
- If asked something outside career guidance, gently redirect

Respond in a conversational, mentor-like tone.`, career.Title, career.Title, career.Description)

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	err = h.provider.GenerateStream(r.Context(), systemPrompt, req.Message, func(token string) {
		cleaned := strings.ReplaceAll(token, "\n", "\\n")
		chunk := models.AIMentorChunk{Token: cleaned}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	})
	if err != nil {
		fmt.Fprintf(w, "data: {\"token\": \"[error: %s]\"}\n\n", err.Error())
		flusher.Flush()
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}
