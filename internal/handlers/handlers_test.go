package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/guruorgoru/carevo/internal/auth"
	"github.com/guruorgoru/carevo/internal/cache"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/guruorgoru/carevo/internal/router"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testSuite struct {
	db       *sqlx.DB
	jwt      *auth.JWTService
	pg       *postgres.PostgresContainer
	router   *chi.Mux
}

var ts *testSuite

func runMigrations(db *sqlx.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type migration struct {
		number int
		path   string
	}

	var upMigrations []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		parts := strings.SplitN(e.Name(), "_", 2)
		n, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		upMigrations = append(upMigrations, migration{number: n, path: filepath.Join(migrationsDir, e.Name())})
	}

	sort.Slice(upMigrations, func(i, j int) bool {
		return upMigrations[i].number < upMigrations[j].number
	})

	for _, m := range upMigrations {
		sqlBytes, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", m.path, err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec migration %s: %w", m.path, err)
		}
	}
	return nil
}

func setupTestSuite() (*testSuite, error) {
	ctx := context.Background()

	pg, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:17-alpine"),
		postgres.WithDatabase("carevo_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres: %w", err)
	}

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pg.Terminate(ctx)
		return nil, fmt.Errorf("connection string: %w", err)
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		pg.Terminate(ctx)
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := runMigrations(db, "../../migrations"); err != nil {
		db.Close()
		pg.Terminate(ctx)
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	jwtService := auth.NewJWTService("test-secret-for-tests-only", 15*time.Minute, 720*time.Hour)

	// Use the real router factory with a dummy cache store.
	// The cache store uses a dummy Redis addr; operations will fail gracefully
	// and handlers will fall back to the database.
	cacheStore := cache.New("localhost:6379")
	r := router.New(db, jwtService, cacheStore, nil, nil)

	return &testSuite{
		db:     db,
		jwt:    jwtService,
		pg:     pg,
		router: r,
	}, nil
}

func (s *testSuite) cleanupData() {
	s.db.MustExec("TRUNCATE categories, careers, career_tags, resources, roadmap_steps, bookmarks, challenges, challenge_checkins, project_ideas, survey_questions, survey_responses, ai_results, career_scores CASCADE")
}

func (s *testSuite) teardown() {
	s.db.Close()
	s.pg.Terminate(context.Background())
}

func TestMain(m *testing.M) {
	var err error
	ts, err = setupTestSuite()
	if err != nil {
		log.Fatalf("setup test suite: %v", err)
	}
	code := m.Run()
	ts.teardown()
	os.Exit(code)
}

// --- helpers ---

