package browser

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Browser represents a Rod browser instance with stealth configuration
type Browser struct {
	instance *rod.Browser
	page     *rod.Page
}

// New creates and configures a new browser instance with anti-detection features
func New() (*Browser, error) {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Configure launcher with stealth settings
	l := launcher.New().
		Headless(false). // Non-headless for realistic behavior
		Set("disable-blink-features", "AutomationControlled").
		Set("excludeSwitches", "enable-automation").
		Set("disable-infobars", "true").
		Delete("enable-automation")

	// Get the launcher URL
	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Connect to browser
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	// Create new page
	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Set random viewport size
	width := rand.Intn(640) + 1280  // 1280-1920
	height := rand.Intn(360) + 720  // 720-1080
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  width,
		Height: height,
	}); err != nil {
		return nil, fmt.Errorf("failed to set viewport: %w", err)
	}

	// Inject stealth scripts to mask automation
	if err := maskFingerprint(page); err != nil {
		return nil, fmt.Errorf("failed to mask fingerprint: %w", err)
	}

	return &Browser{
		instance: browser,
		page:     page,
	}, nil
}

// maskFingerprint injects JavaScript to hide automation indicators
func maskFingerprint(page *rod.Page) error {
	stealthScript := `() => {
		// Override navigator.webdriver
		Object.defineProperty(navigator, 'webdriver', {
			get: () => false,
		});

		// Override chrome runtime
		window.chrome = {
			runtime: {},
		};

		// Override permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
			Promise.resolve({ state: Notification.permission }) :
			originalQuery(parameters)
		);

		// Override plugins
		Object.defineProperty(navigator, 'plugins', {
			get: () => [1, 2, 3, 4, 5],
		});

		// Override languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en'],
		});
	}`

	_, err := page.Eval(stealthScript)
	return err
}

// GetPage returns the current browser page
func (b *Browser) GetPage() *rod.Page {
	return b.page
}

// Close closes the browser instance
func (b *Browser) Close() error {
	if b.instance != nil {
		return b.instance.Close()
	}
	return nil
}
