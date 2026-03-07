package transcriber

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "00:00:00"},
		{15 * time.Second, "00:00:15"},
		{65 * time.Second, "00:01:05"},
		{3661 * time.Second, "01:01:01"},
		{10 * time.Hour, "10:00:00"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.duration)
		if got != tt.expected {
			t.Errorf("formatDuration(%v) = %v; want %v", tt.duration, got, tt.expected)
		}
	}
}
