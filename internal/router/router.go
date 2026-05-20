package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/guruorgoru/carevo/internal/ai"
	"github.com/guruorgoru/carevo/internal/auth"
	"github.com/guruorgoru/carevo/internal/cache"
	"github.com/guruorgoru/carevo/internal/handlers"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/jmoiron/sqlx"
)

func New(db *sqlx.DB, jwtService *auth.JWTService, cacheStore *cache.Store, aiWorker *ai.Worker, aiProvider ai.Provider) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Heartbeat("/ping"))

	authH := handlers.NewAuthHandler(db, jwtService)
	userH := handlers.NewUserHandler(db)
	careerH := handlers.NewCareerHandler(db, cacheStore)
	featureH := handlers.NewFeatureHandler(db, cacheStore)
	aiH := handlers.NewAIHandler(db, aiWorker, aiProvider)
	healthH := handlers.NewHealthHandler(db)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
	})

	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))
		r.Get("/me", userH.GetMe)
		r.Put("/me", userH.UpdateMe)
		r.Post("/me/bookmarks", featureH.BookmarkCareer)
		r.Get("/me/bookmarks", featureH.ListBookmarks)
	})

	r.Route("/challenges", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))
		r.Post("/", featureH.StartChallenge)
		r.Get("/active", featureH.ActiveChallenge)
		r.Post("/{id}/checkin", featureH.DailyCheckin)
		r.Get("/{id}/progress", featureH.ChallengeProgress)
	})

	r.Get("/careers", careerH.List)
	r.Get("/careers/search", careerH.Search)
	r.Get("/careers/{slug}", careerH.GetBySlug)
	r.Get("/careers/{id}/resources", careerH.Resources)
	r.Get("/careers/{id}/roadmap", careerH.Roadmap)
	r.Get("/careers/{id}/projects", featureH.ListProjectIdeas)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))
		r.Post("/careers/{id}/projects", featureH.CreateProjectIdea)
	})

	r.Get("/categories", careerH.Categories)

	r.Route("/ai", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))
		r.Use(middleware.NewRateLimiter(30, time.Minute).Middleware)
		r.Get("/survey/questions", aiH.SurveyQuestions)
		r.Post("/survey/submit", aiH.SubmitSurvey)
		r.Get("/results/me", aiH.GetResults)
		r.Get("/results/me/roadmap", aiH.GetRoadmap)
		r.Get("/careers/{id}/score", aiH.GetCareerScore)

		r.With(middleware.NewRateLimiter(5, time.Minute).Middleware).Post("/chat", aiH.ChatStream)
	})

	r.Get("/health", healthH.Check)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	return r
}
