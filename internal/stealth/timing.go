package stealth

import (
	"math/rand"
	"time"
)

// RandomDelay returns a random delay within the specified range
func RandomDelay(minMs, maxMs int) time.Duration {
	delay := rand.Intn(maxMs-minMs) + minMs
	return time.Duration(delay) * time.Millisecond
}

// ThinkTime simulates human "thinking time" before taking an action
func ThinkTime() time.Duration {
	// 300-1200ms thinking time
	return RandomDelay(300, 1200)
}

// ActionDelay returns a delay between major actions
func ActionDelay() time.Duration {
	// 1-3 seconds between actions
	return RandomDelay(1000, 3000)
}

// ShortPause returns a short pause for minor interactions
func ShortPause() time.Duration {
	// 100-500ms for quick actions
	return RandomDelay(100, 500)
}

// ReadingDelay simulates time spent reading content
func ReadingDelay(contentLength int) time.Duration {
	// Assume ~250 words per minute reading speed
	// Average word length ~5 characters
	words := contentLength / 5
	minutesToRead := float64(words) / 250.0
	secondsToRead := minutesToRead * 60.0
	
	// Add randomness (±20%)
	variance := 0.2
	multiplier := 1.0 + (rand.Float64()*2-1)*variance
	
	finalSeconds := secondsToRead * multiplier
	return time.Duration(finalSeconds * float64(time.Second))
}

// Random returns a random number between min and max (inclusive)
func Random(min, max int) int {
	if max <= min {
		return min
	}
	return rand.Intn(max-min+1) + min
}

// HoverDelay returns time to hover over an element before clicking
func HoverDelay() time.Duration {
	// 200-800ms hover time
	return RandomDelay(200, 800)
}

// PageLoadWait returns time to wait for page to load
func PageLoadWait() time.Duration {
	// 2-4 seconds for page load
	return RandomDelay(2000, 4000)
}

// NetworkDelay simulates variable network conditions
func NetworkDelay() time.Duration {
	// 100-500ms for network variability
	return RandomDelay(100, 500)
}
