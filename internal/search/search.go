package search

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

// Searcher handles LinkedIn profile search and discovery
type Searcher struct {
	page   *rod.Page
	store  *storage.Store
	logger *logger.Logger
}

// NewSearcher creates a new searcher instance
func NewSearcher(page *rod.Page, store *storage.Store, log *logger.Logger) *Searcher {
	return &Searcher{
		page:   page,
		store:  store,
		logger: log,
	}
}

// SearchProfiles searches for LinkedIn profiles based on criteria
func (s *Searcher) SearchProfiles() ([]string, error) {
	keywords := os.Getenv("SEARCH_KEYWORDS")
	location := os.Getenv("SEARCH_LOCATION")
	
	if keywords == "" {
		keywords = "Software Engineer"
	}
	
	s.logger.Infof("Searching for profiles: %s in %s", keywords, location)
	
	// Build search URL
	searchURL := s.buildSearchURL(keywords, location)
	s.logger.Infof("Navigating to: %s", searchURL)
	
	// Navigate to search
	if err := s.page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search: %w", err)
	}
	
	// Wait for page to fully load
	s.logger.Info("Waiting for search results to load...")
	time.Sleep(stealth.PageLoadWait())
	
	// Wait for search results container
	s.page.MustWaitLoad()
	time.Sleep(2 * time.Second) // Extra wait for dynamic content
	
	// Extract profile URLs with scrolling
	profileURLs, err := s.extractProfileURLs()
	if err != nil {
		return nil, fmt.Errorf("failed to extract profiles: %w", err)
	}
	
	s.logger.Infof("Found %d unique profiles", len(profileURLs))
	
	// Save profiles to database
	saved := 0
	for _, profileURL := range profileURLs {
		// Check if already exists
		exists, err := s.store.HasSentRequest(profileURL)
		if err != nil {
			s.logger.Warnf("Error checking profile: %v", err)
			continue
		}
		
		if !exists {
			// Extract name from URL or use placeholder
			name := s.extractNameFromURL(profileURL)
			if err := s.store.SaveProfile(profileURL, name); err != nil {
				s.logger.Warnf("Failed to save profile %s: %v", profileURL, err)
			} else {
				saved++
			}
		}
	}
	
	s.logger.Infof("Saved %d new profiles to database", saved)
	s.store.LogAction("search", searchURL, "success", fmt.Sprintf("Found %d profiles", len(profileURLs)))
	
	return profileURLs, nil
}

// buildSearchURL constructs the LinkedIn search URL with parameters
func (s *Searcher) buildSearchURL(keywords, location string) string {
	baseURL := "https://www.linkedin.com/search/results/people/"
	
	params := url.Values{}
	params.Add("keywords", keywords)
	if location != "" {
		params.Add("location", location)
	}
	params.Add("origin", "FACETED_SEARCH")
	
	return baseURL + "?" + params.Encode()
}

