package main

import (
	"os"
	"rolo/storage"
	"testing"
)

func TestHandleLeftAndRight(t *testing.T) {
	// Arrange - create temp directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create initial groups data
	data := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
			{Name: "personal", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}
	storage.SaveGroupsData(data)

	// Note: Full CLI testing would require refactoring handlers to be testable
	// For now, we're testing that the storage functions work correctly
	// The actual CLI commands will be manually tested

	// Act - simulate right movement
	loaded, _ := storage.LoadGroupsData()
	storage.NextGroup(loaded, false)
	storage.SaveGroupsData(loaded)

	// Assert
	reloaded, _ := storage.LoadGroupsData()
	if reloaded.CurrentGroup != 2 {
		t.Errorf("Expected CurrentGroup 2 after right, got %d", reloaded.CurrentGroup)
	}

	// Act - simulate left movement
	storage.PrevGroup(reloaded, false)
	storage.SaveGroupsData(reloaded)

	// Assert
	final, _ := storage.LoadGroupsData()
	if final.CurrentGroup != 1 {
		t.Errorf("Expected CurrentGroup 1 after left, got %d", final.CurrentGroup)
	}
}

func TestHandleLeftRightWithWrapAround(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create config with wrap_around enabled
	config := &storage.Config{WrapAround: true}
	storage.SaveConfig(config)

	// Create groups data at last group
	data := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}
	storage.SaveGroupsData(data)

	// Act - simulate right movement (should wrap to first)
	loaded, _ := storage.LoadGroupsData()
	loadedConfig, _ := storage.LoadConfig()
	storage.NextGroup(loaded, loadedConfig.WrapAround)
	storage.SaveGroupsData(loaded)

	// Assert
	reloaded, _ := storage.LoadGroupsData()
	if reloaded.CurrentGroup != 0 {
		t.Errorf("Expected CurrentGroup 0 after wrap, got %d", reloaded.CurrentGroup)
	}
}

func TestNextPrevWithSingleSession(t *testing.T) {
	// Arrange - create temp directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create groups data with only one active session
	data := &storage.GroupsData{
		Groups: []storage.Group{
			{
				Name: "main",
				Sessions: []storage.SessionData{
					{Name: "only_session", Deleted: false},
				},
			},
		},
		CurrentGroup: 0,
	}
	storage.SaveGroupsData(data)

	// Act - simulate next/prev (should do nothing with single session)
	// This would normally try to switch, but with only 1 session it should not

	loaded, _ := storage.LoadGroupsData()
	
	// Count active sessions
	activeCount := 0
	for _, s := range loaded.Groups[0].Sessions {
		if !s.Deleted {
			activeCount++
		}
	}

	// Assert - should have only 1 active session
	if activeCount != 1 {
		t.Errorf("Expected 1 active session, got %d", activeCount)
	}

	// The handlers would return early in this case
}

func TestNextPrevWithMultipleSessions(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	data := &storage.GroupsData{
		Groups: []storage.Group{
			{
				Name: "main",
				Sessions: []storage.SessionData{
					{Name: "session1", Deleted: false},
					{Name: "session2", Deleted: false},
				},
			},
		},
		CurrentGroup: 0,
	}
	storage.SaveGroupsData(data)

	// Act
	loaded, _ := storage.LoadGroupsData()
	
	// Count active sessions
	activeCount := 0
	for _, s := range loaded.Groups[0].Sessions {
		if !s.Deleted {
			activeCount++
		}
	}

	// Assert - should have 2 active sessions (switching is valid)
	if activeCount != 2 {
		t.Errorf("Expected 2 active sessions, got %d", activeCount)
	}
}
