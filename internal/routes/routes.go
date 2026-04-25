package routes

import (
	"github.com/gorilla/mux"
	"github.com/nvaditya/forge-backend/internal/config"
	"github.com/nvaditya/forge-backend/internal/handlers"
	"github.com/nvaditya/forge-backend/internal/middleware"
)

func Routes(router *mux.Router, cfg *config.Config, authMiddleware *middleware.AuthMiddleware) {

	// API versioning
	apiPrefix := cfg.APIVersion

	apiRouter := router.PathPrefix(apiPrefix).Subrouter()

	// Public user routes
	apiRouter.HandleFunc("/signup", handlers.CreateUser).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", handlers.LoginUser).Methods("POST", "OPTIONS")

	// Protected user routes
	protected := apiRouter.NewRoute().Subrouter()
	protected.Use(authMiddleware.Authenticate)
	protected.HandleFunc("/logout", handlers.LogoutUser).Methods("POST", "OPTIONS")
}
