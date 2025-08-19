package main

import (
	"log"
	"os"

	"github.com/GlazedCurd/PlataTest/internal/db"
	"github.com/GlazedCurd/PlataTest/internal/handler"
	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

func main() {
	databaseHost := os.Getenv("DATABASE_HOST")
	databasePort := os.Getenv("DATABASE_PORT")
	databaseUser := os.Getenv("DATABASE_USER")
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseName := os.Getenv("DATABASE_NAME")

	if databaseHost == "" || databasePort == "" || databaseUser == "" || databasePassword == "" || databaseName == "" {
		log.Fatal("Database environment variables are not set")
	}

	r := gin.Default()

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger %s", err)
	}
	defer zapLogger.Sync()

	// Initialize database connection
	db, err := db.ConnectDB(databaseHost, databasePort, databaseUser, databasePassword, databaseName)
	if err != nil {
		log.Fatalf("failed to establish connection to database %s", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database %s", err)
		}
	}()

	handler.SetupHandlers(r, db, zapLogger)

	// Start the HTTP server
	servicePort := os.Getenv("SERVICE_PORT")
	if servicePort == "" {
		servicePort = ":8080" // Default port if not set
	}
	r.Run(":" + servicePort)
}
