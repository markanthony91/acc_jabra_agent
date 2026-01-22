package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	dbDir := "./data"
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		os.MkdirAll(dbDir, 0755)
	}

	dbPath := filepath.Join(dbDir, "jabra_telemetry.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS battery_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level INTEGER,
		status TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS hardware_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT,
		description TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);`
	_, err := s.db.Exec(query)
	return err
}

func (s *Store) GetSetting(key, defaultValue string) string {
	var value string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return defaultValue
	}
	return value
}

func (s *Store) SaveSetting(key, value string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}

func (s *Store) GetLogs(limit int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query("SELECT event_type, description, timestamp FROM hardware_events ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var etype, desc, ts string
		rows.Scan(&etype, &desc, &ts)
		logs = append(logs, map[string]interface{}{
			"type": etype, "description": desc, "timestamp": ts,
		})
	}
	return logs, nil
}

func (s *Store) LogBattery(level int, status string) {
	_, err := s.db.Exec("INSERT INTO battery_history (level, status) VALUES (?, ?)", level, status)
	if err != nil {
		log.Printf("[DB] Erro ao logar bateria: %v", err)
	}
}

func (s *Store) LogEvent(eventType, desc string) {
	_, err := s.db.Exec("INSERT INTO hardware_events (event_type, description) VALUES (?, ?)", eventType, desc)
	if err != nil {
		log.Printf("[DB] Erro ao logar evento: %v", err)
	}
}

func (s *Store) GetBatteryHistory(limit int) ([]map[string]interface{}, error) {
	rows, err := s.db.Query("SELECT level, status, timestamp FROM battery_history ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var level int
		var status, ts string
		rows.Scan(&level, &status, &ts)
		history = append(history, map[string]interface{}{
			"level": level, "status": status, "timestamp": ts,
		})
	}
	return history, nil
}
