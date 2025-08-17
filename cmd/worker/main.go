package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GlazedCurd/PlataTest/internal/db"
	quotafetcher "github.com/GlazedCurd/PlataTest/internal/quotafetcher"
	"github.com/GlazedCurd/PlataTest/internal/worker"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
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

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger %s", err)
	}
	defer zapLogger.Sync()

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
	limiter := rate.NewLimiter(rate.Every(1*time.Second), 5)
	httpClient := &http.Client{
		Timeout: 10 * time.Second, // TODO: PASS timeout
	}
	exchangeratesapiApiKey := os.Getenv("EXCHANGERATESAPI_API_KEY")
	if exchangeratesapiApiKey == "" {
		log.Fatal("EXCHANGERATESAPI_API_KEY environment variable is not set")
	}
	exchangeratesapiBaseUrl := os.Getenv("EXCHANGERATESAPI_BASE_URL")
	if exchangeratesapiBaseUrl == "" {
		log.Fatal("EXCHANGERATESAPI_BASE_URL environment variable is not set")
	}
	quotaFetcher := quotafetcher.NewExchangeratesQuotaFetcher(httpClient, limiter, exchangeratesapiApiKey, exchangeratesapiBaseUrl)
	worker.NewWorker(db, 10*time.Second, 5, zapLogger, quotaFetcher).Start()
}
