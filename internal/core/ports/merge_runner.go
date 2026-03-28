package ports

import (
	"context"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
)

// MergeRunner executes one full playlist/EPG merge pass (scheduled or manual).
type MergeRunner interface {
	Run(ctx context.Context) domain.MergeResult
}
