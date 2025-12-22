package cli

import "time"

// API versioning constants - single source of truth
const (
	ApiVersion       = "1.0"
	ApiVersionHeader = "X-Conduktor-API-Version"
)

// HTTP timeouts
const (
	DefaultTimeout      = 30 * time.Second
	CancelTimeout       = 5 * time.Second
	InitialPollInterval = 100 * time.Millisecond
	MaxPollInterval     = 2 * time.Second
)

// Valid strategy values - must match Scala ApplyStrategy enum
var ValidStrategies = map[string]bool{
	"fail-fast":         true,
	"continue-on-error": true,
}

// ValidateStrategy returns an error if the strategy is not valid
func ValidateStrategy(strategy string) error {
	if strategy == "" {
		return nil // empty means default
	}
	if !ValidStrategies[strategy] {
		return &InvalidStrategyError{Strategy: strategy}
	}
	return nil
}

type InvalidStrategyError struct {
	Strategy string
}

func (e *InvalidStrategyError) Error() string {
	return "invalid strategy '" + e.Strategy + "': must be 'fail-fast' or 'continue-on-error'"
}
