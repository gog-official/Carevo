package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/guruorgoru/carevo/internal/cache"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/models"
	"github.com/jmoiron/sqlx"
)

type CareerHandler struct {
	db    *sqlx.DB
	cache *cache.Store
}

func NewCareerHandler(db *sqlx.DB, cacheStore *cache.Store) *CareerHandler {
	return &CareerHandler{db: db, cache: cacheStore}
}

func (h *CareerHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")

	cacheKey := fmt.Sprintf("careers:list:page=%d:limit=%d:category=%s:tag=%s", page, limit, category, tag)
	if h.cache != nil {
		var cachedResp map[string]any
		if hit, err := h.cache.Get(r.Context(), cacheKey, &cachedResp); err == nil && hit {
			writeJSON(w, http.StatusOK, cachedResp)
			return
		}
	}

	where := []string{"1=1"}
	args := []any{}
	argN := 1

	if category != "" {
		where = append(where, "c.category_id = (SELECT id FROM categories WHERE slug = $"+strconv.Itoa(argN)+")")
		args = append(args, category)
		argN++
	}

	if tag != "" {
		where = append(where, "c.id IN (SELECT career_id FROM career_tags WHERE tag = $"+strconv.Itoa(argN)+")")
		args = append(args, tag)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := `SELECT COUNT(*) FROM careers c WHERE ` + whereClause
	if err := h.db.Get(&total, countQuery, args...); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	var careers []models.Career
	query := fmt.Sprintf(`SELECT c.*, cat.name as category_name, cat.slug as category_slug
		FROM careers c
		JOIN categories cat ON cat.id = c.category_id
		WHERE %s
		ORDER BY c.id
		LIMIT $%d OFFSET $%d`, whereClause, argN, argN+1)
	args = append(args, limit, offset)

	if err := h.db.Select(&careers, query, args...); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	type careerResponse struct {
		models.Career
		SalaryRange models.SalaryRange `json:"salary_range"`
		Tags        []string           `json:"tags,omitempty"`
	}

	resp := make([]careerResponse, len(careers))
	for i, c := range careers {
		tags := []string{}
		h.db.Select(&tags, `SELECT tag FROM career_tags WHERE career_id = $1 ORDER BY tag`, c.ID)
		resp[i] = careerResponse{Career: c, SalaryRange: c.SalaryRange(), Tags: tags}
	}

	respData := map[string]any{
		"careers":     resp,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": (total + limit - 1) / limit,
	}

	if h.cache != nil {
		h.cache.Set(r.Context(), cacheKey, respData, 5*time.Minute)
	}

	writeJSON(w, http.StatusOK, respData)
}

func (h *CareerHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	cacheKey := fmt.Sprintf("careers:slug:%s", slug)
	if h.cache != nil {
		var cachedResp map[string]any
		if hit, err := h.cache.Get(r.Context(), cacheKey, &cachedResp); err == nil && hit {
			writeJSON(w, http.StatusOK, cachedResp)
			return
		}
	}

	var career models.Career
	query := `SELECT c.*, cat.name as category_name, cat.slug as category_slug
		FROM careers c
		JOIN categories cat ON cat.id = c.category_id
		WHERE c.slug = $1`
	if err := h.db.Get(&career, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "career not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	tags := []string{}
	h.db.Select(&tags, `SELECT tag FROM career_tags WHERE career_id = $1 ORDER BY tag`, career.ID)

	sources := []models.CareerSource{}
	h.db.Select(&sources, `SELECT * FROM career_sources WHERE career_id = $1 ORDER BY source_type`, career.ID)
	if sources == nil {
		sources = []models.CareerSource{}
	}

	type careerDetail struct {
		models.Career
		SalaryRange models.SalaryRange   `json:"salary_range"`
		Tags        []string             `json:"tags"`
		Sources     []models.CareerSource `json:"sources,omitempty"`
	}

	respData := map[string]any{
		"career": careerDetail{Career: career, SalaryRange: career.SalaryRange(), Tags: tags, Sources: sources},
	}

	if h.cache != nil {
		h.cache.Set(r.Context(), cacheKey, respData, 5*time.Minute)
	}

	writeJSON(w, http.StatusOK, respData)
}

func (h *CareerHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'q' is required"})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	err := h.db.Get(&total, `
		SELECT COUNT(*) FROM careers c
		WHERE to_tsvector('english', coalesce(c.title,'') || ' ' || coalesce(c.description,'') || ' ' || coalesce(c.summary,''))
		@@ plainto_tsquery('english', $1)`, q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search error"})
		return
	}

	var careers []models.Career
	err = h.db.Select(&careers, `
		SELECT c.*, cat.name as category_name, cat.slug as category_slug,
			ts_rank(to_tsvector('english', coalesce(c.title,'') || ' ' || coalesce(c.description,'') || ' ' || coalesce(c.summary,'')),
				plainto_tsquery('english', $1)) as rank
		FROM careers c
		JOIN categories cat ON cat.id = c.category_id
		WHERE to_tsvector('english', coalesce(c.title,'') || ' ' || coalesce(c.description,'') || ' ' || coalesce(c.summary,''))
		@@ plainto_tsquery('english', $1)
		ORDER BY rank DESC
		LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search error"})
		return
	}

	type searchResult struct {
		models.Career
		SalaryRange models.SalaryRange `json:"salary_range"`
	}

	resp := make([]searchResult, len(careers))
	for i, c := range careers {
		resp[i] = searchResult{Career: c, SalaryRange: c.SalaryRange()}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"careers":     resp,
		"query":       q,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *CareerHandler) AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	mode := r.URL.Query().Get("mode")

	where := []string{"1=1"}
	args := []any{}
	argN := 1

	addParam := func(field, operator, rawValue string, parser func(string) (any, error)) {
		if rawValue == "" {
			return
		}
		val, err := parser(rawValue)
		if err != nil {
			return
		}
		where = append(where, fmt.Sprintf("c.%s %s $%d", field, operator, argN))
		args = append(args, val)
		argN++
	}

	parseInt := func(s string) (any, error) { return strconv.Atoi(s) }
	parseBool := func(s string) (any, error) { return strconv.ParseBool(s) }

	if cat := r.URL.Query().Get("category"); cat != "" {
		where = append(where, fmt.Sprintf("c.category_id = (SELECT id FROM categories WHERE slug = $%d)", argN))
		args = append(args, cat)
		argN++
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		where = append(where, fmt.Sprintf("c.id IN (SELECT career_id FROM career_tags WHERE tag = $%d)", argN))
		args = append(args, tag)
		argN++
	}
	if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
		where = append(where, fmt.Sprintf("(to_tsvector('english', coalesce(c.title,'') || ' ' || coalesce(c.description,'') || ' ' || coalesce(c.summary,'')) @@ plainto_tsquery('english', $%d))", argN))
		args = append(args, q)
		argN++
	}

	addParam("salary_max", ">=", r.URL.Query().Get("min_salary"), parseInt)
	addParam("salary_min", "<=", r.URL.Query().Get("max_salary"), parseInt)
	addParam("difficulty", "=", r.URL.Query().Get("difficulty"), parseInt)
	addParam("future_proof_score", ">=", r.URL.Query().Get("min_future_proof"), parseInt)
	addParam("work_life_balance", ">=", r.URL.Query().Get("min_work_life_balance"), parseInt)
	addParam("creative_score", ">=", r.URL.Query().Get("min_creative"), parseInt)
	addParam("technical_score", ">=", r.URL.Query().Get("min_technical"), parseInt)
	addParam("freelance_potential", ">=", r.URL.Query().Get("min_freelance"), parseInt)
	addParam("is_government", "=", r.URL.Query().Get("is_government"), parseBool)
	addParam("is_remote_ok", "=", r.URL.Query().Get("is_remote_ok"), parseBool)

	if sd := r.URL.Query().Get("study_duration"); sd != "" {
		where = append(where, fmt.Sprintf("c.study_duration = $%d", argN))
		args = append(args, sd)
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := h.db.Get(&total, `SELECT COUNT(*) FROM careers c WHERE `+whereClause, args...); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	query := fmt.Sprintf(`SELECT c.*, cat.name as category_name, cat.slug as category_slug
		FROM careers c JOIN categories cat ON cat.id = c.category_id
		WHERE %s ORDER BY c.id LIMIT $%d OFFSET $%d`, whereClause, argN, argN+1)
	fullArgs := append(args, limit, offset)

	var careers []models.Career
	if err := h.db.Select(&careers, query, fullArgs...); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	type careerResponse struct {
		models.Career
		SalaryRange models.SalaryRange `json:"salary_range"`
		Tags        []string           `json:"tags,omitempty"`
	}

	resp := make([]careerResponse, len(careers))
	for i, c := range careers {
		tags := []string{}
		h.db.Select(&tags, `SELECT tag FROM career_tags WHERE career_id = $1 ORDER BY tag`, c.ID)
		resp[i] = careerResponse{Career: c, SalaryRange: c.SalaryRange(), Tags: tags}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"careers":     resp,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": (total + limit - 1) / limit,
		"mode":        mode,
	})
}

