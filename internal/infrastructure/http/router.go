package http

import (
	"net/http"
	"time"

	"apocapoc-api/internal/infrastructure/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "apocapoc-api/docs"
)

func NewRouter(corsOrigins string, habitHandlers *HabitHandlers, authHandlers *AuthHandlers, statsHandlers *StatsHandlers, healthHandlers *HealthHandlers, jwtService *auth.JWTService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{corsOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/api/v1/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/api/v1/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/api/v1/docs/doc.json"),
	))

	r.Get("/api/v1/health", healthHandlers.Health)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 1*time.Minute))
		r.Post("/register", authHandlers.Register)
		r.Post("/login", authHandlers.Login)
		r.Post("/refresh", authHandlers.Refresh)
		r.Post("/logout", authHandlers.Logout)
	})

	r.Route("/api/v1/habits", func(r chi.Router) {
		r.Use(AuthMiddleware(jwtService))
		r.Post("/", habitHandlers.CreateHabit)
		r.Get("/", habitHandlers.GetUserHabits)
		r.Get("/today", habitHandlers.GetTodaysHabits)
		r.Get("/{id}", habitHandlers.GetHabitByID)
		r.Put("/{id}", habitHandlers.UpdateHabit)
		r.Delete("/{id}", habitHandlers.ArchiveHabit)
		r.Get("/{id}/entries", habitHandlers.GetHabitEntries)
		r.Post("/{id}/mark", habitHandlers.MarkHabit)
		r.Delete("/{id}/entries/{date}", habitHandlers.UnmarkHabit)
	})

	r.Route("/api/v1/stats", func(r chi.Router) {
		r.Use(AuthMiddleware(jwtService))
		r.Get("/habits/{id}", statsHandlers.GetHabitStats)
	})

	return r
}
