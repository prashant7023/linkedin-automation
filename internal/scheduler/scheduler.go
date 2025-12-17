package scheduler

import (
	"time"
)

// Scheduler handles activity timing and rate limiting
type Scheduler struct {
	workHourStart int
	workHourEnd   int
	breakInterval time.Duration
	lastBreak     time.Time
}

// NewScheduler creates a new scheduler instance
func NewScheduler(startHour, endHour int) *Scheduler {
	return &Scheduler{
		workHourStart: startHour,
		workHourEnd:   endHour,
		breakInterval: time.Hour,
		lastBreak:     time.Now(),
	}
}

// IsWorkingHours checks if current time is within working hours
func (s *Scheduler) IsWorkingHours() bool {
	hour := time.Now().Hour()
	return hour >= s.workHourStart && hour < s.workHourEnd
}

// ShouldTakeBreak determines if it's time for a break
func (s *Scheduler) ShouldTakeBreak() bool {
	return time.Since(s.lastBreak) >= s.breakInterval
}

// TakeBreak simulates a break period
func (s *Scheduler) TakeBreak() time.Duration {
	s.lastBreak = time.Now()
	// Break duration: 10-15 minutes
	return time.Duration(10+time.Now().Unix()%5) * time.Minute
}