func (h *CareerHandler) Categories(w http.ResponseWriter, r *http.Request) {
	var categories []models.CategoryWithCount
	err := h.db.Select(&categories, `
		SELECT cat.*, COUNT(c.id) as career_count
		FROM categories cat
		LEFT JOIN careers c ON c.category_id = cat.id
		GROUP BY cat.id
		ORDER BY cat.name`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if categories == nil {
		categories = []models.CategoryWithCount{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"categories": categories})
}

func (h *CareerHandler) Resources(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid career id"})
		return
	}

	var resources []models.Resource
	err = h.db.Select(&resources, `SELECT * FROM resources WHERE career_id = $1 ORDER BY sort_order`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if resources == nil {
		resources = []models.Resource{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"resources": resources})
}

func (h *CareerHandler) Roadmap(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid career id"})
		return
	}

	var steps []models.RoadmapStep
	err = h.db.Select(&steps, `SELECT * FROM roadmap_steps WHERE career_id = $1 ORDER BY step_number`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if steps == nil {
		steps = []models.RoadmapStep{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"roadmap": steps})
}

func (h *CareerHandler) Compare(w http.ResponseWriter, r *http.Request) {
	var req models.CompareCareerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.CareerIDs) < 2 || len(req.CareerIDs) > 3 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "compare 2-3 careers at a time"})
		return
	}

	result := make([]models.CareerCompare, 0, len(req.CareerIDs))
	for _, id := range req.CareerIDs {
		var career models.Career
		err := h.db.Get(&career, `SELECT c.*, cat.name as category_name, cat.slug as category_slug
			FROM careers c JOIN categories cat ON cat.id = c.category_id WHERE c.id = $1`, id)
		if err != nil {
			continue
		}

		tags := []string{}
		h.db.Select(&tags, `SELECT tag FROM career_tags WHERE career_id = $1 ORDER BY tag`, career.ID)

		result = append(result, models.CareerCompare{
			Career:      career,
			SalaryRange: career.SalaryRange(),
			Tags:        tags,
		})
	}

	if len(result) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "not enough valid careers found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"comparison": result})
}

