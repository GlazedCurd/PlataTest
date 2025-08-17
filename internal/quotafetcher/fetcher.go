package quotafetcher

import (
	"context"

	"go.uber.org/zap"
)

type QuotaFetcher interface {
	FetchQuota(ctx context.Context, code string, logger *zap.Logger) (float64, error)
}
