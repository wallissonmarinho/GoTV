package clock

import (
	"time"

	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// System implements ports.Clock.
type System struct{}

var _ ports.Clock = (*System)(nil)

// Now returns current time.
func (System) Now() time.Time {
	return time.Now()
}
