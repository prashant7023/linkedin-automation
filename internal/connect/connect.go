package connect

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

// Connector handles sending connection requests
type Connector struct {
	page           *rod.Page
	store          *storage.Store
	logger         *logger.Logger
	dailyLimit     int
	cooldownMin    int
	cooldownMax    int
	requestsSentToday int
}

// NewConnector creates a new connector instance
func NewConnector(page *rod.Page, store *storage.Store, log *logger.Logger) *Connector {
	dailyLimit := 10
	if limit := os.Getenv("DAILY_CONNECTION_LIMIT"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			dailyLimit = val
		}
	}
	
	cooldownMin := 5
	if min := os.Getenv("CONNECTION_COOLDOWN_MIN"); min != "" {
		if val, err := strconv.Atoi(min); err == nil {
			cooldownMin = val
		}
	}
	
	cooldownMax := 15
	if max := os.Getenv("CONNECTION_COOLDOWN_MAX"); max != "" {
		if val, err := strconv.Atoi(max); err == nil {
			cooldownMax = val
		}
	}
	
	return &Connector{
		page:           page,
		store:          store,
		logger:         log,
		dailyLimit:     dailyLimit,
		cooldownMin:    cooldownMin,
		cooldownMax:    cooldownMax,
		requestsSentToday: 0,
	}
}

// SendConnectionRequests sends connection requests to pending profiles
func (c *Connector) SendConnectionRequests() error {
	c.logger.Info("Starting connection request automation...")
	c.logger.Infof("Daily limit: %d requests", c.dailyLimit)
	
	// Get pending profiles from database
	profiles, err := c.store.GetPendingProfiles(c.dailyLimit)
	if err != nil {
		return fmt.Errorf("failed to get profiles: %w", err)
	}
	
	if len(profiles) == 0 {
		c.logger.Info("No pending profiles to process")
		return nil
	}
	
	c.logger.Infof("Found %d profiles to process", len(profiles))
	
	// Process each profile
	for i, profile := range profiles {
		if c.requestsSentToday >= c.dailyLimit {
			c.logger.Warn("⚠️  Daily limit reached, stopping for today")
			break
		}
		
		c.logger.Infof("Processing profile %d/%d: %s", i+1, len(profiles), profile.URL)
		
		// Check if already sent
		sent, err := c.store.HasSentRequest(profile.URL)
		if err != nil {
			c.logger.Warnf("Error checking profile: %v", err)
			continue
		}
		
		if sent {
			c.logger.Info("Already sent request to this profile, skipping")
			continue
		}
		
		// Send connection request
		if err := c.sendRequest(profile.URL, profile.Name); err != nil {
			c.logger.Errorf("Failed to send request: %v", err)
			c.store.LogAction("connect", profile.URL, "failure", err.Error())
			continue
		}
		
		c.requestsSentToday++
		c.logger.Infof("✅ Request sent successfully (%d/%d today)", c.requestsSentToday, c.dailyLimit)
		
		// Cooldown between requests (10 seconds for testing)
		if i < len(profiles)-1 && c.requestsSentToday < c.dailyLimit {
			cooldownSeconds := 10
			c.logger.Infof("Cooldown: waiting %d seconds before next request...", cooldownSeconds)
			time.Sleep(time.Duration(cooldownSeconds) * time.Second)
		}
	}
	
	c.logger.Infof("Connection request session complete. Sent %d requests today.", c.requestsSentToday)
	return nil
}

