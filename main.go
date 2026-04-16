package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sarunm/arise-test/internal/account"
	"github.com/sarunm/arise-test/internal/customer"
	"github.com/sarunm/arise-test/internal/transaction"
	"github.com/sarunm/arise-test/pkg/cache"
	"github.com/sarunm/arise-test/pkg/database"
)

func main() {
	// Database
	db, err := database.New()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Cache
	c := cache.New()

	// Modules
	customerMod := customer.NewModule(db, c)
	accountMod := account.NewModule(db, c)
	transactionMod := transaction.NewModule(db, c, accountMod.Service)

	// Router
	r := gin.Default()

	v1 := r.Group("/api/v1")
	customerMod.RegisterRoutes(v1)
	accountMod.RegisterRoutes(v1)
	transactionMod.RegisterRoutes(v1)

	// Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
