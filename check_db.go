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
	
	fmt.Println("=== CONNECTION REQUESTS ===")
	rows, err := db.Query("SELECT profile_url, sent_at FROM connection_requests ORDER BY sent_at DESC")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var url, sentAt string
		rows.Scan(&url, &sentAt)
		fmt.Printf("%d. %s (sent: %s)\n", count+1, url, sentAt)
		count++
	}
	
	fmt.Printf("\nTotal requests sent: %d\n", count)
	
	fmt.Println("\n=== ALL PROFILES ===")
	rows2, err := db.Query("SELECT url, status FROM profiles ORDER BY discovered_at DESC")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()
	
	total := 0
	for rows2.Next() {
		var url, status string
		rows2.Scan(&url, &status)
		total++
	}
	
	fmt.Printf("Total profiles in database: %d\n", total)
}