func (s *testSuite) newRequest(method, target string, body any) *http.Request {
	var buf io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, target, buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (s *testSuite) authenticatedRequest(method, target string, body any, token string) *http.Request {
	req := s.newRequest(method, target, body)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func (s *testSuite) getAuthToken(t *testing.T) string {
	t.Helper()

	payload := map[string]string{
		"email":    fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		"password": "password123",
		"name":     "Test User",
	}
	req := s.newRequest("POST", "/auth/register", payload)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Tokens struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"tokens"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Tokens.AccessToken)
	return resp.Tokens.AccessToken
}

// registerUser is like getAuthToken but returns the full response including user and tokens
func (s *testSuite) registerUser(t *testing.T, email, password, name string) (accessToken, refreshToken string) {
	t.Helper()
	payload := map[string]string{
		"email":    email,
		"password": password,
		"name":     name,
	}
	req := s.newRequest("POST", "/auth/register", payload)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	var resp struct {
		Tokens struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"tokens"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp.Tokens.AccessToken, resp.Tokens.RefreshToken
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, dest any) {
	t.Helper()
	err := json.Unmarshal(rec.Body.Bytes(), dest)
	require.NoError(t, err)
}

// --- Health ---

func TestHealthCheck(t *testing.T) {
	req := ts.newRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, true, resp["database"])
}

// --- Ping ---

func TestPing(t *testing.T) {
	req := ts.newRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- Auth ---

func TestAuthRegister_Success(t *testing.T) {
	email := fmt.Sprintf("register-success-%d@example.com", time.Now().UnixNano())
	req := ts.newRequest("POST", "/auth/register", map[string]string{
		"email":    email,
		"password": "securePass1",
		"name":     "Alice",
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "user")
	assert.Contains(t, resp, "tokens")

	user := resp["user"].(map[string]any)
	assert.Equal(t, "Alice", user["name"])
	assert.Equal(t, email, user["email"])
	assert.NotEmpty(t, user["id"])

	tokens := resp["tokens"].(map[string]any)
	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
}

func TestAuthRegister_DuplicateEmail(t *testing.T) {
	email := fmt.Sprintf("duplicate-%d@example.com", time.Now().UnixNano())
	payload := map[string]string{"email": email, "password": "pass123", "name": "Bob"}

	req1 := ts.newRequest("POST", "/auth/register", payload)
	rec1 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusCreated, rec1.Code)

	req2 := ts.newRequest("POST", "/auth/register", payload)
	rec2 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusConflict, rec2.Code)

	var errResp map[string]string
	decodeBody(t, rec2, &errResp)
	assert.Contains(t, errResp["error"], "email already in use")
}

func TestAuthRegister_EmptyEmail(t *testing.T) {
	req := ts.newRequest("POST", "/auth/register", map[string]string{
		"email":    "",
		"password": "pass123",
		"name":     "NoEmail",
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthRegister_EmptyPassword(t *testing.T) {
	req := ts.newRequest("POST", "/auth/register", map[string]string{
		"email":    fmt.Sprintf("emptypw-%d@example.com", time.Now().UnixNano()),
		"password": "",
		"name":     "NoPW",
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthLogin_Success(t *testing.T) {
	email := fmt.Sprintf("login-success-%d@example.com", time.Now().UnixNano())
	_, _ = ts.registerUser(t, email, "mypassword", "LoginTest")

	req := ts.newRequest("POST", "/auth/login", map[string]string{
		"email":    email,
		"password": "mypassword",
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "user")
	assert.Contains(t, resp, "tokens")

	tokens := resp["tokens"].(map[string]any)
	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	email := fmt.Sprintf("login-wrongpw-%d@example.com", time.Now().UnixNano())
	ts.registerUser(t, email, "correctpw", "WrongPW")

	req := ts.newRequest("POST", "/auth/login", map[string]string{
		"email":    email,
		"password": "wrongpassword",
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthLogin_NonexistentEmail(t *testing.T) {
	req := ts.newRequest("POST", "/auth/login", map[string]string{
		"email":    "doesnotexist@example.com",
		"password": "somepassword",
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthRefresh_Success(t *testing.T) {
	email := fmt.Sprintf("refresh-success-%d@example.com", time.Now().UnixNano())
	_, refreshToken := ts.registerUser(t, email, "password", "RefreshTest")
	require.NotEmpty(t, refreshToken)

	req := ts.newRequest("POST", "/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "tokens")
	tokens := resp["tokens"].(map[string]any)
	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
}

func TestAuthRefresh_InvalidToken(t *testing.T) {
	req := ts.newRequest("POST", "/auth/refresh", map[string]string{
		"refresh_token": "invalid-refresh-token-value",
	})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthRefresh_MissingToken(t *testing.T) {
	req := ts.newRequest("POST", "/auth/refresh", map[string]string{})
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Users ---

func TestUserGetMe(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("GET", "/users/me", nil, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "user")
	user := resp["user"].(map[string]any)
	assert.NotEmpty(t, user["id"])
	assert.NotEmpty(t, user["email"])
}

func TestUserGetMe_Unauthenticated(t *testing.T) {
	req := ts.newRequest("GET", "/users/me", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUserUpdateMe(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("PUT", "/users/me", map[string]string{
		"name":       "Updated Name",
		"bio":        "This is my bio",
		"avatar_url": "https://example.com/avatar.png",
	}, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "user")
	user := resp["user"].(map[string]any)
	assert.Equal(t, "Updated Name", user["name"])
	assert.Equal(t, "This is my bio", user["bio"])
	assert.Equal(t, "https://example.com/avatar.png", user["avatar_url"])
}

// --- Careers ---

func TestCareersList_Empty(t *testing.T) {
	req := ts.newRequest("GET", "/careers", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "careers")
	assert.Contains(t, resp, "page")
	assert.Contains(t, resp, "limit")
	assert.Contains(t, resp, "total")
	assert.Contains(t, resp, "total_pages")

	careers := resp["careers"].([]any)
	assert.Empty(t, careers)
	assert.Equal(t, float64(1), resp["page"])
	assert.Equal(t, float64(20), resp["limit"])
	assert.Equal(t, float64(0), resp["total"])
}

func TestCareersList_Pagination(t *testing.T) {
	req := ts.newRequest("GET", "/careers?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Equal(t, float64(1), resp["page"])
	assert.Equal(t, float64(10), resp["limit"])
}

func TestCareersGetBySlug_NotFound(t *testing.T) {
	req := ts.newRequest("GET", "/careers/nonexistent-slug", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCareersSearch_NoResults(t *testing.T) {
	req := ts.newRequest("GET", "/careers/search?q=zzzznotfound", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "careers")
	assert.Contains(t, resp, "query")
	careers := resp["careers"].([]any)
	assert.Empty(t, careers)
}

func TestCareersSearch_MissingQuery(t *testing.T) {
	req := ts.newRequest("GET", "/careers/search", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Categories ---

func TestCategories_Empty(t *testing.T) {
	req := ts.newRequest("GET", "/categories", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "categories")
	categories := resp["categories"].([]any)
	assert.Empty(t, categories)
}

// --- Resources & Roadmap ---

func TestResources_CareerNotFound(t *testing.T) {
	req := ts.newRequest("GET", "/careers/999/resources", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "resources")
	resources := resp["resources"].([]any)
	assert.Empty(t, resources)
}

func TestRoadmap_CareerNotFound(t *testing.T) {
	req := ts.newRequest("GET", "/careers/999/roadmap", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "roadmap")
	roadmap := resp["roadmap"].([]any)
	assert.Empty(t, roadmap)
}

// --- Bookmarks ---

func TestBookmarks_CareerNotFound(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("POST", "/users/me/bookmarks", map[string]int64{
		"career_id": 99999,
	}, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBookmarks_ListEmpty(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("GET", "/users/me/bookmarks", nil, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	decodeBody(t, rec, &resp)
	assert.Contains(t, resp, "bookmarks")
	bookmarks := resp["bookmarks"].([]any)
	assert.Empty(t, bookmarks)
}

// --- Challenges ---

func TestChallenges_StartInvalidCareer(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("POST", "/challenges", map[string]int64{
		"career_id": 99999,
	}, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestChallenges_ActiveNotFound(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("GET", "/challenges/active", nil, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Projects ---

func TestProjectIdeas_CreateAndList(t *testing.T) {
	ts.cleanupData()
	token := ts.getAuthToken(t)

	// create a career and category directly in DB
	var catID int64
	err := ts.db.QueryRow(
		`INSERT INTO categories (name, slug, description, icon) VALUES ($1, $2, $3, $4) RETURNING id`,
		"Test Category", "test-category", "desc", "icon",
	).Scan(&catID)
	require.NoError(t, err)

	var careerID int64
	err = ts.db.QueryRow(
		`INSERT INTO careers (title, slug, summary, description, category_id, difficulty, future_proof_score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"Test Career", "test-career", "summary", "description", catID, 2, 50,
	).Scan(&careerID)
	require.NoError(t, err)
	require.Greater(t, careerID, int64(0))

	// create a project idea
	req := ts.authenticatedRequest("POST", fmt.Sprintf("/careers/%d/projects", careerID), map[string]string{
		"title":       "My Project",
		"description": "A great project idea",
	}, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var createResp map[string]any
	decodeBody(t, rec, &createResp)
	assert.Contains(t, createResp, "project_idea")
	idea := createResp["project_idea"].(map[string]any)
	assert.Equal(t, "My Project", idea["title"])
	assert.Equal(t, "A great project idea", idea["description"])

	// list project ideas for that career
	req2 := ts.authenticatedRequest("GET", fmt.Sprintf("/careers/%d/projects", careerID), nil, token)
	rec2 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
	var listResp map[string]any
	decodeBody(t, rec2, &listResp)
	assert.Contains(t, listResp, "project_ideas")
	ideas := listResp["project_ideas"].([]any)
	assert.Len(t, ideas, 1)
}

func TestProjectIdeas_CareerNotFound(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("POST", "/careers/99999/projects", map[string]string{
		"title":       "Test",
		"description": "Test desc",
	}, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Middleware level tests ---

func TestUnauthenticatedRoutes(t *testing.T) {
	// routes that should 401 without a token
	protectedRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/users/me"},
		{"PUT", "/users/me"},
		{"POST", "/challenges"},
		{"GET", "/challenges/active"},
	}

	for _, rt := range protectedRoutes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := ts.newRequest(rt.method, rt.path, nil)
			rec := httptest.NewRecorder()
			ts.router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

// Test the AuthMiddleware in isolation for edge cases
func TestAuthMiddleware_InvalidBearer(t *testing.T) {
	req := ts.newRequest("GET", "/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	req := ts.newRequest("GET", "/users/me", nil)
	req.Header.Set("Authorization", "NotBearer token123")
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// Test that the middleware correctly sets user in context
func TestAuthMiddleware_ClaimsInContext(t *testing.T) {
	email := fmt.Sprintf("claims-test-%d@example.com", time.Now().UnixNano())
	token, _ := ts.registerUser(t, email, "password", "ClaimsTest")

	// hand-build a handler that reads from context
	handler := middleware.AuthMiddleware(ts.jwt)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.UserFromCtx(r.Context())
		assert.NotNil(t, claims)
		assert.Equal(t, email, claims.Email)
		assert.Greater(t, claims.UserID, int64(0))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- Integration: full bookmark + challenge flow with a real career ---

func TestBookmarkAndChallengeFlow(t *testing.T) {
	ts.cleanupData()
	token := ts.getAuthToken(t)

	// create a career
	var catID int64
	err := ts.db.QueryRow(
		`INSERT INTO categories (name, slug, description, icon) VALUES ($1, $2, $3, $4) RETURNING id`,
		"Flow Category", "flow-category", "desc", "icon",
	).Scan(&catID)
	require.NoError(t, err)

	var careerID int64
	err = ts.db.QueryRow(
		`INSERT INTO careers (title, slug, summary, description, category_id, difficulty, future_proof_score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"Flow Career", "flow-career", "summary", "description", catID, 3, 60,
	).Scan(&careerID)
	require.NoError(t, err)

	// bookmark
	req := ts.authenticatedRequest("POST", "/users/me/bookmarks", map[string]int64{"career_id": careerID}, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// list bookmarks - should have 1
	req2 := ts.authenticatedRequest("GET", "/users/me/bookmarks", nil, token)
	rec2 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
	var bmResp map[string]any
	decodeBody(t, rec2, &bmResp)
	bookmarks := bmResp["bookmarks"].([]any)
	assert.Len(t, bookmarks, 1)

	// start challenge
	req3 := ts.authenticatedRequest("POST", "/challenges", map[string]int64{"career_id": careerID}, token)
	rec3 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusCreated, rec3.Code)

	// get active challenge
	req4 := ts.authenticatedRequest("GET", "/challenges/active", nil, token)
	rec4 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec4, req4)
	assert.Equal(t, http.StatusOK, rec4.Code)
	var chResp map[string]any
	decodeBody(t, rec4, &chResp)
	assert.Contains(t, chResp, "challenge")
	challenge := chResp["challenge"].(map[string]any)
	assert.True(t, challenge["is_active"].(bool))
}

// --- Career with data test ---

func TestCareerWithData(t *testing.T) {
	ts.cleanupData()
	// create a category and career, then test get-by-slug and list

	var catID int64
	err := ts.db.QueryRow(
		`INSERT INTO categories (name, slug, description, icon) VALUES ($1, $2, $3, $4) RETURNING id`,
		"Data Category", "data-category", "desc", "icon",
	).Scan(&catID)
	require.NoError(t, err)

	_, err = ts.db.Exec(
		`INSERT INTO careers (title, slug, summary, description, category_id, difficulty, future_proof_score, skills, salary_min, salary_max)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		"Data Career", "data-career", "A data career summary", "Full description", catID, 2, 70,
		pq.StringArray{"skill1", "skill2"}, 50000, 120000,
	)
	require.NoError(t, err)

	// list should have 1
	req := ts.newRequest("GET", "/careers", nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var listResp map[string]any
	decodeBody(t, rec, &listResp)
	careers := listResp["careers"].([]any)
	assert.Len(t, careers, 1)

	// get by slug
	req2 := ts.newRequest("GET", "/careers/data-career", nil)
	rec2 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
	var slugResp map[string]any
	decodeBody(t, rec2, &slugResp)
	assert.Contains(t, slugResp, "career")
	career := slugResp["career"].(map[string]any)
	assert.Equal(t, "data-career", career["slug"])
}

// --- Resources and Roadmap with data ---

func TestResourcesAndRoadmapWithData(t *testing.T) {
	ts.cleanupData()
	var catID int64
	err := ts.db.QueryRow(
		`INSERT INTO categories (name, slug, description, icon) VALUES ($1, $2, $3, $4) RETURNING id`,
		"RR Category", "rr-category", "desc", "icon",
	).Scan(&catID)
	require.NoError(t, err)

	var careerID int64
	err = ts.db.QueryRow(
		`INSERT INTO careers (title, slug, summary, description, category_id, difficulty, future_proof_score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"RR Career", "rr-career", "summary", "description", catID, 1, 30,
	).Scan(&careerID)
	require.NoError(t, err)

	// add resources
	_, err = ts.db.Exec(
		`INSERT INTO resources (career_id, title, url, description, is_free, sort_order) VALUES ($1, $2, $3, $4, $5, $6)`,
		careerID, "Free Course", "https://example.com", "A free course", true, 1,
	)
	require.NoError(t, err)

	// add roadmap steps
	_, err = ts.db.Exec(
		`INSERT INTO roadmap_steps (career_id, step_number, title, description, duration, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		careerID, 1, "Step One", "First step", "1 month", 1,
	)
	require.NoError(t, err)

	// get resources
	req := ts.newRequest("GET", fmt.Sprintf("/careers/%d/resources", careerID), nil)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resResp map[string]any
	decodeBody(t, rec, &resResp)
	resources := resResp["resources"].([]any)
	assert.Len(t, resources, 1)

	// get roadmap
	req2 := ts.newRequest("GET", fmt.Sprintf("/careers/%d/roadmap", careerID), nil)
	rec2 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
	var rmResp map[string]any
	decodeBody(t, rec2, &rmResp)
	roadmap := rmResp["roadmap"].([]any)
	assert.Len(t, roadmap, 1)
}

// --- Edge cases ---

func TestAuthLogin_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte(`not json`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserUpdateMe_EmptyBody(t *testing.T) {
	token := ts.getAuthToken(t)

	req := ts.authenticatedRequest("PUT", "/users/me", map[string]string{}, token)
	rec := httptest.NewRecorder()
	ts.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestChallenges_DuplicateActive(t *testing.T) {
	ts.cleanupData()
	token := ts.getAuthToken(t)

	var catID int64
	err := ts.db.QueryRow(
		`INSERT INTO categories (name, slug, description, icon) VALUES ($1, $2, $3, $4) RETURNING id`,
		"Dup Category", "dup-category", "desc", "icon",
	).Scan(&catID)
	require.NoError(t, err)

	var careerID int64
	err = ts.db.QueryRow(
		`INSERT INTO careers (title, slug, summary, description, category_id, difficulty, future_proof_score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"Dup Career", "dup-career", "summary", "description", catID, 1, 30,
	).Scan(&careerID)
	require.NoError(t, err)

	// start first challenge
	req1 := ts.authenticatedRequest("POST", "/challenges", map[string]int64{"career_id": careerID}, token)
	rec1 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusCreated, rec1.Code)

	// try to start a second one (different career)
	var careerID2 int64
	err = ts.db.QueryRow(
		`INSERT INTO careers (title, slug, summary, description, category_id, difficulty, future_proof_score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		"Dup Career 2", "dup-career-2", "summary", "description", catID, 1, 30,
	).Scan(&careerID2)
	require.NoError(t, err)

	req2 := ts.authenticatedRequest("POST", "/challenges", map[string]int64{"career_id": careerID2}, token)
	rec2 := httptest.NewRecorder()
	ts.router.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusConflict, rec2.Code)
}
