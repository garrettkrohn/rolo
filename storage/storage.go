package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultGroupName is the name of the default group
	DefaultGroupName = "main"
)

// SessionData represents a session with its deleted state and context flag
type SessionData struct {
	Name      string `json:"name"`
	Deleted   bool   `json:"deleted"`
	InContext bool   `json:"in_context,omitempty"`
}

// Group represents a named collection of sessions
type Group struct {
	Name     string        `json:"name"`
	Sessions []SessionData `json:"sessions"`
}

// GroupsData represents all groups and the current active group
type GroupsData struct {
	Groups       []Group `json:"groups"`
	CurrentGroup int     `json:"current_group"`
}

// Config represents the rolo configuration settings
type Config struct {
	WrapAround  bool `json:"wrap_around"`
	ShowDeleted bool `json:"show_deleted"`
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

// GetConfigSettingsPath returns the path to the rolo settings config file
func GetConfigSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "rolo", "config.json"), nil
}

// GetGroupsJSONPath returns the path to the groups JSON config file
func GetGroupsJSONPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "rolo", "groups.json"), nil
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

// LoadConfig reads the configuration settings from the config file
// Returns default config if the file doesn't exist
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigSettingsPath()
	if err != nil {
		return nil, err
	}
	
	// If file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			WrapAround:  false,
			ShowDeleted: true,
		}, nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return &config, nil
}

// SaveConfig writes the configuration settings to the config file
func SaveConfig(config *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigSettingsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadGroupsData reads groups data from groups.json
// Returns default "main" group if file doesn't exist or groups list is empty
func LoadGroupsData() (*GroupsData, error) {
	groupsPath, err := GetGroupsJSONPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return default main group
	if _, err := os.Stat(groupsPath); os.IsNotExist(err) {
		return createDefaultGroupsData(), nil
	}

	data, err := os.ReadFile(groupsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read groups file: %w", err)
	}

	var groupsData GroupsData
	if err := json.Unmarshal(data, &groupsData); err != nil {
		return nil, fmt.Errorf("failed to parse groups file: %w", err)
	}

	// If groups list is empty, return default
	if len(groupsData.Groups) == 0 {
		return createDefaultGroupsData(), nil
	}

	// Validate current_group index
	if groupsData.CurrentGroup < 0 || groupsData.CurrentGroup >= len(groupsData.Groups) {
		return nil, fmt.Errorf("invalid current_group index %d (have %d groups)", groupsData.CurrentGroup, len(groupsData.Groups))
	}

	return &groupsData, nil
}

// SaveGroupsData writes groups data to groups.json
func SaveGroupsData(data *GroupsData) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	groupsPath, err := GetGroupsJSONPath()
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal groups data: %w", err)
	}

	if err := os.WriteFile(groupsPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write groups file: %w", err)
	}

	return nil
}

// createDefaultGroupsData creates a default GroupsData with a "main" group
func createDefaultGroupsData() *GroupsData {
	return &GroupsData{
		Groups: []Group{
			{
				Name:     DefaultGroupName,
				Sessions: []SessionData{},
			},
		},
		CurrentGroup: 0,
	}
}

// MigrateToGroupsFormat migrates old rolo.json format to new groups.json format
// Skips migration if groups.json already exists
func MigrateToGroupsFormat() error {
	groupsPath, err := GetGroupsJSONPath()
	if err != nil {
		return err
	}

	// Skip if groups.json already exists
	if _, err := os.Stat(groupsPath); err == nil {
		return nil
	}

	// Load old format sessions
	oldSessions, err := LoadSessionsData()
	if err != nil {
		// If old format doesn't exist, nothing to migrate
		return nil
	}

	// Create new format with all sessions in "main" group
	groupsData := &GroupsData{
		Groups: []Group{
			{
				Name:     DefaultGroupName,
				Sessions: oldSessions,
			},
		},
		CurrentGroup: 0,
	}

	// Save to new format
	return SaveGroupsData(groupsData)
}

// CreateGroup adds a new group with the given name
func CreateGroup(data *GroupsData, name string) error {
	if err := validateGroupName(name); err != nil {
		return err
	}

	if groupNameExists(data, name) {
		return fmt.Errorf("group with name '%s' already exists", name)
	}

	newGroup := Group{
		Name:     name,
		Sessions: []SessionData{},
	}

	data.Groups = append(data.Groups, newGroup)
	return nil
}

// RenameGroup renames the group at the given index
func RenameGroup(data *GroupsData, index int, newName string) error {
	if err := validateGroupName(newName); err != nil {
		return err
	}

	if index < 0 || index >= len(data.Groups) {
		return fmt.Errorf("invalid group index %d", index)
	}

	// Allow renaming to same name (no-op)
	if data.Groups[index].Name == newName {
		return nil
	}

	if groupNameExists(data, newName) {
		return fmt.Errorf("group with name '%s' already exists", newName)
	}

	data.Groups[index].Name = newName
	return nil
}

