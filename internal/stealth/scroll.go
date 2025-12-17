package stealth

import (
	"math/rand"
	"time"
)

// ScrollParams defines parameters for a scroll action
type ScrollParams struct {
	Distance      int           // Total distance to scroll (pixels)
	Steps         int           // Number of scroll steps
	Acceleration  bool          // Whether to accelerate/decelerate
	RandomPause   bool          // Add random pauses during scroll
	ScrollBack    bool          // Occasionally scroll back up
}

// GetScrollDelay returns a random delay between scroll actions
func GetScrollDelay() time.Duration {
	// 100-300ms between scroll steps
	return time.Duration(rand.Intn(200)+100) * time.Millisecond
}

// GetScrollDistance returns a random scroll distance
func GetScrollDistance() int {
	// Scroll 300-800 pixels at a time
	return rand.Intn(500) + 300
}

// ShouldScrollBack determines if we should scroll back up (10% chance)
func ShouldScrollBack() bool {
	return rand.Float64() < 0.10
}

// GetScrollBackDistance returns how far to scroll back
func GetScrollBackDistance(originalDistance int) int {
	// Scroll back 20-50% of original distance
	percentage := float64(rand.Intn(30)+20) / 100.0
	return int(float64(originalDistance) * percentage)
}

// GetAcceleratedDelay returns a delay that changes based on scroll progression
// Simulates natural acceleration and deceleration
func GetAcceleratedDelay(step, totalSteps int) time.Duration {
	// Create acceleration curve
	progress := float64(step) / float64(totalSteps)
	
	var multiplier float64
	if progress < 0.2 {
		// Start slow (acceleration phase)
		multiplier = 2.0 - (progress * 5)
	} else if progress > 0.8 {
		// End slow (deceleration phase)
		multiplier = 1.0 + ((progress - 0.8) * 5)
	} else {
		// Middle: full speed
		multiplier = 1.0
	}
	
	baseDelay := 150.0
	delayMs := int(baseDelay * multiplier)
	
	// Add random variation
	delayMs += rand.Intn(50) - 25
	
	return time.Duration(delayMs) * time.Millisecond
}

// ShouldPauseDuringScroll determines if we should pause while scrolling (15% chance)
func ShouldPauseDuringScroll() bool {
	return rand.Float64() < 0.15
}

// GetScrollPauseDuration returns how long to pause during scrolling
func GetScrollPauseDuration() time.Duration {
	// Pause for 500-1500ms (reading content)
	return time.Duration(rand.Intn(1000)+500) * time.Millisecond
}
