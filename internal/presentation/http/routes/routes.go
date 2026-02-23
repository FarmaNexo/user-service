// internal/presentation/http/routes/routes.go
package routes

import (
	"net/http"

	"github.com/farmanexo/user-service/internal/presentation/http/controllers"
	"github.com/farmanexo/user-service/internal/presentation/http/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// SetupRoutes configura todas las rutas del servicio
func SetupRoutes(
	userController *controllers.UserController,
	authMiddleware *middlewares.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// ========================================
	// MIDDLEWARES GLOBALES
	// ========================================

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://farmanexo.pe"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middlewares.CorrelationID)

	// ========================================
	// SWAGGER DOCUMENTATION
	// ========================================

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:4002/swagger/doc.json"),
	))

	// ========================================
	// HEALTH CHECK
	// ========================================

	r.Get("/health", userController.HealthCheck)
	r.Get("/", userController.HealthCheck)

	// ========================================
	// API ROUTES - VERSION 1 (todas protegidas)
	// ========================================

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			// Todas las rutas de usuarios requieren autenticación
			r.Use(authMiddleware.RequireAuth)

			// Perfil
			r.Get("/me", userController.GetProfile)
			r.Put("/me", userController.UpdateProfile)

			// Avatar
			r.Put("/me/avatar", userController.UploadAvatar)
			r.Delete("/me/avatar", userController.DeleteAvatar)

			// Direcciones
			r.Get("/me/addresses", userController.ListAddresses)
			r.Post("/me/addresses", userController.CreateAddress)
			r.Put("/me/addresses/{id}", userController.UpdateAddress)
			r.Delete("/me/addresses/{id}", userController.DeleteAddress)

			// Preferencias
			r.Get("/me/preferences", userController.GetPreferences)
			r.Put("/me/preferences", userController.UpdatePreferences)
		})
	})

	// ========================================
	// API ROUTES - VERSION 2 (Futuro)
	// ========================================

	r.Route("/api/v2", func(r chi.Router) {
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API v2 - Próximamente"))
		})
	})

	return r
}