// DeleteGroup removes the group at the given index
func DeleteGroup(data *GroupsData, index int) error {
	if index < 0 || index >= len(data.Groups) {
		return fmt.Errorf("invalid group index %d", index)
	}

	if len(data.Groups) == 1 {
		return fmt.Errorf("cannot delete the last remaining group")
	}

	// Remove group from slice
	data.Groups = append(data.Groups[:index], data.Groups[index+1:]...)

	// Adjust current_group index
	if data.CurrentGroup == index {
		// Deleting current group - move to previous group if possible
		if data.CurrentGroup > 0 {
			data.CurrentGroup--
		}
		// Otherwise stay at 0 (which is now a different group)
	} else if data.CurrentGroup > index {
		// Deleting a group before current - adjust index down
		data.CurrentGroup--
	}

	return nil
}

// validateGroupName checks if a group name is valid
func validateGroupName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	return nil
}

// groupNameExists checks if a group with the given name already exists
func groupNameExists(data *GroupsData, name string) bool {
	for _, group := range data.Groups {
		if group.Name == name {
			return true
		}
	}
	return false
}

// NextGroup moves to the next group
func NextGroup(data *GroupsData, wrapAround bool) error {
	if len(data.Groups) == 0 {
		return fmt.Errorf("no groups available")
	}

	if len(data.Groups) == 1 {
		// Only one group, can't move
		return nil
	}

	// Check if at last group
	if data.CurrentGroup == len(data.Groups)-1 {
		if wrapAround {
			data.CurrentGroup = 0
		}
		// Otherwise stay at current position
	} else {
		data.CurrentGroup++
	}

	return nil
}

// PrevGroup moves to the previous group
func PrevGroup(data *GroupsData, wrapAround bool) error {
	if len(data.Groups) == 0 {
		return fmt.Errorf("no groups available")
	}

	if len(data.Groups) == 1 {
		// Only one group, can't move
		return nil
	}

	// Check if at first group
	if data.CurrentGroup == 0 {
		if wrapAround {
			data.CurrentGroup = len(data.Groups) - 1
		}
		// Otherwise stay at current position
	} else {
		data.CurrentGroup--
	}

	return nil
}

// MoveSessionToNextGroup moves a session from current group to the next group
func MoveSessionToNextGroup(data *GroupsData, sessionIndex int, wrapAround bool) error {
	if len(data.Groups) <= 1 {
		return fmt.Errorf("need at least 2 groups to move sessions")
	}

	currentGroup := &data.Groups[data.CurrentGroup]

	if sessionIndex < 0 || sessionIndex >= len(currentGroup.Sessions) {
		return fmt.Errorf("invalid session index %d", sessionIndex)
	}

	// Determine target group index
	targetIndex := data.CurrentGroup + 1
	if targetIndex >= len(data.Groups) {
		if wrapAround {
			targetIndex = 0
		} else {
			return nil // No wrap, no move
		}
	}

	// Move session
	session := currentGroup.Sessions[sessionIndex]
	targetGroup := &data.Groups[targetIndex]

	// Remove from current group
	currentGroup.Sessions = append(currentGroup.Sessions[:sessionIndex], currentGroup.Sessions[sessionIndex+1:]...)

	// Add to target group
	targetGroup.Sessions = append(targetGroup.Sessions, session)

	return nil
}

// MoveSessionToPrevGroup moves a session from current group to the previous group
func MoveSessionToPrevGroup(data *GroupsData, sessionIndex int, wrapAround bool) error {
	if len(data.Groups) <= 1 {
		return fmt.Errorf("need at least 2 groups to move sessions")
	}

	currentGroup := &data.Groups[data.CurrentGroup]

	if sessionIndex < 0 || sessionIndex >= len(currentGroup.Sessions) {
		return fmt.Errorf("invalid session index %d", sessionIndex)
	}

	// Determine target group index
	targetIndex := data.CurrentGroup - 1
	if targetIndex < 0 {
		if wrapAround {
			targetIndex = len(data.Groups) - 1
		} else {
			return nil // No wrap, no move
		}
	}

	// Move session
	session := currentGroup.Sessions[sessionIndex]
	targetGroup := &data.Groups[targetIndex]

	// Remove from current group
	currentGroup.Sessions = append(currentGroup.Sessions[:sessionIndex], currentGroup.Sessions[sessionIndex+1:]...)

	// Add to target group
	targetGroup.Sessions = append(targetGroup.Sessions, session)

	return nil
}
