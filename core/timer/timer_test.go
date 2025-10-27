package timer

import (
	"testing"
	"time"
)

func TestNewTimer(t *testing.T) {
	duration := 5 * time.Minute
	timer := NewTimer(duration, "test")

	if timer.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, timer.Duration)
	}

	if timer.Label != "test" {
		t.Errorf("Expected label 'test', got '%s'", timer.Label)
	}

	if timer.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestTimerRemaining(t *testing.T) {
	duration := 2 * time.Second
	timer := NewTimer(duration, "")

	// Immediately after creation, remaining should be approximately equal to duration
	remaining := timer.Remaining()
	if remaining > duration || remaining < duration-100*time.Millisecond {
		t.Errorf("Expected remaining ~%v, got %v", duration, remaining)
	}

	// Wait a bit and check again
	time.Sleep(500 * time.Millisecond)
	remaining = timer.Remaining()
	expected := duration - 500*time.Millisecond
	if remaining > expected+100*time.Millisecond || remaining < expected-100*time.Millisecond {
		t.Errorf("Expected remaining ~%v, got %v", expected, remaining)
	}
}

func TestTimerIsExpired(t *testing.T) {
	duration := 100 * time.Millisecond
	timer := NewTimer(duration, "")

	if timer.IsExpired() {
		t.Error("Timer should not be expired immediately after creation")
	}

	time.Sleep(150 * time.Millisecond)

	if !timer.IsExpired() {
		t.Error("Timer should be expired after waiting longer than duration")
	}
}

func TestTimerPercentComplete(t *testing.T) {
	duration := 1 * time.Second
	timer := NewTimer(duration, "")

	// Should be near 0% at start
	percent := timer.PercentComplete()
	if percent > 5.0 {
		t.Errorf("Expected percent near 0, got %f", percent)
	}

	// Wait half the duration
	time.Sleep(500 * time.Millisecond)
	percent = timer.PercentComplete()
	if percent < 40.0 || percent > 60.0 {
		t.Errorf("Expected percent around 50, got %f", percent)
	}

	// Wait past expiration
	time.Sleep(600 * time.Millisecond)
	percent = timer.PercentComplete()
	if percent != 100.0 {
		t.Errorf("Expected percent 100 after expiration, got %f", percent)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{5 * time.Second, "5s"},
		{30 * time.Second, "30s"},
		{1 * time.Minute, "1m0s"},
		{2*time.Minute + 30*time.Second, "2m30s"},
		{1 * time.Hour, "1h0m0s"},
		{1*time.Hour + 30*time.Minute + 45*time.Second, "1h30m45s"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

func TestFormatDurationShort(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{5 * time.Second, "5s"},
		{30 * time.Second, "30s"},
		{1 * time.Minute, "1m0s"},
		{2*time.Minute + 30*time.Second, "2m30s"},
		{1 * time.Hour, "1h0m"},
		{1*time.Hour + 30*time.Minute + 45*time.Second, "1h30m"},
	}

	for _, tt := range tests {
		result := FormatDurationShort(tt.duration)
		if result != tt.expected {
			t.Errorf("FormatDurationShort(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

