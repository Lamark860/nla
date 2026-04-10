package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"nla/internal/handler"
	authMw "nla/internal/middleware"
)

func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Get("/health", h.Health)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth (public)
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		// Bonds (public)
		if h.Bond != nil {
			r.Get("/bonds/monthly", h.Bond.GetMonthlyBonds)
			r.Get("/bonds/grouped", h.Bond.GetBondsGrouped)
			r.Get("/bonds", h.Bond.ListBonds)
			r.Get("/bonds/{secid}", h.Bond.GetBond)
			r.Get("/bonds/{secid}/coupons", h.Bond.GetBondCoupons)
			r.Get("/bonds/{secid}/history", h.Bond.GetBondHistory)
			r.Post("/bonds/clear-cache", h.Bond.ClearCache)
			r.Post("/issuers/{id}/toggle", h.Bond.ToggleIssuer)
		}

		// Dohod.ru details (public)
		if h.Details != nil {
			r.Get("/bonds/{secid}/dohod", h.Details.GetDohodDetails)
		}

		// AI Analysis & Queue (public read, auth for write)
		if h.Analysis != nil {
			r.Get("/bonds/{secid}/analyses", h.Analysis.GetAnalyses)
			r.Get("/bonds/{secid}/analysis-stats", h.Analysis.GetAnalysisStats)
			r.Get("/analyses/{id}", h.Analysis.GetAnalysis)
			r.Get("/analyses/bulk-stats", h.Analysis.GetBulkAnalysisStats)
			r.Get("/jobs/{id}", h.Analysis.GetJobStatus)
			r.Get("/queue/stats", h.Analysis.GetQueueStats)
			r.Post("/bonds/{secid}/analyze", h.Analysis.StartAnalysis)
		}

		// Credit Ratings (public read)
		if h.Rating != nil {
			r.Get("/ratings", h.Rating.GetAllRatings)
			r.Get("/ratings/search", h.Rating.GetIssuerRating)
		}

		// Chat (public)
		if h.Chat != nil {
			r.Get("/chat/agents", h.Chat.GetAgents)
			r.Get("/chat/sessions", h.Chat.ListSessions)
			r.Post("/chat/sessions", h.Chat.CreateSession)
			r.Get("/chat/sessions/{id}", h.Chat.GetSession)
			r.Delete("/chat/sessions/{id}", h.Chat.DeleteSession)
			r.Post("/chat/sessions/{id}/messages", h.Chat.SendMessage)
		}

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.JWT(h.JWTSecret))

			r.Get("/auth/me", h.Me)

			// Rating management (auth required)
			if h.Rating != nil {
				r.Post("/ratings", h.Rating.UpsertRating)
				r.Post("/ratings/bulk", h.Rating.BulkUpsertRatings)
			}

			// Favorites (auth required)
			if h.Favorite != nil {
				r.Get("/favorites", h.Favorite.ListFavorites)
				r.Post("/favorites/toggle", h.Favorite.ToggleFavorite)
				r.Get("/favorites/check", h.Favorite.CheckFavorites)
				r.Post("/favorites/{secid}", h.Favorite.AddFavorite)
				r.Delete("/favorites/{secid}", h.Favorite.RemoveFavorite)
			}
		})
	})

	return r
}
