package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SessionData represents a session with its deleted state
type SessionData struct {
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
}

// GetConfigPath returns the path to the rolo config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "rolo", "rolo.txt"), nil
}

// GetConfigJSONPath returns the path to the rolo JSON config file
func GetConfigJSONPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "rolo", "rolo.json"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	return nil
}

// LoadSessionsData reads the session list with deleted state from the JSON config file
// Falls back to old txt format if JSON doesn't exist
func LoadSessionsData() ([]SessionData, error) {
	jsonPath, err := GetConfigJSONPath()
	if err != nil {
		return nil, err
	}
	
	// Try to load JSON format first
	if _, err := os.Stat(jsonPath); err == nil {
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read JSON config file: %w", err)
		}
		
		var sessions []SessionData
		if err := json.Unmarshal(data, &sessions); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
		}
		
		return sessions, nil
	}
	
	// Fall back to old txt format
	sessions, err := LoadSessions()
	if err != nil {
		return nil, err
	}
	
	// Convert to SessionData format
	sessionData := make([]SessionData, len(sessions))
	for i, name := range sessions {
		sessionData[i] = SessionData{Name: name, Deleted: false}
	}
	
	return sessionData, nil
}

// LoadSessions reads the session list from the config file
// Returns an empty slice if the file doesn't exist
func LoadSessions() ([]string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	
	// If file doesn't exist, return empty slice
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Split by newlines and filter out empty lines
	content := strings.TrimSpace(string(data))
	if content == "" {
		return []string{}, nil
	}
	
	lines := strings.Split(content, "\n")
	sessions := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	
	return sessions, nil
}

// SaveSessionsData writes the session list with deleted state to the JSON config file
func SaveSessionsData(sessions []SessionData) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	
	jsonPath, err := GetConfigJSONPath()
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}
	
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON config file: %w", err)
	}
	
	return nil
}

// SaveSessions writes the session list to the config file
func SaveSessions(sessions []string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	
	content := strings.Join(sessions, "\n")
	if content != "" {
		content += "\n" // Add trailing newline
	}
	
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}
