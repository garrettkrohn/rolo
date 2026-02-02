package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetConfigPath returns the path to the rolo config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "rolo", "rolo.txt"), nil
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
