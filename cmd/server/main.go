package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nvaditya/forge-backend/internal/config"
	"github.com/nvaditya/forge-backend/internal/db"
	"github.com/nvaditya/forge-backend/internal/middleware"
	"github.com/nvaditya/forge-backend/internal/routes"
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
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))

}
