package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nvaditya/forge-backend/internal/config"
	"github.com/nvaditya/forge-backend/internal/db"
	sqlcdb "github.com/nvaditya/forge-backend/internal/db/sqlc"
	"github.com/nvaditya/forge-backend/internal/handlers"
	"github.com/nvaditya/forge-backend/internal/middleware"
	"github.com/nvaditya/forge-backend/internal/repository"
	"github.com/nvaditya/forge-backend/internal/routes"
	"github.com/nvaditya/forge-backend/internal/utils"
)

func main() {

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	// Connect to the database
	pool, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer pool.Close()

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")

	queries := sqlcdb.New(pool)
	repos := repository.NewRepositories(queries)
	handlers.SetRepositories(repos)
	handlers.SetJWTSecret(cfg.JWTSecret)
	handlers.SetDB(pool)

	// Start background cleanup of expired in-memory revoked-token entries.
	utils.StartTokenCleanup(context.Background(), 5*time.Minute)

	router := mux.NewRouter()
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	//middleware logic
	router.Use(middleware.CORS)
	router.Use(middleware.Logging)

	// Handle CORS preflight for all routes
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	//router logic here
	routes.Routes(router, cfg, authMiddleware)

	// Start the server
	fmt.Printf("Server is running on port %s\n", cfg.Port)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())

}
