package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
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
		log.Fatalf("Initializing zap logger %s", err)
	}
	defer func() {
		err := zapLogger.Sync()
		if err != nil {
			log.Fatalf("Syncing zap logger %s", err)
		}
	}()

	db, err := db.ConnectDB(databaseHost, databasePort, databaseUser, databasePassword, databaseName)
	if err != nil {
		log.Fatalf("Establishing connection to database %s", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("Closing database %s", err)
		}
	}()

	httpRequestTimeout := os.Getenv("HTTP_TIMEOUT")
	if httpRequestTimeout == "" {
		httpRequestTimeout = "10s"
	}

	httpRequestTimeoutDuration, err := time.ParseDuration(httpRequestTimeout)
	if err != nil {
		log.Fatalf("Invalid HTTP_TIMEOUT duration %s", err)
	}

	workerIteration := os.Getenv("WORKER_ITERATION")
	if workerIteration == "" {
		workerIteration = "30s"
	}

	workerIterationDuration, err := time.ParseDuration(workerIteration)
	if err != nil {
		log.Fatalf("Invalid WORKER_ITERATION duration %s", err)
	}

	rateLimit := os.Getenv("RATE_LIMIT")
	if rateLimit == "" {
		rateLimit = "1"
	}

	rateLimitInt, err := strconv.Atoi(rateLimit)
	if err != nil {
		log.Fatalf("Invalid RATE_LIMIT value %s", err)
	}

	numWorkers := os.Getenv("NUM_WORKERS")
	if numWorkers == "" {
		numWorkers = "5"
	}

	numWorkersInt, err := strconv.Atoi(numWorkers)
	if err != nil {
		log.Fatalf("Invalid NUM_WORKERS value %s", err)
	}

	retriesNum := os.Getenv("RETRIES_NUM")
	if retriesNum == "" {
		retriesNum = "5"
	}

	retriesNumInt, err := strconv.Atoi(retriesNum)
	if err != nil {
		log.Fatalf("Invalid RETRIES_NUM value %s", err)
	}

	limiter := rate.NewLimiter(rate.Every(10*time.Second), rateLimitInt)
	httpClient := &http.Client{
		Timeout: httpRequestTimeoutDuration,
	}
	exchangeratesapiApiKey := os.Getenv("EXCHANGERATESAPI_API_KEY")
	if exchangeratesapiApiKey == "" {
		log.Fatal("EXCHANGERATESAPI_API_KEY environment variable is not set")
	}
	exchangeratesapiBaseUrl := os.Getenv("EXCHANGERATESAPI_BASE_URL")
	if exchangeratesapiBaseUrl == "" {
		log.Fatal("EXCHANGERATESAPI_BASE_URL environment variable is not set")
	}
	quotaFetcher := quotafetcher.NewExchangeratesQuotaFetcher(httpClient, limiter, exchangeratesapiApiKey, exchangeratesapiBaseUrl, retriesNumInt)
	worker.NewWorker(db, workerIterationDuration, numWorkersInt, zapLogger, quotaFetcher).Start()
}
