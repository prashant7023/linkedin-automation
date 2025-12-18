package main

import (
	"database/sql"
	"fmt"
	"log"
	
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./linkedin_bot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// Mark first 3 connections as accepted for testing
	urls := []string{
		"https://www.linkedin.com/in/ssquare07/",
		"https://www.linkedin.com/in/aniketbewal/",
		"https://www.linkedin.com/in/shivam-agarwal-5b081222a/",
	}
	
	for _, url := range urls {
		_, err := db.Exec("UPDATE connection_requests SET accepted = 1 WHERE profile_url = ?", url)
		if err != nil {
			log.Printf("Failed to update %s: %v", url, err)
		} else {
			fmt.Printf("✅ Marked %s as accepted\n", url)
		}
	}
	
	fmt.Println("\n✅ Done! Run the bot again to see messaging in action.")
}
