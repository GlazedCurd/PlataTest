package quotafetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type exchangeratesQuotaFetcher struct {
	httpClient   *http.Client
	rateLimiter  *rate.Limiter
	apiKey       string
	baseUrl      string
	retriesLimit int
}

type exchangeratesResponse struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

func NewExchangeratesQuotaFetcher(httpClient *http.Client, limiter *rate.Limiter, apiKey string, baseUrl string, retriesLimit int) QuotaFetcher {
	return &exchangeratesQuotaFetcher{
		httpClient:   httpClient,
		rateLimiter:  limiter,
		apiKey:       apiKey,
		baseUrl:      baseUrl,
		retriesLimit: retriesLimit,
	}
}

func (q *exchangeratesQuotaFetcher) DoRequest(ctx context.Context, url *url.URL, to string) (float64, bool, error) {
	if err := q.rateLimiter.Wait(ctx); err != nil {
		return 0, false, fmt.Errorf("rate limit canceled: %w", err)
	}
	resp, err := q.httpClient.Get(url.String())
	if err != nil {
		return 0, true, fmt.Errorf("failed to fetch quota: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return 0, true, fmt.Errorf("server error: %s", resp.Status)
	}

	if resp.StatusCode >= 400 {
		return 0, false, fmt.Errorf("client request error: %s", resp.Status)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response exchangeratesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, false, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return 0, false, fmt.Errorf("API request was not successful")
	}

	rate, ok := response.Rates[to]
	if !ok {
		return 0, false, fmt.Errorf("rate not found for currency: %s", to)
	}
	return rate, false, nil
}

func (q *exchangeratesQuotaFetcher) FetchQuota(ctx context.Context, code string, logger *zap.Logger) (float64, error) {
	u, err := url.Parse(q.baseUrl)
	if err != nil {
		return 0, fmt.Errorf("failed to parse base URL: %w", err)
	}

	u.Path = "v1/latest"

	parts := strings.Split(code, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid code format: %s", code)
	}
	from := parts[0]
	to := parts[1]

	query := u.Query()
	query.Set("access_key", q.apiKey)
	query.Set("base", from)
	query.Set("symbols", to)
	u.RawQuery = query.Encode()
	currTimeout := 1
	var lastError error
	for i := 0; i < q.retriesLimit; i++ {
		quota, retry, err := q.DoRequest(ctx, u, to)
		if err != nil {
			logger.Error("Failed to fetch quota", zap.Error(err), zap.Int("retry", i))
			lastError = err
			if !retry {
				break
			}
			time.Sleep(time.Duration(currTimeout) * time.Second)
			currTimeout *= 2
			continue
		}
		return quota, nil
	}

	return 0, fmt.Errorf("failed to fetch quota after retiries %w", lastError)
}
