package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kurtgray/blog-api-go/internal/config"
	"github.com/kurtgray/blog-api-go/internal/database"
	"github.com/kurtgray/blog-api-go/internal/handlers"
	"github.com/kurtgray/blog-api-go/internal/middleware"
	"github.com/kurtgray/blog-api-go/internal/repository"
	"github.com/kurtgray/blog-api-go/internal/router"
)

func main() {
	// load config
	cfg := config.Load()

	// connect db w. config string
	db, err := database.Connect(cfg.MongoDB)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Disconnect()

	// init repos
	userRepo := repository.NewUserRepository(db.Database)
	postRepo := repository.NewPostRepository(db.Database)
	commentRepo := repository.NewCommentRepository(db.Database)

	// init auth service
	authService := middleware.NewAuthService(userRepo, cfg.JWTSecret)

	// init handlers
	userHandler := handlers.NewUserHandler(userRepo, authService)
	postHandler := handlers.NewPostHandler(postRepo, userRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo)

	// init CORS
	corsMiddleware := middleware.SetupCORS()

	// router setup
	rt := router.New(userHandler, postHandler, commentHandler, authService, corsMiddleware)
	r := rt.Setup()

	// create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
