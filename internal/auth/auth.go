package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

const (
	linkedInURL      = "https://www.linkedin.com"
	loginURL         = "https://www.linkedin.com/login"
	checkpointURL    = "/checkpoint"
	feedURL          = "/feed"
)

// Authenticator handles LinkedIn authentication and session management
type Authenticator struct {
	page   *rod.Page
	store  *storage.Store
	logger *logger.Logger
}

// NewAuthenticator creates a new authenticator instance
func NewAuthenticator(page *rod.Page, store *storage.Store, log *logger.Logger) *Authenticator {
	return &Authenticator{
		page:   page,
		store:  store,
		logger: log,
	}
}

// Login performs the login flow or restores existing session
func (a *Authenticator) Login() error {
	a.logger.Info("Starting authentication...")

	// Try to load existing session
	if err := a.tryRestoreSession(); err == nil {
		a.logger.Info("Session restored successfully")
		return nil
	}

	a.logger.Info("No valid session found, performing login...")
	
	// Get credentials from environment
	email := os.Getenv("LINKEDIN_EMAIL")
	password := os.Getenv("LINKEDIN_PASSWORD")
	
	if email == "" || password == "" {
		return fmt.Errorf("LINKEDIN_EMAIL and LINKEDIN_PASSWORD must be set in .env file")
	}

	// Navigate to login page
	a.logger.Info("Navigating to LinkedIn login page...")
	if err := a.page.Navigate(loginURL); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}
	
	// Wait for page to load
	time.Sleep(stealth.PageLoadWait())
	
	// Fill email field
	a.logger.Info("Filling email field...")
	if err := a.typeIntoField("#username", email); err != nil {
		return fmt.Errorf("failed to fill email: %w", err)
	}
	
	time.Sleep(stealth.ShortPause())
	
	// Fill password field
	a.logger.Info("Filling password field...")
	if err := a.typeIntoField("#password", password); err != nil {
		return fmt.Errorf("failed to fill password: %w", err)
	}
	
	time.Sleep(stealth.ThinkTime())
	
	// Click sign in button
	a.logger.Info("Clicking Sign in button...")
	if err := a.clickButton("button[type='submit']"); err != nil {
		return fmt.Errorf("failed to click sign in: %w", err)
	}
	
	// Wait for navigation
	time.Sleep(stealth.PageLoadWait())
	
	// Check for errors or checkpoints
	currentURL := a.page.MustInfo().URL
	
	if strings.Contains(currentURL, checkpointURL) {
		return a.handleCheckpoint()
	}
	
	if strings.Contains(currentURL, "/login") {
		return fmt.Errorf("login failed - check credentials or account status")
	}
	
	// Login successful - save session
	a.logger.Info("Login successful!")
	if err := a.saveSession(); err != nil {
		a.logger.Warnf("Failed to save session: %v", err)
	}
	
	// Log action to database
	a.store.LogAction("login", loginURL, "success", "Successfully logged in")
	
	return nil
}

// typeIntoField types text into an input field with human-like behavior
func (a *Authenticator) typeIntoField(selector, text string) error {
	element, err := a.page.Element(selector)
	if err != nil {
		return fmt.Errorf("element not found: %w", err)
	}
	
	// Click to focus
	if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}
	
	time.Sleep(stealth.ShortPause())
	
	// Type each character with human-like delays
	for _, char := range text {
		if err := a.page.Keyboard.Type(input.Key(char)); err != nil {
			return err
		}
		time.Sleep(stealth.TypeCharacter())
	}
	
	return nil
}

// clickButton clicks a button with human-like mouse movement
func (a *Authenticator) clickButton(selector string) error {
	element, err := a.page.Element(selector)
	if err != nil {
		return fmt.Errorf("button not found: %w", err)
	}
	
	// Hover before clicking
	time.Sleep(stealth.HoverDelay())
	
	// Click element
	if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}
	
	return nil
}

