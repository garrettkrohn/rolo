package main

import (
	"os"
	"rolo/storage"
	"testing"
)

func TestFullWorkflowWithGroups(t *testing.T) {
	// Arrange - create temp directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Act & Assert - Create initial groups data
	groupsData := &storage.GroupsData{
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

	// Save initial state
	err := storage.SaveGroupsData(groupsData)
	if err != nil {
		t.Fatalf("Failed to save initial groups: %v", err)
	}

	// Create a new group
	err = storage.CreateGroup(groupsData, "work")
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Move session to new group
	err = storage.MoveSessionToNextGroup(groupsData, 0, false)
	if err != nil {
		t.Fatalf("Failed to move session: %v", err)
	}

	// Switch to new group
	err = storage.NextGroup(groupsData, false)
	if err != nil {
		t.Fatalf("Failed to switch group: %v", err)
	}

	// Save and reload
	err = storage.SaveGroupsData(groupsData)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	loaded, err := storage.LoadGroupsData()
	if err != nil {
		t.Fatalf("Failed to load groups: %v", err)
	}

	// Assert - verify state persisted
	if len(loaded.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(loaded.Groups))
	}

	if loaded.CurrentGroup != 1 {
		t.Errorf("Expected current group 1, got %d", loaded.CurrentGroup)
	}

	if len(loaded.Groups[0].Sessions) != 1 {
		t.Errorf("Expected 1 session in main, got %d", len(loaded.Groups[0].Sessions))
	}

	if len(loaded.Groups[1].Sessions) != 1 {
		t.Errorf("Expected 1 session in work, got %d", len(loaded.Groups[1].Sessions))
	}

	if loaded.Groups[1].Sessions[0].Name != "session1" {
		t.Error("Session not moved correctly")
	}
}

func TestMigrationFromOldFormat(t *testing.T) {
	// Arrange - create temp directory with old format
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create old format data
	oldSessions := []storage.SessionData{
		{Name: "old_session1", Deleted: false},
		{Name: "old_session2", Deleted: true},
	}
	storage.SaveSessionsData(oldSessions)

	// Act - run migration
	err := storage.MigrateToGroupsFormat()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Assert - verify migration successful
	groupsData, err := storage.LoadGroupsData()
	if err != nil {
		t.Fatalf("Failed to load migrated data: %v", err)
	}

	if len(groupsData.Groups) != 1 {
		t.Errorf("Expected 1 group after migration, got %d", len(groupsData.Groups))
	}

	if groupsData.Groups[0].Name != storage.DefaultGroupName {
		t.Errorf("Expected default group name '%s', got '%s'", storage.DefaultGroupName, groupsData.Groups[0].Name)
	}

	if len(groupsData.Groups[0].Sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(groupsData.Groups[0].Sessions))
	}

	if groupsData.Groups[0].Sessions[1].Deleted != true {
		t.Error("Deleted state not preserved during migration")
	}
}

func TestGroupDeletionWorkflow(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{{Name: "s1", Deleted: false}}},
			{Name: "work", Sessions: []storage.SessionData{{Name: "s2", Deleted: false}}},
			{Name: "personal", Sessions: []storage.SessionData{{Name: "s3", Deleted: false}}},
		},
		CurrentGroup: 1,
	}

	// Act - delete current group
	err := storage.DeleteGroup(groupsData, 1)
	if err != nil {
		t.Fatalf("Failed to delete group: %v", err)
	}

	// Assert - verify deletion and adjustment
	if len(groupsData.Groups) != 2 {
		t.Errorf("Expected 2 groups after deletion, got %d", len(groupsData.Groups))
	}

	if groupsData.CurrentGroup != 0 {
		t.Errorf("Expected current group adjusted to 0, got %d", groupsData.CurrentGroup)
	}

	// Verify correct groups remain
	if groupsData.Groups[0].Name != "main" || groupsData.Groups[1].Name != "personal" {
		t.Error("Wrong groups remained after deletion")
	}
}

func TestWrapAroundNavigation(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act - next with wrap around (at last group)
	storage.NextGroup(groupsData, true)

	// Assert - should wrap to first
	if groupsData.CurrentGroup != 0 {
		t.Errorf("Expected wrap to group 0, got %d", groupsData.CurrentGroup)
	}

	// Act - prev with wrap around (at first group)
	storage.PrevGroup(groupsData, true)

	// Assert - should wrap to last
	if groupsData.CurrentGroup != 1 {
		t.Errorf("Expected wrap to group 1, got %d", groupsData.CurrentGroup)
	}

	// Act - next without wrap around (at last group)
	groupsData.CurrentGroup = 1
	storage.NextGroup(groupsData, false)

	// Assert - should stay at last
	if groupsData.CurrentGroup != 1 {
		t.Errorf("Expected to stay at group 1, got %d", groupsData.CurrentGroup)
	}
}

func TestCurrentGroupPersistsImmediately(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{
				Name: "main",
				Sessions: []storage.SessionData{
					{Name: "rolo", Deleted: false},
					{Name: "rx", Deleted: false},
				},
			},
			{
				Name: "rolo_group",
				Sessions: []storage.SessionData{
					{Name: "rolo", Deleted: false},
					{Name: "other", Deleted: true}, // Deleted
				},
			},
		},
		CurrentGroup: 0, // Start at main
	}

	storage.SaveGroupsData(groupsData)

	// Act - Simulate switching to next group (like pressing 'l' in TUI)
	storage.NextGroup(groupsData, false)
	// In the TUI, we now auto-save after group switch
	storage.SaveGroupsData(groupsData)

	// Reload to simulate CLI command reading the file
	reloaded, err := storage.LoadGroupsData()
	if err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}

	// Assert - current group should be persisted
	if reloaded.CurrentGroup != 1 {
		t.Errorf("Expected current_group 1 (rolo_group), got %d", reloaded.CurrentGroup)
	}

	// Count active sessions in the current group
	activeCount := 0
	currentSessions := reloaded.Groups[reloaded.CurrentGroup].Sessions
	for _, s := range currentSessions {
		if !s.Deleted {
			activeCount++
		}
	}

	// Assert - rolo_group should have only 1 active session
	if activeCount != 1 {
		t.Errorf("Expected 1 active session in rolo_group, got %d", activeCount)
	}

	// This means "rolo next" would correctly do nothing (only 1 session)
}

func TestSaveSessionsInTUI(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create initial empty group
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{
				Name:     "main",
				Sessions: []storage.SessionData{},
			},
		},
		CurrentGroup: 0,
	}
	storage.SaveGroupsData(groupsData)

	// Act - Simulate adding sessions (like pressing 'p' in TUI)
	groupsData.Groups[0].Sessions = []storage.SessionData{
		{Name: "session1", Deleted: false},
		{Name: "session2", Deleted: false},
	}

	// Simulate saving (like pressing 's' or Enter in TUI)
	storage.SaveGroupsData(groupsData)

	// Reload to verify it was saved
	reloaded, err := storage.LoadGroupsData()
	if err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}

	// Assert
	if len(reloaded.Groups[0].Sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(reloaded.Groups[0].Sessions))
	}

	if reloaded.Groups[0].Sessions[0].Name != "session1" {
		t.Errorf("Expected 'session1', got '%s'", reloaded.Groups[0].Sessions[0].Name)
	}
}