func (h *CareerHandler) SaveComparison(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req struct {
		CareerIDs []int64 `json:"career_ids"`
		Name      string  `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.CareerIDs) < 2 || len(req.CareerIDs) > 3 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "save 2-3 careers per comparison"})
		return
	}

	idsJSON, _ := json.Marshal(req.CareerIDs)
	var comp models.SavedComparison
	err := h.db.QueryRowx(
		`INSERT INTO career_comparisons (user_id, career_ids, name) VALUES ($1, $2, $3) RETURNING id, user_id, career_ids, name, created_at`,
		claims.UserID, idsJSON, req.Name,
	).StructScan(&comp)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save comparison"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"comparison": comp})
}

func (h *CareerHandler) ListComparisons(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var comps []models.SavedComparison
	err := h.db.Select(&comps, `SELECT * FROM career_comparisons WHERE user_id = $1 ORDER BY created_at DESC`, claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if comps == nil {
		comps = []models.SavedComparison{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"comparisons": comps})
}

func (h *CareerHandler) GetComparison(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	compID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid comparison id"})
		return
	}

	var comp models.SavedComparison
	err = h.db.Get(&comp, `SELECT * FROM career_comparisons WHERE id = $1 AND user_id = $2`, compID, claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "comparison not found"})
		return
	}

	result := make([]models.CareerCompare, 0, len(comp.CareerIDs))
	for _, id := range comp.CareerIDs {
		var career models.Career
		err := h.db.Get(&career, `SELECT c.*, cat.name as category_name, cat.slug as category_slug
			FROM careers c JOIN categories cat ON cat.id = c.category_id WHERE c.id = $1`, id)
		if err != nil {
			continue
		}
		tags := []string{}
		h.db.Select(&tags, `SELECT tag FROM career_tags WHERE career_id = $1 ORDER BY tag`, career.ID)
		result = append(result, models.CareerCompare{
			Career:      career,
			SalaryRange: career.SalaryRange(),
			Tags:        tags,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"comparison": comp, "careers": result})
}

func (h *CareerHandler) DeleteComparison(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromCtx(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	compID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid comparison id"})
		return
	}

	_, err = h.db.Exec(`DELETE FROM career_comparisons WHERE id = $1 AND user_id = $2`, compID, claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete comparison"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *CareerHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	careerID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid career id"})
		return
	}

	var sources []models.CareerSource
	err = h.db.Select(&sources, `SELECT * FROM career_sources WHERE career_id = $1 ORDER BY source_type`, careerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if sources == nil {
		sources = []models.CareerSource{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"sources": sources})
}