// extractProfileURLs extracts profile URLs from search results with scrolling
func (s *Searcher) extractProfileURLs() ([]string, error) {
	profileURLs := make(map[string]bool) // Use map to avoid duplicates
	lastCount := 0
	noChangeCount := 0
	maxNoChange := 3 // Stop if no new profiles after 3 scrolls
	
	s.logger.Info("Starting profile extraction with infinite scroll...")
	
	// Try multiple selectors to find profile links
	selectors := []string{
		"a.app-aware-link[href*='/in/']",
		"a[href*='/in/'][href*='linkedin.com']",
		"div.entity-result__item a[href*='/in/']",
		"li.reusable-search__result-container a[href*='/in/']",
		"a[href*='/in/']",
	}
	
	for {
		foundAny := false
		
		// Try each selector until we find profiles
		for _, selector := range selectors {
			elements, err := s.page.Elements(selector)
			if err != nil || len(elements) == 0 {
				continue
			}
			
			s.logger.Debugf("Found %d elements with selector: %s", len(elements), selector)
			
			// Extract URLs
			for _, element := range elements {
				href, err := element.Property("href")
				if err != nil {
					continue
				}
				
				hrefStr := href.String()
				
				// Filter valid profile URLs
				if strings.Contains(hrefStr, "/in/") && 
				   !strings.Contains(hrefStr, "/company/") &&
				   !strings.Contains(hrefStr, "/school/") &&
				   !strings.Contains(hrefStr, "/posts/") {
					// Clean URL (remove query parameters)
					if idx := strings.Index(hrefStr, "?"); idx != -1 {
						hrefStr = hrefStr[:idx]
					}
					// Ensure it ends with /
					if !strings.HasSuffix(hrefStr, "/") {
						hrefStr += "/"
					}
					
					// Only add if it looks like a valid profile URL
					if s.isValidProfileURL(hrefStr) {
						profileURLs[hrefStr] = true
						foundAny = true
					}
				}
			}
			
			// If we found profiles with this selector, don't try others
			if foundAny {
				break
			}
		}
		
		currentCount := len(profileURLs)
		s.logger.Infof("Extracted %d unique profiles so far...", currentCount)
		
		// Debug: If still no profiles after first attempt, check page content
		if currentCount == 0 && lastCount == 0 {
			s.logger.Warn("No profiles found. Checking page state...")
			
			// Check if we're actually on search results page
			pageURL := s.page.MustInfo().URL
			s.logger.Infof("Current page URL: %s", pageURL)
			
			// Check for common LinkedIn error messages
			if strings.Contains(pageURL, "/authwall") {
				s.logger.Error("Hit LinkedIn auth wall - session may have expired")
				break
			}
			
			// Try to find any link on the page to verify page loaded
			anyLinks, _ := s.page.Elements("a")
			s.logger.Infof("Total links found on page: %d", len(anyLinks))
		}
		
		// Check if we found new profiles
		if currentCount == lastCount {
			noChangeCount++
			if noChangeCount >= maxNoChange {
				s.logger.Info("No new profiles found after scrolling, stopping")
				break
			}
		} else {
			noChangeCount = 0
		}
		lastCount = currentCount
		
		// Scroll down with human-like behavior
		if err := s.humanScroll(); err != nil {
			s.logger.Warn("Scroll failed, stopping extraction")
			break
		}
		
		// Wait for content to load
		time.Sleep(stealth.GetScrollDelay())
		
		// Check max profiles limit
		maxProfiles := 50 // Default
		if maxStr := os.Getenv("MAX_PROFILES_TO_PROCESS"); maxStr != "" {
			fmt.Sscanf(maxStr, "%d", &maxProfiles)
		}
		
		if currentCount >= maxProfiles {
			s.logger.Infof("Reached maximum profile limit (%d)", maxProfiles)
			break
		}
	}
	
	// Convert map to slice
	result := make([]string, 0, len(profileURLs))
	for url := range profileURLs {
		result = append(result, url)
	}
	
	return result, nil
}

// humanScroll performs human-like scrolling
func (s *Searcher) humanScroll() error {
	// Get current scroll position
	currentY, err := s.page.Eval("() => window.scrollY")
	if err != nil {
		return err
	}
	
	// Scroll distance
	scrollDistance := stealth.GetScrollDistance()
	
	// Scroll with acceleration
	steps := 10 + (scrollDistance / 100) // More steps for longer scrolls
	
	for i := 0; i < steps; i++ {
		// Calculate intermediate position with easing
		progress := float64(i) / float64(steps)
		easedProgress := s.easeInOutQuad(progress)
		intermediateY := currentY.Value.Int() + int(float64(scrollDistance)*easedProgress)
		
		// Scroll to intermediate position
		s.page.Eval(fmt.Sprintf("() => window.scrollTo(0, %d)", intermediateY))
		
		// Variable delay
		time.Sleep(stealth.GetAcceleratedDelay(i, steps))
		
		// Random pause while scrolling (simulate reading)
		if stealth.ShouldPauseDuringScroll() {
			time.Sleep(stealth.GetScrollPauseDuration())
		}
	}
	
	// Occasionally scroll back a bit (human behavior)
	if stealth.ShouldScrollBack() {
		scrollBack := stealth.GetScrollBackDistance(scrollDistance)
		time.Sleep(stealth.ShortPause())
		s.page.Eval(fmt.Sprintf("() => window.scrollBy(0, -%d)", scrollBack))
	}
	
	return nil
}

// easeInOutQuad provides smooth acceleration/deceleration
func (s *Searcher) easeInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// isValidProfileURL checks if a URL is a valid LinkedIn profile
func (s *Searcher) isValidProfileURL(url string) bool {
	// Must contain /in/
	if !strings.Contains(url, "/in/") {
		return false
	}
	
	// Must be from linkedin.com
	if !strings.Contains(url, "linkedin.com") {
		return false
	}
	
	// Extract the username part
	parts := strings.Split(url, "/in/")
	if len(parts) < 2 {
		return false
	}
	
	username := strings.TrimSuffix(parts[1], "/")
	
	// Username should not be empty and should be reasonable length
	if len(username) < 3 || len(username) > 100 {
		return false
	}
	
	return true
}

// extractNameFromURL extracts a display name from profile URL
func (s *Searcher) extractNameFromURL(profileURL string) string {
	// Extract username from URL like https://www.linkedin.com/in/john-doe-123/
	parts := strings.Split(profileURL, "/in/")
	if len(parts) < 2 {
		return "Unknown"
	}
	
	username := strings.TrimSuffix(parts[1], "/")
	
	// Convert dashes to spaces and capitalize
	name := strings.ReplaceAll(username, "-", " ")
	name = strings.Title(strings.ToLower(name))
	
	return name
}
