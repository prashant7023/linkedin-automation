package message

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

// Messenger handles sending messages to accepted connections
type Messenger struct {
	page   *rod.Page
	store  *storage.Store
	logger *logger.Logger
}

// NewMessenger creates a new messenger instance
func NewMessenger(page *rod.Page, store *storage.Store, log *logger.Logger) *Messenger {
	return &Messenger{
		page:   page,
		store:  store,
		logger: log,
	}
}

// SendFollowUpMessages sends messages to accepted connections
func (m *Messenger) SendFollowUpMessages() (int, error) {
	m.logger.Info("Starting follow-up messaging session...")
	
	// Get accepted connections from database
	acceptedConnections, err := m.store.GetAcceptedConnections()
	if err != nil {
		return 0, fmt.Errorf("failed to get accepted connections: %w", err)
	}
	
	if len(acceptedConnections) == 0 {
		m.logger.Info("No accepted connections found to message")
		return 0, nil
	}
	
	m.logger.Infof("Found %d accepted connections to message", len(acceptedConnections))
	
	messagesSent := 0
	
	for i, profile := range acceptedConnections {
		m.logger.Infof("Processing connection %d/%d: %s", i+1, len(acceptedConnections), profile.URL)
		
		// Check if we already messaged this person
		hasMessaged, err := m.store.HasSentMessage(profile.URL)
		if err != nil {
			m.logger.Warnf("Error checking message status: %v", err)
			continue
		}
		
		if hasMessaged {
			m.logger.Info("Already sent message to this connection, skipping")
			continue
		}
		
		// Send message
		if err := m.sendMessage(profile.URL, profile.Name); err != nil {
			m.logger.Errorf("Failed to send message: %v", err)
			m.store.LogAction("message", profile.URL, "failure", err.Error())
			continue
		}
		
		// Save to database
		firstName := m.extractFirstName(profile.Name)
		messageText := fmt.Sprintf("Thanks for connecting, %s! Looking forward to staying in touch.", firstName)
		if err := m.store.SaveMessage(profile.URL, messageText); err != nil {
			m.logger.Warnf("Failed to save message to database: %v", err)
		}
		
		// Update profile status
		if err := m.store.UpdateProfileStatus(profile.URL, "messaged"); err != nil {
			m.logger.Warnf("Failed to update profile status: %v", err)
		}
		
		m.logger.Info("✅ Message sent successfully!")
		m.store.LogAction("message", profile.URL, "success", "Follow-up message sent")
		messagesSent++
		
		// Wait between messages to avoid detection
		if i < len(acceptedConnections)-1 {
			delay := time.Duration(10+stealth.Random(0, 20)) * time.Second
			m.logger.Infof("Waiting %v before next message...", delay)
			time.Sleep(delay)
		}
	}
	
	m.logger.Infof("Follow-up messaging complete. Sent %d messages.", messagesSent)
	return messagesSent, nil
}

// sendMessage sends a message to a connection
func (m *Messenger) sendMessage(profileURL, name string) error {
	// Navigate to profile
	m.logger.Info("Navigating to profile...")
	if err := m.page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	
	time.Sleep(stealth.PageLoadWait())
	
	// Random scroll
	m.randomScroll()
	
	// Find Message button
	m.logger.Info("Looking for Message button...")
	messageButton, err := m.findMessageButton()
	if err != nil {
		return fmt.Errorf("message button not found: %w", err)
	}
	
	// Click Message button
	m.logger.Info("Clicking Message button...")
	if err := m.clickElement(messageButton); err != nil {
		return fmt.Errorf("failed to click message button: %w", err)
	}
	
	time.Sleep(2 * time.Second)
	
	// Find message input
	m.logger.Info("Looking for message input...")
	messageInput, err := m.findMessageInput()
	if err != nil {
		return fmt.Errorf("message input not found: %w", err)
	}
	
	// Type message
	m.logger.Info("Typing message...")
	firstName := m.extractFirstName(name)
	messageText := fmt.Sprintf("Thanks for connecting, %s! Looking forward to staying in touch.", firstName)
	
	if err := m.typeMessage(messageInput, messageText); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}
	
	time.Sleep(stealth.ThinkTime())
	
	// Find and click Send button
	m.logger.Info("Looking for Send button...")
	sendButton, err := m.findSendButton()
	if err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}
	
	m.logger.Info("Clicking Send button...")
	if err := m.clickElement(sendButton); err != nil {
		return fmt.Errorf("failed to click send: %w", err)
	}
	
	time.Sleep(2 * time.Second)
	
	return nil
}