// sendRequest sends a connection request to a single profile
func (c *Connector) sendRequest(profileURL, name string) error {
	// Navigate to profile
	c.logger.Infof("Navigating to profile: %s", profileURL)
	if err := c.page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	
	time.Sleep(stealth.PageLoadWait())
	
	// Random scroll on profile (human behavior)
	c.logger.Info("Scrolling profile page...")
	c.randomProfileScroll()
	
	// Wait a bit (simulating reading)
	time.Sleep(stealth.ThinkTime())
	
	// Find and click Connect button
	c.logger.Info("Looking for Connect button...")
	
	// Try to find Connect button (multiple possible selectors)
	connectButton, err := c.findConnectButton()
	if err != nil {
		return fmt.Errorf("connect button not found: %w", err)
	}
	
	// Hover before clicking
	c.logger.Info("Hovering over Connect button...")
	time.Sleep(stealth.HoverDelay())
	
	// Click Connect with human-like mouse movement
	c.logger.Info("Clicking Connect button...")
	if err := c.clickWithMouse(connectButton); err != nil {
		return fmt.Errorf("failed to click connect: %w", err)
	}
	
	// Wait a bit for potential modal
	c.logger.Info("Waiting for response...")
	time.Sleep(2 * time.Second)
	
	// Check if note modal appeared (try multiple selectors with timeout)
	modalDetected := false
	modalSelectors := []string{
		"button[aria-label*='Send without a note']",
		"button[aria-label*='Send now']",
		"div[role='dialog'] button:has-text('Send')",
		"button.artdeco-button--primary:has-text('Send')",
	}
	
	for _, selector := range modalSelectors {
		has, _, _ := c.page.Timeout(3 * time.Second).Has(selector)
		if has {
			c.logger.Infof("Note modal detected (found: %s)", selector)
			modalDetected = true
			break
		}
	}
	
	if modalDetected {
		c.logger.Info("Adding personalized message...")
		if err := c.addNote(name); err != nil {
			c.logger.Warnf("Failed to add note: %v, clicking Send without note", err)
			// Try to send without note
			for _, selector := range []string{
				"button[aria-label*='Send without a note']",
				"button[aria-label*='Send now']",
			} {
				sendBtn, err := c.page.Timeout(2 * time.Second).Element(selector)
				if err == nil && sendBtn != nil {
					c.logger.Info("Clicking Send button...")
					sendBtn.Click(proto.InputMouseButtonLeft, 1)
					break
				}
			}
		}
	} else {
		c.logger.Info("No note modal detected, assuming request sent directly")
	}
	
	time.Sleep(1 * time.Second)
	
	// Save to database
	firstName := c.extractFirstName(name)
	note := fmt.Sprintf("Hi %s, great to connect!", firstName)
	if err := c.store.SaveConnectionRequest(profileURL, note); err != nil {
		c.logger.Warnf("Failed to save to database: %v", err)
	}
	
	c.store.LogAction("connect", profileURL, "success", "Connection request sent")
	
	return nil
}

// findConnectButton finds the Connect button using multiple strategies
func (c *Connector) findConnectButton() (*rod.Element, error) {
	// Try different selectors
	selectors := []string{
		"button[aria-label*='Invite'][aria-label*='to connect']",
		"button:has-text('Connect')",
		"button.pvs-profile-actions__action:has-text('Connect')",
		"div.pvs-profile-actions button:nth-child(1)",
	}
	
	for _, selector := range selectors {
		element, err := c.page.Element(selector)
		if err == nil && element != nil {
			return element, nil
		}
	}
	
	return nil, fmt.Errorf("connect button not found with any selector")
}

