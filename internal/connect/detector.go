package connect

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

// ConnectionDetector detects accepted connection requests
type ConnectionDetector struct {
	page   *rod.Page
	store  *storage.Store
	logger *logger.Logger
}

// NewConnectionDetector creates a new detector instance
func NewConnectionDetector(page *rod.Page, store *storage.Store, log *logger.Logger) *ConnectionDetector {
	return &ConnectionDetector{
		page:   page,
		store:  store,
		logger: log,
	}
}

// DetectAcceptedConnections checks which sent connections were accepted
func (d *ConnectionDetector) DetectAcceptedConnections() (int, error) {
	d.logger.Info("Checking for accepted connections...")
	
	// Navigate to "My Network" page
	myNetworkURL := "https://www.linkedin.com/mynetwork/invite-connect/connections/"
	d.logger.Info("Navigating to My Network...")
	if err := d.page.Navigate(myNetworkURL); err != nil {
		return 0, fmt.Errorf("failed to navigate to my network: %w", err)
	}
	
	time.Sleep(stealth.PageLoadWait())
	
	// Scroll to load connections
	d.logger.Info("Loading connections list...")
	for i := 0; i < 3; i++ {
		d.page.Mouse.Scroll(0, float64(stealth.Random(500, 1000)), 10)
		time.Sleep(time.Duration(stealth.Random(1000, 2000)) * time.Millisecond)
	}
	
	// Get all pending connection requests from database
	pendingConnections, err := d.store.GetPendingConnectionRequests()
	if err != nil {
		return 0, fmt.Errorf("failed to get pending connections: %w", err)
	}
	
	if len(pendingConnections) == 0 {
		d.logger.Info("No pending connection requests to check")
		return 0, nil
	}
	
	d.logger.Infof("Checking %d pending connection requests...", len(pendingConnections))
	
	acceptedCount := 0
	
	// For each pending connection, visit profile and check if connected
	for i, profile := range pendingConnections {
		d.logger.Infof("Checking %d/%d: %s", i+1, len(pendingConnections), profile.URL)
		
		if err := d.checkIfAccepted(profile.URL); err != nil {
			d.logger.Warnf("Error checking profile: %v", err)
			continue
		}
		
		// Check if profile now shows "Message" button instead of "Connect"
		isAccepted, err := d.isConnectionAccepted()
		if err != nil {
			d.logger.Warnf("Could not determine connection status: %v", err)
			continue
		}
		
		if isAccepted {
			d.logger.Info("✅ Connection accepted!")
			// Mark as accepted in database
			if err := d.store.MarkConnectionAccepted(profile.URL); err != nil {
				d.logger.Warnf("Failed to mark as accepted: %v", err)
			}
			acceptedCount++
		} else {
			d.logger.Info("⏳ Still pending")
		}
		
		// Small delay between checks
		if i < len(pendingConnections)-1 {
			time.Sleep(time.Duration(stealth.Random(2, 5)) * time.Second)
		}
	}
	
	d.logger.Infof("Detection complete. %d new accepted connections found.", acceptedCount)
	return acceptedCount, nil
}

// checkIfAccepted navigates to a profile to check connection status
func (d *ConnectionDetector) checkIfAccepted(profileURL string) error {
	if err := d.page.Navigate(profileURL); err != nil {
		return err
	}
	
	time.Sleep(stealth.PageLoadWait())
	return nil
}

// isConnectionAccepted checks if a profile shows as connected (Message button present)
func (d *ConnectionDetector) isConnectionAccepted() (bool, error) {
	// Look for Message button (indicates connected)
	messageSelectors := []string{
		"button[aria-label*='Message']",
		"a[aria-label*='Message']",
		"button:has-text('Message')",
	}
	
	for _, selector := range messageSelectors {
		hasMessage, _, _ := d.page.Timeout(2 * time.Second).Has(selector)
		if hasMessage {
			return true, nil
		}
	}
	
	// Look for Connect button (indicates NOT connected yet)
	connectSelectors := []string{
		"button[aria-label*='Invite'][aria-label*='to connect']",
		"button:has-text('Connect')",
	}
	
	for _, selector := range connectSelectors {
		hasConnect, _, _ := d.page.Timeout(2 * time.Second).Has(selector)
		if hasConnect {
			return false, nil // Still pending
		}
	}
	
	// Look for "Pending" button (connection request sent but not accepted)
	pendingSelectors := []string{
		"button[aria-label*='Pending']",
		"button:has-text('Pending')",
	}
	
	for _, selector := range pendingSelectors {
		hasPending, _, _ := d.page.Timeout(2 * time.Second).Has(selector)
		if hasPending {
			return false, nil // Still pending
		}
	}
	
	return false, fmt.Errorf("could not determine connection status")
}