// handleCheckpoint handles security checkpoints (2FA, captcha, etc.)
func (a *Authenticator) handleCheckpoint() error {
	a.logger.Warn("⚠️  Security checkpoint detected!")
	a.logger.Warn("This may be:")
	a.logger.Warn("  - Two-factor authentication (2FA)")
	a.logger.Warn("  - Email verification")
	a.logger.Warn("  - Captcha challenge")
	a.logger.Warn("  - Unusual activity detection")
	a.logger.Warn("")
	a.logger.Warn("Action required:")
	a.logger.Warn("  1. Complete the verification manually in the browser")
	a.logger.Warn("  2. The bot will wait for 60 seconds")
	a.logger.Warn("  3. If successful, the bot will continue")
	
	// Log checkpoint
	a.store.LogAction("login", "checkpoint", "blocked", "Security checkpoint detected")
	
	// Wait for manual intervention
	a.logger.Info("Waiting 60 seconds for manual verification...")
	time.Sleep(60 * time.Second)
	
	// Check if checkpoint was cleared
	currentURL := a.page.MustInfo().URL
	if strings.Contains(currentURL, checkpointURL) {
		return fmt.Errorf("checkpoint still active - manual intervention required")
	}
	
	a.logger.Info("Checkpoint cleared successfully!")
	
	// Save session after checkpoint
	if err := a.saveSession(); err != nil {
		a.logger.Warnf("Failed to save session: %v", err)
	}
	
	return nil
}

// saveSession saves the current browser cookies to database
func (a *Authenticator) saveSession() error {
	cookies, err := a.page.Cookies([]string{linkedInURL})
	if err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}
	
	// Serialize cookies to JSON
	cookiesJSON, err := json.Marshal(cookies)
	if err != nil {
		return fmt.Errorf("failed to serialize cookies: %w", err)
	}
	
	// Save to database
	if err := a.store.SaveSession(string(cookiesJSON)); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	
	a.logger.Info("Session saved successfully")
	return nil
}

// tryRestoreSession attempts to restore a previous session
func (a *Authenticator) tryRestoreSession() error {
	cookiesJSON, err := a.store.LoadSession()
	if err != nil {
		return fmt.Errorf("no session found: %w", err)
	}
	
	if cookiesJSON == "" {
		return fmt.Errorf("no session data")
	}
	
	// Deserialize cookies
	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal([]byte(cookiesJSON), &cookies); err != nil {
		return fmt.Errorf("failed to deserialize cookies: %w", err)
	}
	
	// Navigate to LinkedIn first
	if err := a.page.Navigate(linkedInURL); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	
	time.Sleep(2 * time.Second)
	
	// Set cookies
	cookieParams := make([]*proto.NetworkCookieParam, len(cookies))
	for i, c := range cookies {
		cookieParams[i] = &proto.NetworkCookieParam{
			Name:   c.Name,
			Value:  c.Value,
			Domain: c.Domain,
			Path:   c.Path,
		}
	}
	if err := a.page.SetCookies(cookieParams); err != nil {
		return fmt.Errorf("failed to set cookies: %w", err)
	}
	
	// Navigate to feed to verify session
	if err := a.page.Navigate(linkedInURL + feedURL); err != nil {
		return fmt.Errorf("failed to navigate to feed: %w", err)
	}
	
	time.Sleep(stealth.PageLoadWait())
	
	// Check if we're logged in
	currentURL := a.page.MustInfo().URL
	if strings.Contains(currentURL, "/login") {
		return fmt.Errorf("session expired")
	}
	
	a.logger.Info("Session is still valid")
	return nil
}

// IsLoggedIn checks if currently logged into LinkedIn
func (a *Authenticator) IsLoggedIn() bool {
	currentURL := a.page.MustInfo().URL
	return !strings.Contains(currentURL, "/login") && !strings.Contains(currentURL, checkpointURL)
}
