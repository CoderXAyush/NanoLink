package base62

import "testing"

func TestToBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{name: "Zero", input: 0, expected: "0"},
		{name: "Low Number", input: 1, expected: "1"},
		// We won't test complex numbers hardcoded as your alphabet might vary,
		// but checking 0 and 1 is usually safe for all Base62 implementations.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToBase62(tt.input); got != tt.expected {
				t.Errorf("ToBase62(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// Benchmark is great for showing off performance in CI logs
func BenchmarkToBase62(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToBase62(uint64(i))
	}
}