// findMessageButton finds the Message button on a profile
func (m *Messenger) findMessageButton() (*rod.Element, error) {
	selectors := []string{
		"button[aria-label*='Message']",
		"a[aria-label*='Message']",
		"button:has-text('Message')",
		"a.message-anywhere-button",
		"div.pvs-profile-actions button:has-text('Message')",
	}
	
	for _, selector := range selectors {
		element, err := m.page.Timeout(3 * time.Second).Element(selector)
		if err == nil && element != nil {
			m.logger.Infof("Message button found with selector: %s", selector)
			return element, nil
		}
	}
	
	return nil, fmt.Errorf("message button not found with any selector")
}

// findMessageInput finds the message input field
func (m *Messenger) findMessageInput() (*rod.Element, error) {
	selectors := []string{
		"div.msg-form__contenteditable",
		"div[role='textbox']",
		"div.msg-form__msg-content-container--scrollable",
		"div[contenteditable='true']",
	}
	
	for _, selector := range selectors {
		m.logger.Infof("Trying message input selector: %s", selector)
		element, err := m.page.Timeout(3 * time.Second).Element(selector)
		if err == nil && element != nil {
			m.logger.Infof("Message input found with selector: %s", selector)
			return element, nil
		}
	}
	
	return nil, fmt.Errorf("message input not found with any selector")
}

// findSendButton finds the Send button in messaging
func (m *Messenger) findSendButton() (*rod.Element, error) {
	selectors := []string{
		"button[type='submit'].msg-form__send-button",
		"button[aria-label*='Send']",
		"button.msg-form__send-button",
		"button:has-text('Send')",
	}
	
	for _, selector := range selectors {
		m.logger.Infof("Trying send button selector: %s", selector)
		element, err := m.page.Timeout(3 * time.Second).Element(selector)
		if err == nil && element != nil {
			m.logger.Infof("Send button found with selector: %s", selector)
			return element, nil
		}
	}
	
	return nil, fmt.Errorf("send button not found with any selector")
}

// typeMessage types a message with human-like behavior
func (m *Messenger) typeMessage(element *rod.Element, text string) error {
	// Click to focus
	element.Click(proto.InputMouseButtonLeft, 1)
	time.Sleep(stealth.ShortPause())
	
	// Type character by character
	for _, char := range text {
		m.page.Keyboard.Type(input.Key(char))
		time.Sleep(stealth.TypeCharacter())
	}
	
	return nil
}

// clickElement clicks an element with JavaScript (most reliable)
func (m *Messenger) clickElement(element *rod.Element) error {
	// Scroll into view
	_, err := element.Eval("() => this.scrollIntoView({behavior: 'smooth', block: 'center'})")
	if err != nil {
		m.logger.Warnf("Failed to scroll into view: %v", err)
	}
	
	time.Sleep(500 * time.Millisecond)
	time.Sleep(stealth.HoverDelay())
	
	// Try JavaScript click first (most reliable)
	_, jsErr := element.Eval("() => this.click()")
	if jsErr == nil {
		return nil
	}
	
	// Fallback to regular click
	return element.Click(proto.InputMouseButtonLeft, 1)
}

// randomScroll performs random scrolling on page
func (m *Messenger) randomScroll() {
	scrolls := stealth.Random(2, 4)
	for i := 0; i < scrolls; i++ {
		amount := stealth.Random(300, 800)
		m.page.Mouse.Scroll(0, float64(amount), 10)
		time.Sleep(stealth.GetAcceleratedDelay(i, scrolls))
	}
}

// extractFirstName extracts first name from full name
func (m *Messenger) extractFirstName(name string) string {
	if name == "" {
		return "there"
	}
	parts := strings.Fields(name)
	if len(parts) > 0 {
		return parts[0]
	}
	return "there"
}
