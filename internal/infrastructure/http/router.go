package http

import (
	"net/http"

	"apocapoc-api/internal/infrastructure/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(corsOrigins string, habitHandlers *HabitHandlers, authHandlers *AuthHandlers, jwtService *auth.JWTService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{corsOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandlers.Register)
		r.Post("/login", authHandlers.Login)
	})

	r.Route("/api/v1/habits", func(r chi.Router) {
		r.Use(AuthMiddleware(jwtService))
		r.Post("/", habitHandlers.CreateHabit)
		r.Get("/today", habitHandlers.GetTodaysHabits)
		r.Post("/{id}/mark", habitHandlers.MarkHabit)
	})

	return r
}
