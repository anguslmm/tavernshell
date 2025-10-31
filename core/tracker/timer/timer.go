package timer

import (
	"fmt"
	"time"
)

// Timer represents a countdown timer
type Timer struct {
	ID        string
	StartTime time.Time
	Duration  time.Duration
	Label     string // optional label for the timer
}

// NewTimer creates a new timer with the specified duration
func NewTimer(duration time.Duration, label string) *Timer {
	return &Timer{
		ID:        generateID(),
		StartTime: time.Now(),
		Duration:  duration,
		Label:     label,
	}
}

// Remaining returns the time remaining on the timer
func (t *Timer) Remaining() time.Duration {
	elapsed := time.Since(t.StartTime)
	remaining := t.Duration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Elapsed returns the time elapsed since the timer started
func (t *Timer) Elapsed() time.Duration {
	elapsed := time.Since(t.StartTime)
	if elapsed > t.Duration {
		return t.Duration
	}
	return elapsed
}

// PercentComplete returns the percentage of time elapsed (0-100)
func (t *Timer) PercentComplete() float64 {
	if t.Duration == 0 {
		return 100.0
	}
	elapsed := time.Since(t.StartTime)
	percent := (float64(elapsed) / float64(t.Duration)) * 100.0
	if percent > 100.0 {
		return 100.0
	}
	return percent
}

// IsExpired returns true if the timer has expired
func (t *Timer) IsExpired() bool {
	return time.Since(t.StartTime) >= t.Duration
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// FormatDurationShort formats a duration in a compact way
func FormatDurationShort(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// generateID generates a simple unique ID for a timer
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

