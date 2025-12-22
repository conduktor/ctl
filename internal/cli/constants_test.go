package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeoutConstants(t *testing.T) {
	// Verify timeout values are reasonable
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.Equal(t, 5*time.Second, CancelTimeout)
	assert.Equal(t, 100*time.Millisecond, InitialPollInterval)
	assert.Equal(t, 2*time.Second, MaxPollInterval)

	// Ensure logical relationships
	assert.True(t, InitialPollInterval < MaxPollInterval, "initial poll interval should be less than max")
	assert.True(t, CancelTimeout < DefaultTimeout, "cancel timeout should be less than default timeout")
}

func TestValidStrategies(t *testing.T) {
	// Verify valid strategies map
	assert.True(t, ValidStrategies["fail-fast"])
	assert.True(t, ValidStrategies["continue-on-error"])
	assert.False(t, ValidStrategies["invalid"])
	assert.False(t, ValidStrategies[""])

	// Should have exactly 2 valid strategies
	assert.Len(t, ValidStrategies, 2)
}

func TestValidateStrategy_Extended(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantErr  bool
		errType  string
	}{
		{
			name:     "empty string is valid (means default)",
			strategy: "",
			wantErr:  false,
		},
		{
			name:     "fail-fast is valid",
			strategy: "fail-fast",
			wantErr:  false,
		},
		{
			name:     "continue-on-error is valid",
			strategy: "continue-on-error",
			wantErr:  false,
		},
		{
			name:     "invalid strategy returns error",
			strategy: "invalid",
			wantErr:  true,
			errType:  "*cli.InvalidStrategyError",
		},
		{
			name:     "typo in fail-fast",
			strategy: "fail-fsat",
			wantErr:  true,
		},
		{
			name:     "typo in continue-on-error",
			strategy: "continue-on-eror",
			wantErr:  true,
		},
		{
			name:     "case sensitive - uppercase rejected",
			strategy: "FAIL-FAST",
			wantErr:  true,
		},
		{
			name:     "extra whitespace rejected",
			strategy: " fail-fast",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStrategy(tt.strategy)
			if tt.wantErr {
				assert.Error(t, err)
				// Verify it's the correct error type
				_, ok := err.(*InvalidStrategyError)
				assert.True(t, ok, "expected InvalidStrategyError, got %T", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInvalidStrategyError(t *testing.T) {
	err := &InvalidStrategyError{Strategy: "bad-strategy"}

	// Test error message format
	assert.Equal(t, "invalid strategy 'bad-strategy': must be 'fail-fast' or 'continue-on-error'", err.Error())

	// Verify it implements error interface
	var _ error = err
}

func TestInvalidStrategyError_ErrorMessage(t *testing.T) {
	tests := []struct {
		strategy string
		expected string
	}{
		{
			strategy: "invalid",
			expected: "invalid strategy 'invalid': must be 'fail-fast' or 'continue-on-error'",
		},
		{
			strategy: "",
			expected: "invalid strategy '': must be 'fail-fast' or 'continue-on-error'",
		},
		{
			strategy: "FAIL-FAST",
			expected: "invalid strategy 'FAIL-FAST': must be 'fail-fast' or 'continue-on-error'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			err := &InvalidStrategyError{Strategy: tt.strategy}
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}
