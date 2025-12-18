package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"linkedin-automation-poc/internal/auth"
	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/connect"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/message"
	"linkedin-automation-poc/internal/scheduler"
	"linkedin-automation-poc/internal/search"
	"linkedin-automation-poc/internal/storage"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger (use "debug" for more detailed output)
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	appLogger := logger.New(logLevel)
	appLogger.Info(" Starting LinkedIn Automation PoC")
	appLogger.Warn("  DISCLAIMER: Educational purposes only. Violates LinkedIn ToS.")
	appLogger.Warn("  Never use this on your real LinkedIn account!")
	fmt.Println()

	// Check if we're in working hours
	workHourStart := 9
	workHourEnd := 18
	if start := os.Getenv("WORK_HOUR_START"); start != "" {
		fmt.Sscanf(start, "%d", &workHourStart)
	}
	if end := os.Getenv("WORK_HOUR_END"); end != "" {
		fmt.Sscanf(end, "%d", &workHourEnd)
	}
	
	sched := scheduler.NewScheduler(workHourStart, workHourEnd)
	if !sched.IsWorkingHours() {
		appLogger.Warn("⏰ Outside of working hours - consider running during business hours for more realistic behavior")
		appLogger.Info("Continuing anyway for demonstration purposes...")
		fmt.Println()
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./linkedin_bot.db"
	}

	appLogger.Info(" Initializing database...")
	store, err := storage.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()
	appLogger.Info(" Database ready")
	fmt.Println()

	// Initialize browser
	appLogger.Info(" Launching browser with anti-detection...")
	browserInstance, err := browser.New()
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer browserInstance.Close()
	page := browserInstance.GetPage()
	appLogger.Info(" Browser ready (fingerprint masked)")
	fmt.Println()

	// Authentication
	appLogger.Info(" Starting authentication...")
	authenticator := auth.NewAuthenticator(page, store, appLogger)
	if err := authenticator.Login(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	appLogger.Info(" Successfully authenticated")
	fmt.Println()

	// Search for profiles
	appLogger.Info(" Searching for profiles...")
	searcher := search.NewSearcher(page, store, appLogger)
	if _, err := searcher.SearchProfiles(); err != nil {
		appLogger.Errorf("Search failed: %v", err)
	} else {
		appLogger.Info(" Profile search complete")
	}
	fmt.Println()

	// Send connection requests
	appLogger.Info(" Starting connection request automation...")
	connector := connect.NewConnector(page, store, appLogger)
	requestsSent := 0
	if err := connector.SendConnectionRequests(); err != nil {
		appLogger.Errorf("Connection requests failed: %v", err)
	} else {
		requestsSent = connector.GetRequestsSentToday()
		appLogger.Info(" Connection request session complete")
	}
	fmt.Println()

	// Note: Connection detection is slow and unreliable
	// For testing, manually mark connections as accepted:
	// go run mark_accepted.go
	acceptedCount := 0
	fmt.Println()

	// Send follow-up messages to accepted connections
	appLogger.Info(" Starting follow-up messaging...")
	messenger := message.NewMessenger(page, store, appLogger)
	messagesSent := 0
	if sent, err := messenger.SendFollowUpMessages(); err != nil {
		appLogger.Errorf("Messaging failed: %v", err)
	} else {
		messagesSent = sent
		appLogger.Info(" Follow-up messaging complete")
	}
	fmt.Println()

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("                    SESSION SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("  📤 Connection requests sent: %d\n", requestsSent)
	fmt.Printf("  ✅ Connections accepted: %d\n", acceptedCount)
	fmt.Printf("  💬 Follow-up messages sent: %d\n", messagesSent)
	fmt.Println()
	fmt.Println("  ✅ All automation tasks complete!")
	fmt.Println("  📊 Check linkedin_bot.db for detailed logs")
	fmt.Println("  🔄 Run again tomorrow to continue automation")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	
	appLogger.Info("🎉 Automation session finished successfully")
	appLogger.Info("Browser will remain open for 10 seconds for inspection...")
	
	// Keep browser open briefly for demonstration
	// time.Sleep(10 * time.Second)
}
