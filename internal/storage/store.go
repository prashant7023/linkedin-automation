package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store represents the database storage layer
type Store struct {
	db *sql.DB
}

// Profile represents a LinkedIn profile
type Profile struct {
	ID           int
	URL          string
	Name         string
	Status       string
	DiscoveredAt time.Time
	LastActionAt *time.Time
}

// NewStore creates a new database store and initializes the schema
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &Store{db: db}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates all required database tables
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT UNIQUE NOT NULL,
		name TEXT,
		status TEXT DEFAULT 'discovered',
		discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_action_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS connection_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id INTEGER,
		profile_url TEXT NOT NULL,
		note TEXT,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		accepted BOOLEAN DEFAULT 0,
		FOREIGN KEY (profile_id) REFERENCES profiles(id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile_id INTEGER,
		profile_url TEXT NOT NULL,
		message_text TEXT,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (profile_id) REFERENCES profiles(id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cookies TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		is_valid BOOLEAN DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS actions_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action_type TEXT,
		target_url TEXT,
		status TEXT,
		details TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_profiles_status ON profiles(status);
	CREATE INDEX IF NOT EXISTS idx_profiles_url ON profiles(url);
	CREATE INDEX IF NOT EXISTS idx_actions_timestamp ON actions_log(timestamp);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveProfile saves a profile to the database
func (s *Store) SaveProfile(url, name string) error {
	query := `INSERT OR IGNORE INTO profiles (url, name) VALUES (?, ?)`
	_, err := s.db.Exec(query, url, name)
	return err
}

// HasSentRequest checks if a connection request was already sent
func (s *Store) HasSentRequest(profileURL string) (bool, error) {
	query := `SELECT COUNT(*) FROM connection_requests WHERE profile_url = ?`
	var count int
	err := s.db.QueryRow(query, profileURL).Scan(&count)
	return count > 0, err
}

// SaveConnectionRequest records a connection request
func (s *Store) SaveConnectionRequest(profileURL, note string) error {
	query := `INSERT INTO connection_requests (profile_url, note) VALUES (?, ?)`
	_, err := s.db.Exec(query, profileURL, note)
	return err
}

// SaveMessage records a sent message
func (s *Store) SaveMessage(profileURL, messageText string) error {
	query := `INSERT INTO messages (profile_url, message_text) VALUES (?, ?)`
	_, err := s.db.Exec(query, profileURL, messageText)
	return err
}

// LogAction logs an action to the actions_log table
func (s *Store) LogAction(actionType, targetURL, status, details string) error {
	query := `INSERT INTO actions_log (action_type, target_url, status, details) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, actionType, targetURL, status, details)
	return err
}

// GetPendingProfiles retrieves profiles that haven't been processed yet
func (s *Store) GetPendingProfiles(limit int) ([]Profile, error) {
	query := `SELECT id, url, name, status, discovered_at FROM profiles 
	          WHERE status = 'discovered' LIMIT ?`
	
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.URL, &p.Name, &p.Status, &p.DiscoveredAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}

	return profiles, rows.Err()
}

// SaveSession saves browser session cookies
func (s *Store) SaveSession(cookies string) error {
	query := `INSERT INTO sessions (cookies) VALUES (?)`
	_, err := s.db.Exec(query, cookies)
	return err
}

// LoadSession loads the most recent valid session
func (s *Store) LoadSession() (string, error) {
	query := `SELECT cookies FROM sessions WHERE is_valid = 1 
	          ORDER BY last_used_at DESC LIMIT 1`
	
	var cookies string
	err := s.db.QueryRow(query).Scan(&cookies)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return cookies, err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
