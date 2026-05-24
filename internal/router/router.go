package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/guruorgoru/carevo/internal/ai"
	"github.com/guruorgoru/carevo/internal/auth"
	"github.com/guruorgoru/carevo/internal/cache"
	"github.com/guruorgoru/carevo/internal/email"
	"github.com/guruorgoru/carevo/internal/handlers"
	"github.com/guruorgoru/carevo/internal/middleware"
	"github.com/jmoiron/sqlx"
)

type Dependencies struct {
	DB               *sqlx.DB
	JWTService       *auth.JWTService
	CacheStore       *cache.Store
	AIWorker         *ai.Worker
	AIProvider       ai.Provider
	FrontendURL      string
	CloudinaryCloud  string
	CloudinaryKey    string
	CloudinarySecret string
	Mailer           *email.Sender
}

func New(deps Dependencies) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Mode", "Accept-Language"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Heartbeat("/ping"))

	authH := handlers.NewAuthHandler(deps.DB, deps.JWTService, deps.Mailer)
	oauthH := handlers.NewOAuthHandler(deps.DB, deps.JWTService)
	careerH := handlers.NewCareerHandler(deps.DB, deps.CacheStore)
	featureH := handlers.NewFeatureHandler(deps.DB, deps.CacheStore)
	aiH := handlers.NewAIHandler(deps.DB, deps.AIWorker, deps.AIProvider)
	healthH := handlers.NewHealthHandler(deps.DB)
	uploadH := handlers.NewUploadHandler(deps.DB, deps.CloudinaryCloud, deps.CloudinaryKey, deps.CloudinarySecret)
	adminH := handlers.NewAdminHandler(deps.DB, deps.Mailer)
	suggestionH := handlers.NewSuggestionHandler(deps.DB)
	verifyH := handlers.NewVerifyHandler(deps.DB, deps.Mailer, deps.JWTService)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		r.Post("/oauth", oauthH.Login)
		r.Post("/verify-send", verifyH.SendCode)
		r.Post("/verify", verifyH.Verify)
		r.Post("/resend-code", verifyH.SendCode)
	})

	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(deps.JWTService))
		r.Get("/me", oauthH.Me)
		r.Put("/me", oauthH.UpdateMe)
		r.Post("/me/avatar", uploadH.UploadAvatar)
		r.Post("/me/bookmarks", featureH.BookmarkCareer)
		r.Get("/me/bookmarks", featureH.ListBookmarks)
		r.Delete("/me/bookmarks/{career_id}", featureH.UnbookmarkCareer)
	})

	r.Route("/challenges", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(deps.JWTService))
		r.Post("/", featureH.StartChallenge)
		r.Get("/active", featureH.ActiveChallenge)
		r.Post("/{id}/checkin", featureH.DailyCheckin)
		r.Get("/{id}/progress", featureH.ChallengeProgress)
	})

	r.Route("/careers", func(r chi.Router) {
		r.Get("/", careerH.List)
		r.Get("/search", careerH.Search)
		r.Get("/advanced-search", careerH.AdvancedSearch)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(deps.JWTService))
			r.Post("/compare", careerH.Compare)
			r.Post("/compare/save", careerH.SaveComparison)
			r.Get("/compare", careerH.ListComparisons)
			r.Get("/compare/{id}", careerH.GetComparison)
			r.Delete("/compare/{id}", careerH.DeleteComparison)
		})

		r.With(middleware.NewRateLimiter(10, time.Minute).Middleware).Post("/{slug}/chat", aiH.CareerChat)
		r.Get("/{slug}", careerH.GetBySlug)
		r.Get("/{id}/resources", careerH.Resources)
		r.Get("/{id}/roadmap", careerH.Roadmap)
		r.Get("/{id}/projects", featureH.ListProjectIdeas)
		r.Get("/{id}/sources", careerH.ListSources)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(deps.JWTService))
			r.Post("/{id}/projects", featureH.CreateProjectIdea)
		})
	})

	r.Route("/suggestions", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(deps.JWTService))
		r.Post("/", suggestionH.Create)
		r.Get("/mine", suggestionH.ListMy)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(deps.JWTService))
		r.Group(func(r chi.Router) {
			r.Use(adminH.RequireAdmin)
			r.Get("/suggestions", adminH.ListSuggestions)
			r.Put("/suggestions/{id}/status", adminH.UpdateSuggestionStatus)
			r.Delete("/suggestions/{id}", adminH.DeleteSuggestion)
			r.Post("/promote", adminH.PromoteToAdmin)
		})
	})

	r.Get("/categories", careerH.Categories)

	r.Route("/ai", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(deps.JWTService))
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
