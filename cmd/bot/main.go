package main

import (
	"fmt"
	"log"
	"os"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/storage"
)

func main() {
	// Initialize logger
	appLogger := logger.New("info")
	appLogger.Info("Starting LinkedIn Automation PoC")
	appLogger.Warn("⚠️  DISCLAIMER: Educational purposes only. Violates LinkedIn ToS.")

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./linkedin_bot.db"
	}

	store, err := storage.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()
	appLogger.Info("Database initialized successfully")

	// Initialize browser
	appLogger.Info("Launching browser...")
	browserInstance, err := browser.New()
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer browserInstance.Close()
	appLogger.Info("Browser launched successfully")

	fmt.Println("\n✅ Project initialized successfully!")
	fmt.Println("📁 All packages are ready")
	fmt.Println("🗄️  Database schema created")
	fmt.Println("🌐 Browser ready for automation")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Copy .env.example to .env and add your credentials")
	fmt.Println("2. Review config/config.yaml for settings")
	fmt.Println("3. Run: go mod tidy")
	fmt.Println("4. Start building Day 1 features!")
}
