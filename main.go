package main

import (
	"log"
	"os"

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
	customerMod    := customer.NewModule(db, c)
	accountMod     := account.NewModule(db, c)
	transactionMod := transaction.NewModule(db, c, accountMod.Service)

	// Router
	r := gin.Default()

	v1 := r.Group("/api/v1")
	customerMod.RegisterRoutes(v1)
	accountMod.RegisterRoutes(v1)
	transactionMod.RegisterRoutes(v1)

	// Start
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
