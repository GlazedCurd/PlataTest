package quotafetcher

import "context"

type QuotaFetcher interface {
	FetchQuota(ctx context.Context, code string) (float64, error)
}
