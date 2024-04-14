package column

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsValidCol(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "simple_A",
			input:    "A",
			expected: true,
		},
		{
			name:     "simple_B",
			input:    "B",
			expected: true,
		},
		{
			name:     "multi",
			input:    "BA",
			expected: true,
		},
		{
			name:     "empty",
			input:    "",
			expected: false,
		},
		{
			name:     "lowercase",
			input:    "ab",
			expected: false,
		},
		{
			name:     "num",
			input:    "12",
			expected: false,
		},
		{
			name:     "invalid",
			input:    "❤️",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsValidCol(tc.input))
		})
	}
}