// addNote adds a personalized note to the connection request
func (c *Connector) addNote(name string) error {
	// Find note textarea (try multiple selectors)
	c.logger.Info("Looking for note textarea...")
	var textarea *rod.Element
	var err error
	
	textareaSelectors := []string{
		"textarea[name='message']",
		"textarea[id*='custom-message']",
		"textarea[placeholder*='Add a note']",
		"div[role='dialog'] textarea",
		"textarea.msg-form__textarea",
	}
	
	for _, selector := range textareaSelectors {
		c.logger.Infof("Trying selector: %s", selector)
		textarea, err = c.page.Timeout(3 * time.Second).Element(selector)
		if err == nil && textarea != nil {
			c.logger.Infof("Textarea found with selector: %s", selector)
			break
		}
	}
	
	if textarea == nil {
		return fmt.Errorf("note textarea not found with any selector")
	}
	
	// Scroll into view and ensure visible
	c.logger.Info("Scrolling textarea into view...")
	textarea.ScrollIntoView()
	time.Sleep(300 * time.Millisecond)
	textarea.WaitVisible()
	
	// Click to focus
	c.logger.Info("Clicking textarea to focus...")
	textarea.Click(proto.InputMouseButtonLeft, 1)
	time.Sleep(stealth.ShortPause())
	
	// Generate personalized note
	firstName := c.extractFirstName(name)
	note := fmt.Sprintf("Hi %s, I came across your profile and would love to connect!", firstName)
	
	// Type with human-like behavior
	c.logger.Info("Typing personalized note...")
	for _, char := range note {
		c.page.Keyboard.Type(input.Key(char))
		time.Sleep(stealth.TypeCharacter())
	}
	
	time.Sleep(stealth.ThinkTime())
	
	// Find and click Send button
	c.logger.Info("Looking for Send button...")
	var sendButton *rod.Element
	
	sendButtonSelectors := []string{
		"button[aria-label='Send now']",
		"button[aria-label*='Send invitation']",
		"button:has-text('Send')",
		"div[role='dialog'] button.artdeco-button--primary",
		"button.ml1",
	}
	
	for _, selector := range sendButtonSelectors {
		c.logger.Infof("Trying selector: %s", selector)
		sendButton, err = c.page.Timeout(3 * time.Second).Element(selector)
		if err == nil && sendButton != nil {
			c.logger.Infof("Send button found with selector: %s", selector)
			break
		}
	}
	
	if sendButton == nil {
		return fmt.Errorf("send button not found with any selector")
	}
	
	c.logger.Info("Clicking Send button...")
	if err := c.clickWithMouse(sendButton); err != nil {
		return err
	}
	
	return nil
}

// clickWithMouse clicks an element with human-like mouse movement
func (c *Connector) clickWithMouse(element *rod.Element) error {
	// Scroll element into view using JavaScript (more reliable than ScrollIntoView)
	c.logger.Info("Scrolling element into view...")
	_, err := element.Eval("() => this.scrollIntoView({behavior: 'smooth', block: 'center'})")
	if err != nil {
		c.logger.Warnf("Failed to scroll into view with JS: %v", err)
	}
	
	time.Sleep(800 * time.Millisecond) // Wait for smooth scroll
	
	// Try to wait for visibility but don't block if it fails
	c.logger.Info("Checking element visibility...")
	visibleErr := element.Timeout(2 * time.Second).WaitVisible()
	if visibleErr != nil {
		c.logger.Warnf("Element visibility check failed, proceeding anyway: %v", visibleErr)
	} else {
		c.logger.Info("Element is visible")
	}
	
	time.Sleep(stealth.HoverDelay())
	
	// Try multiple click methods
	c.logger.Info("Attempting to click element...")
	
	// Method 1: JavaScript click (most reliable, use first for stubborn elements)
	c.logger.Info("Trying JavaScript click...")
	_, jsErr := element.Eval("() => this.click()")
	if jsErr == nil {
		c.logger.Info("JavaScript click successful!")
		return nil
	}
	c.logger.Warnf("JavaScript click failed: %v, trying direct click", jsErr)
	
	// Method 2: Direct click with proto
	c.logger.Info("Trying direct click...")
	err = element.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		c.logger.Errorf("Direct click also failed: %v", err)
		
		// Method 3: Force click with MustClick (will panic if fails, but we'll recover)
		c.logger.Info("Trying MustClick as last resort...")
		defer func() {
			if r := recover(); r != nil {
				c.logger.Errorf("MustClick panicked: %v", r)
			}
		}()
		element.MustClick()
		c.logger.Info("MustClick succeeded!")
		return nil
	}
	
	c.logger.Info("Direct click successful!")
	return nil
}

// randomProfileScroll performs random scrolling on profile page
func (c *Connector) randomProfileScroll() {
	// Scroll down a bit
	distance := 300 + (500 / 2)
	c.page.Eval(fmt.Sprintf("() => window.scrollBy(0, %d)", distance))
	time.Sleep(stealth.GetScrollDelay())
	
	// Maybe scroll back
	if stealth.ShouldScrollBack() {
		c.page.Eval("() => window.scrollBy(0, -150)")
		time.Sleep(stealth.ShortPause())
	}
}

// extractFirstName extracts first name from full name
func (c *Connector) extractFirstName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[0]
	}
	return "there"
}

// GetRequestsSentToday returns the number of requests sent today
func (c *Connector) GetRequestsSentToday() int {
	return c.requestsSentToday
}
