package stealth

import (
	"math/rand"
	"time"
)

// TypeCharacter simulates human typing for a single character
// Returns the delay before the next character should be typed
func TypeCharacter() time.Duration {
	// Base delay between 80-250ms
	baseDelay := rand.Intn(170) + 80
	return time.Duration(baseDelay) * time.Millisecond
}

// TypeString returns the total time it would take to type a string
// considering word pauses and natural rhythm
func TypeString(text string) time.Duration {
	totalDelay := time.Duration(0)
	
	for i, char := range text {
		// Regular typing delay
		totalDelay += TypeCharacter()
		
		// Longer pause after punctuation
		if char == '.' || char == ',' || char == '!' || char == '?' {
			totalDelay += time.Duration(rand.Intn(200)+100) * time.Millisecond
		}
		
		// Pause after space (between words)
		if char == ' ' && i > 0 {
			totalDelay += time.Duration(rand.Intn(100)+50) * time.Millisecond
		}
	}
	
	return totalDelay
}

// ShouldMakeTypo determines if a typo should be made (5% probability)
func ShouldMakeTypo() bool {
	return rand.Float64() < 0.05
}

// SimulateTypo generates a random typo character
// Usually a character adjacent on keyboard
func SimulateTypo(char rune) rune {
	// Common typos: adjacent keys on QWERTY keyboard
	adjacentKeys := map[rune][]rune{
		'a': {'s', 'q', 'w', 'z'},
		'e': {'w', 'r', 'd'},
		'i': {'u', 'o', 'k', 'j'},
		'o': {'i', 'p', 'l', 'k'},
		's': {'a', 'd', 'w', 'x'},
		't': {'r', 'y', 'g', 'f'},
	}
	
	if adjacent, exists := adjacentKeys[char]; exists {
		return adjacent[rand.Intn(len(adjacent))]
	}
	
	// Return same char if no adjacent keys defined
	return char
}

// GetWordPauseDelay returns a natural pause between words
func GetWordPauseDelay() time.Duration {
	return time.Duration(rand.Intn(150)+50) * time.Millisecond
}

// GetThinkingDelay returns a delay simulating thinking/reading time
func GetThinkingDelay() time.Duration {
	// 300-1200ms thinking time
	return time.Duration(rand.Intn(900)+300) * time.Millisecond
}
