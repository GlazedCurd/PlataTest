package quotafetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/time/rate"
)

type exchangeratesQuotaFetcher struct {
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	apiKey      string
	baseUrl     string
}

type exchangeratesResponse struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

func NewExchangeratesQuotaFetcher(httpClient *http.Client, limiter *rate.Limiter, apiKey string, baseUrl string) QuotaFetcher {
	return &exchangeratesQuotaFetcher{
		httpClient:  httpClient,
		rateLimiter: limiter,
		apiKey:      apiKey,
		baseUrl:     baseUrl,
	}
}

func (q *exchangeratesQuotaFetcher) FetchQuota(ctx context.Context, code string) (float64, error) {
	if err := q.rateLimiter.Wait(ctx); err != nil {
		return 0, fmt.Errorf("rate limit canceled: %w", err)
	}

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

	resp, err := q.httpClient.Get(u.String())
	if err != nil {
		return 0, fmt.Errorf("failed to fetch quota: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// TODO: отдельная обработка статусов
		return 0, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var response exchangeratesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return 0, fmt.Errorf("API request was not successful")
	}

	rate, ok := response.Rates[to]
	if !ok {
		return 0, fmt.Errorf("rate not found for currency: %s", to)
	}

	return rate, nil
}
