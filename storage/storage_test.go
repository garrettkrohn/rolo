package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadGroupsData(t *testing.T) {
	// Arrange - create temp directory for test
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	testData := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
					{Name: "session2", Deleted: true},
				},
			},
			{
				Name: "work",
				Sessions: []SessionData{
					{Name: "session3", Deleted: false},
				},
			},
		},
		CurrentGroup: 0,
	}

	// Act - save groups data
	err := SaveGroupsData(testData)

	// Assert - should save without error
	if err != nil {
		t.Fatalf("SaveGroupsData failed: %v", err)
	}

	// Act - load groups data
	loaded, err := LoadGroupsData()

	// Assert - should load successfully
	if err != nil {
		t.Fatalf("LoadGroupsData failed: %v", err)
	}

	// Assert - data should match
	if len(loaded.Groups) != len(testData.Groups) {
		t.Errorf("Expected %d groups, got %d", len(testData.Groups), len(loaded.Groups))
	}

	if loaded.CurrentGroup != testData.CurrentGroup {
		t.Errorf("Expected CurrentGroup %d, got %d", testData.CurrentGroup, loaded.CurrentGroup)
	}

	// Assert - first group should match
	if loaded.Groups[0].Name != "main" {
		t.Errorf("Expected first group name 'main', got '%s'", loaded.Groups[0].Name)
	}

	if len(loaded.Groups[0].Sessions) != 2 {
		t.Errorf("Expected 2 sessions in first group, got %d", len(loaded.Groups[0].Sessions))
	}
}

func TestLoadGroupsDataWithEmptyGroupsList(t *testing.T) {
	// Arrange - create temp directory with empty groups
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	emptyData := &GroupsData{
		Groups:       []Group{},
		CurrentGroup: 0,
	}

	// Act - save empty groups data
	err := SaveGroupsData(emptyData)
	if err != nil {
		t.Fatalf("SaveGroupsData failed: %v", err)
	}

	// Act - load groups data
	loaded, err := LoadGroupsData()

	// Assert - should create default "main" group
	if err != nil {
		t.Fatalf("LoadGroupsData failed: %v", err)
	}

	if len(loaded.Groups) != 1 {
		t.Errorf("Expected 1 default group, got %d", len(loaded.Groups))
	}

	if loaded.Groups[0].Name != DefaultGroupName {
		t.Errorf("Expected default group name '%s', got '%s'", DefaultGroupName, loaded.Groups[0].Name)
	}

	if loaded.CurrentGroup != 0 {
		t.Errorf("Expected CurrentGroup 0, got %d", loaded.CurrentGroup)
	}
}

func TestLoadGroupsDataValidatesCurrentGroupIndex(t *testing.T) {
	// Arrange - create temp directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Act - save data with valid current_group
	validData := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
		},
		CurrentGroup: 1,
	}
	err := SaveGroupsData(validData)
	if err != nil {
		t.Fatalf("SaveGroupsData failed: %v", err)
	}

	// Act - load data
	loaded, err := LoadGroupsData()

	// Assert - should load successfully
	if err != nil {
		t.Fatalf("LoadGroupsData failed: %v", err)
	}

	if loaded.CurrentGroup != 1 {
		t.Errorf("Expected CurrentGroup 1, got %d", loaded.CurrentGroup)
	}
}

func TestLoadGroupsDataRejectsInvalidCurrentGroupIndex(t *testing.T) {
	// Arrange - create temp directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Manually write invalid JSON file
	configPath, _ := GetGroupsJSONPath()
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Act - write data with invalid current_group (out of bounds)
	invalidJSON := `{"groups":[{"name":"main","sessions":[]}],"current_group":5}`
	os.WriteFile(configPath, []byte(invalidJSON), 0644)

	// Act - load data
	_, err := LoadGroupsData()

	// Assert - should return error
	if err == nil {
		t.Fatal("Expected error for invalid current_group index, got nil")
	}
}

func TestGroupsDataRoundTrip(t *testing.T) {
	// Arrange - create temp directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	originalData := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
					{Name: "session2", Deleted: true},
				},
			},
			{
				Name: "APICAL-123",
				Sessions: []SessionData{
					{Name: "backend", Deleted: false},
					{Name: "frontend", Deleted: false},
					{Name: "test", Deleted: true},
				},
			},
		},
		CurrentGroup: 1,
	}

	// Act - save and load
	err := SaveGroupsData(originalData)
	if err != nil {
		t.Fatalf("SaveGroupsData failed: %v", err)
	}

	loaded, err := LoadGroupsData()
	if err != nil {
		t.Fatalf("LoadGroupsData failed: %v", err)
	}

	// Assert - all fields should be preserved
	if len(loaded.Groups) != len(originalData.Groups) {
		t.Errorf("Group count mismatch: expected %d, got %d", len(originalData.Groups), len(loaded.Groups))
	}

	if loaded.CurrentGroup != originalData.CurrentGroup {
		t.Errorf("CurrentGroup mismatch: expected %d, got %d", originalData.CurrentGroup, loaded.CurrentGroup)
	}

	// Check first group
	if loaded.Groups[0].Name != "main" {
		t.Errorf("First group name mismatch: expected 'main', got '%s'", loaded.Groups[0].Name)
	}

	if len(loaded.Groups[0].Sessions) != 2 {
		t.Errorf("First group session count: expected 2, got %d", len(loaded.Groups[0].Sessions))
	}

	// Check session details
	if loaded.Groups[0].Sessions[0].Name != "session1" {
		t.Errorf("Session name mismatch")
	}

	if loaded.Groups[0].Sessions[0].Deleted != false {
		t.Errorf("Session deleted state mismatch")
	}

	if loaded.Groups[0].Sessions[1].Deleted != true {
		t.Errorf("Second session should be deleted")
	}

	// Check second group
	if loaded.Groups[1].Name != "APICAL-123" {
		t.Errorf("Second group name mismatch: expected 'APICAL-123', got '%s'", loaded.Groups[1].Name)
	}

	if len(loaded.Groups[1].Sessions) != 3 {
		t.Errorf("Second group session count: expected 3, got %d", len(loaded.Groups[1].Sessions))
	}
}

func TestMigrateToGroupsFormat(t *testing.T) {
	// Arrange - create temp directory with old rolo.json
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create old format data
	oldSessions := []SessionData{
		{Name: "session1", Deleted: false},
		{Name: "session2", Deleted: true},
		{Name: "session3", Deleted: false},
	}
	err := SaveSessionsData(oldSessions)
	if err != nil {
		t.Fatalf("Failed to save old format: %v", err)
	}

	// Act - migrate to new format
	err = MigrateToGroupsFormat()

	// Assert - should migrate successfully
	if err != nil {
		t.Fatalf("MigrateToGroupsFormat failed: %v", err)
	}

	// Assert - groups.json should exist
	groupsPath, _ := GetGroupsJSONPath()
	if _, err := os.Stat(groupsPath); os.IsNotExist(err) {
		t.Fatal("groups.json was not created")
	}

	// Assert - load and verify migrated data
	loaded, err := LoadGroupsData()
	if err != nil {
		t.Fatalf("Failed to load migrated data: %v", err)
	}

	if len(loaded.Groups) != 1 {
		t.Errorf("Expected 1 group (main), got %d", len(loaded.Groups))
	}

	if loaded.Groups[0].Name != DefaultGroupName {
		t.Errorf("Expected group name '%s', got '%s'", DefaultGroupName, loaded.Groups[0].Name)
	}

	if len(loaded.Groups[0].Sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(loaded.Groups[0].Sessions))
	}
}

func TestMigratePreservesSessionOrder(t *testing.T) {
	// Arrange - create temp directory with ordered sessions
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	oldSessions := []SessionData{
		{Name: "first", Deleted: false},
		{Name: "second", Deleted: false},
		{Name: "third", Deleted: false},
	}
	SaveSessionsData(oldSessions)

	// Act - migrate
	err := MigrateToGroupsFormat()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Assert - verify order preserved
	loaded, _ := LoadGroupsData()
	sessions := loaded.Groups[0].Sessions

	if sessions[0].Name != "first" {
		t.Errorf("Expected first session 'first', got '%s'", sessions[0].Name)
	}
	if sessions[1].Name != "second" {
		t.Errorf("Expected second session 'second', got '%s'", sessions[1].Name)
	}
	if sessions[2].Name != "third" {
		t.Errorf("Expected third session 'third', got '%s'", sessions[2].Name)
	}
}

func TestMigratePreservesDeletedState(t *testing.T) {
	// Arrange - create sessions with mixed deleted state
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	oldSessions := []SessionData{
		{Name: "active", Deleted: false},
		{Name: "deleted", Deleted: true},
	}
	SaveSessionsData(oldSessions)

	// Act - migrate
	MigrateToGroupsFormat()

	// Assert - verify deleted state preserved
	loaded, _ := LoadGroupsData()
	sessions := loaded.Groups[0].Sessions

	if sessions[0].Deleted != false {
		t.Error("First session should not be deleted")
	}
	if sessions[1].Deleted != true {
		t.Error("Second session should be deleted")
	}
}

func TestMigrateHandlesEmptyRoloJson(t *testing.T) {
	// Arrange - create empty old format
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	SaveSessionsData([]SessionData{})

	// Act - migrate
	err := MigrateToGroupsFormat()

	// Assert - should create empty main group
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	loaded, _ := LoadGroupsData()
	if len(loaded.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(loaded.Groups))
	}
	if len(loaded.Groups[0].Sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(loaded.Groups[0].Sessions))
	}
}

func TestMigrateDoesNotOverwriteExistingGroupsJson(t *testing.T) {
	// Arrange - create both old and new formats
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create old format
	oldSessions := []SessionData{
		{Name: "old_session", Deleted: false},
	}
	SaveSessionsData(oldSessions)

	// Create new format
	existingGroups := &GroupsData{
		Groups: []Group{
			{
				Name: "existing",
				Sessions: []SessionData{
					{Name: "existing_session", Deleted: false},
				},
			},
		},
		CurrentGroup: 0,
	}
	SaveGroupsData(existingGroups)

	// Act - attempt migration
	MigrateToGroupsFormat()

	// Assert - should not overwrite (skips silently)
	// Existing data should be preserved
	loaded, _ := LoadGroupsData()
	if loaded.Groups[0].Name != "existing" {
		t.Error("Migration overwrote existing groups.json")
	}
}

func TestMigrateSkipsIfGroupsJsonExists(t *testing.T) {
	// Arrange - create groups.json but no rolo.json
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	existingGroups := &GroupsData{
		Groups: []Group{
			{Name: "existing", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}
	SaveGroupsData(existingGroups)

	// Act - attempt migration
	err := MigrateToGroupsFormat()

	// Assert - should skip (not error, just skip)
	if err != nil {
		t.Fatalf("Migration should skip when groups.json exists, got error: %v", err)
	}

	// Assert - existing data unchanged
	loaded, _ := LoadGroupsData()
	if loaded.Groups[0].Name != "existing" {
		t.Error("Existing groups.json was modified")
	}
}

func TestCreateGroup(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := CreateGroup(data, "work")

	// Assert
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if len(data.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(data.Groups))
	}

	if data.Groups[1].Name != "work" {
		t.Errorf("Expected new group name 'work', got '%s'", data.Groups[1].Name)
	}

	if len(data.Groups[1].Sessions) != 0 {
		t.Errorf("Expected new group to have 0 sessions, got %d", len(data.Groups[1].Sessions))
	}
}

func TestCreateGroupAppendsToEnd(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	CreateGroup(data, "personal")

	// Assert - new group should be at the end
	if len(data.Groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(data.Groups))
	}

	if data.Groups[2].Name != "personal" {
		t.Errorf("Expected last group to be 'personal', got '%s'", data.Groups[2].Name)
	}
}

func TestRenameGroup(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "old_name", Sessions: []SessionData{{Name: "session1", Deleted: false}}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := RenameGroup(data, 1, "new_name")

	// Assert
	if err != nil {
		t.Fatalf("RenameGroup failed: %v", err)
	}

	if data.Groups[1].Name != "new_name" {
		t.Errorf("Expected renamed group 'new_name', got '%s'", data.Groups[1].Name)
	}

	// Assert - sessions should be preserved
	if len(data.Groups[1].Sessions) != 1 {
		t.Error("Sessions were lost during rename")
	}
}

func TestCreateGroupRejectsDuplicateNames(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := CreateGroup(data, "work")

	// Assert
	if err == nil {
		t.Fatal("Expected error for duplicate group name, got nil")
	}

	if len(data.Groups) != 2 {
		t.Error("Group was added despite duplicate name")
	}
}

func TestRenameGroupRejectsDuplicateNames(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
			{Name: "personal", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act - try to rename "personal" to "work" (which already exists)
	err := RenameGroup(data, 2, "work")

	// Assert
	if err == nil {
		t.Fatal("Expected error for duplicate group name, got nil")
	}

	if data.Groups[2].Name != "personal" {
		t.Error("Group was renamed despite duplicate name")
	}
}

func TestDeleteGroup(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
			{Name: "personal", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := DeleteGroup(data, 1)

	// Assert
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	if len(data.Groups) != 2 {
		t.Errorf("Expected 2 groups after deletion, got %d", len(data.Groups))
	}

	// Verify "work" was deleted and "personal" shifted down
	if data.Groups[0].Name != "main" {
		t.Error("First group should still be 'main'")
	}

	if data.Groups[1].Name != "personal" {
		t.Error("Second group should now be 'personal'")
	}
}

func TestDeleteGroupAdjustsCurrentGroupWhenDeletingBefore(t *testing.T) {
	// Arrange - current group is index 2
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
			{Name: "personal", Sessions: []SessionData{}},
		},
		CurrentGroup: 2,
	}

	// Act - delete group at index 1 (before current)
	DeleteGroup(data, 1)

	// Assert - current group should shift down
	if data.CurrentGroup != 1 {
		t.Errorf("Expected CurrentGroup to be adjusted to 1, got %d", data.CurrentGroup)
	}
}

func TestDeleteGroupAdjustsCurrentGroupWhenDeletingCurrent(t *testing.T) {
	// Arrange - current group is index 1
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
			{Name: "personal", Sessions: []SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act - delete current group
	DeleteGroup(data, 1)

	// Assert - current group should move to previous (or stay at 0 if already at start)
	if data.CurrentGroup != 0 {
		t.Errorf("Expected CurrentGroup to move to 0, got %d", data.CurrentGroup)
	}
}

func TestDeleteGroupRejectsLastRemainingGroup(t *testing.T) {
	// Arrange - only one group
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := DeleteGroup(data, 0)

	// Assert
	if err == nil {
		t.Fatal("Expected error when deleting last group, got nil")
	}

	if len(data.Groups) != 1 {
		t.Error("Last group was deleted")
	}
}

func TestCreateGroupRejectsEmptyName(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := CreateGroup(data, "")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty group name, got nil")
	}

	if len(data.Groups) != 1 {
		t.Error("Group was added despite empty name")
	}
}

func TestRenameGroupRejectsEmptyName(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := RenameGroup(data, 0, "")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty group name, got nil")
	}

	if data.Groups[0].Name != "main" {
		t.Error("Group name was changed despite empty name")
	}
}

func TestNextGroup(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
			{Name: "personal", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := NextGroup(data, false)

	// Assert
	if err != nil {
		t.Fatalf("NextGroup failed: %v", err)
	}

	if data.CurrentGroup != 1 {
		t.Errorf("Expected CurrentGroup 1, got %d", data.CurrentGroup)
	}
}

func TestNextGroupWrapsWhenEnabled(t *testing.T) {
	// Arrange - at last group
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act - next with wrap around
	err := NextGroup(data, true)

	// Assert
	if err != nil {
		t.Fatalf("NextGroup failed: %v", err)
	}

	if data.CurrentGroup != 0 {
		t.Errorf("Expected to wrap to group 0, got %d", data.CurrentGroup)
	}
}

func TestNextGroupDoesNotWrapWhenDisabled(t *testing.T) {
	// Arrange - at last group
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act - next without wrap around
	err := NextGroup(data, false)

	// Assert - should not move or error
	if err != nil {
		t.Fatalf("NextGroup failed: %v", err)
	}

	if data.CurrentGroup != 1 {
		t.Errorf("Expected to stay at group 1, got %d", data.CurrentGroup)
	}
}

func TestPrevGroup(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
			{Name: "personal", Sessions: []SessionData{}},
		},
		CurrentGroup: 2,
	}

	// Act
	err := PrevGroup(data, false)

	// Assert
	if err != nil {
		t.Fatalf("PrevGroup failed: %v", err)
	}

	if data.CurrentGroup != 1 {
		t.Errorf("Expected CurrentGroup 1, got %d", data.CurrentGroup)
	}
}

func TestPrevGroupWrapsWhenEnabled(t *testing.T) {
	// Arrange - at first group
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act - prev with wrap around
	err := PrevGroup(data, true)

	// Assert
	if err != nil {
		t.Fatalf("PrevGroup failed: %v", err)
	}

	if data.CurrentGroup != 1 {
		t.Errorf("Expected to wrap to group 1, got %d", data.CurrentGroup)
	}
}

func TestPrevGroupDoesNotWrapWhenDisabled(t *testing.T) {
	// Arrange - at first group
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
			{Name: "work", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act - prev without wrap around
	err := PrevGroup(data, false)

	// Assert - should not move or error
	if err != nil {
		t.Fatalf("PrevGroup failed: %v", err)
	}

	if data.CurrentGroup != 0 {
		t.Errorf("Expected to stay at group 0, got %d", data.CurrentGroup)
	}
}

func TestNextGroupWithSingleGroup(t *testing.T) {
	// Arrange - only one group
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := NextGroup(data, true)

	// Assert - should not move
	if err != nil {
		t.Fatalf("NextGroup failed: %v", err)
	}

	if data.CurrentGroup != 0 {
		t.Errorf("Expected to stay at group 0, got %d", data.CurrentGroup)
	}
}

func TestPrevGroupWithSingleGroup(t *testing.T) {
	// Arrange - only one group
	data := &GroupsData{
		Groups: []Group{
			{Name: "main", Sessions: []SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := PrevGroup(data, true)

	// Assert - should not move
	if err != nil {
		t.Fatalf("PrevGroup failed: %v", err)
	}

	if data.CurrentGroup != 0 {
		t.Errorf("Expected to stay at group 0, got %d", data.CurrentGroup)
	}
}

func TestNextGroupWithEmptyGroupsList(t *testing.T) {
	// Arrange - no groups
	data := &GroupsData{
		Groups:       []Group{},
		CurrentGroup: 0,
	}

	// Act
	err := NextGroup(data, false)

	// Assert - should return error
	if err == nil {
		t.Fatal("Expected error for empty groups list, got nil")
	}
}

func TestPrevGroupWithEmptyGroupsList(t *testing.T) {
	// Arrange - no groups
	data := &GroupsData{
		Groups:       []Group{},
		CurrentGroup: 0,
	}

	// Act
	err := PrevGroup(data, false)

	// Assert - should return error
	if err == nil {
		t.Fatal("Expected error for empty groups list, got nil")
	}
}

func TestMoveSessionToNextGroup(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
					{Name: "session2", Deleted: false},
				},
			},
			{
				Name:     "work",
				Sessions: []SessionData{},
			},
		},
		CurrentGroup: 0,
	}

	// Act - move first session to next group
	err := MoveSessionToNextGroup(data, 0, false)

	// Assert
	if err != nil {
		t.Fatalf("MoveSessionToNextGroup failed: %v", err)
	}

	// Assert - session removed from current group
	if len(data.Groups[0].Sessions) != 1 {
		t.Errorf("Expected 1 session in main group, got %d", len(data.Groups[0].Sessions))
	}

	if data.Groups[0].Sessions[0].Name != "session2" {
		t.Error("Wrong session remained in main group")
	}

	// Assert - session added to work group
	if len(data.Groups[1].Sessions) != 1 {
		t.Errorf("Expected 1 session in work group, got %d", len(data.Groups[1].Sessions))
	}

	if data.Groups[1].Sessions[0].Name != "session1" {
		t.Error("Session not added to work group correctly")
	}
}

func TestMoveSessionToPrevGroup(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{
				Name:     "main",
				Sessions: []SessionData{},
			},
			{
				Name: "work",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
					{Name: "session2", Deleted: true},
				},
			},
		},
		CurrentGroup: 1,
	}

	// Act - move first session to previous group
	err := MoveSessionToPrevGroup(data, 0, false)

	// Assert
	if err != nil {
		t.Fatalf("MoveSessionToPrevGroup failed: %v", err)
	}

	// Assert - session removed from work group
	if len(data.Groups[1].Sessions) != 1 {
		t.Errorf("Expected 1 session in work group, got %d", len(data.Groups[1].Sessions))
	}

	// Assert - session added to main group
	if len(data.Groups[0].Sessions) != 1 {
		t.Errorf("Expected 1 session in main group, got %d", len(data.Groups[0].Sessions))
	}

	if data.Groups[0].Sessions[0].Name != "session1" {
		t.Error("Session not moved to main group correctly")
	}
}

func TestMoveSessionPreservesData(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "deleted_session", Deleted: true},
				},
			},
			{
				Name:     "work",
				Sessions: []SessionData{},
			},
		},
		CurrentGroup: 0,
	}

	// Act
	MoveSessionToNextGroup(data, 0, false)

	// Assert - deleted state preserved
	if data.Groups[1].Sessions[0].Deleted != true {
		t.Error("Deleted state not preserved during move")
	}

	if data.Groups[1].Sessions[0].Name != "deleted_session" {
		t.Error("Session name not preserved during move")
	}
}

func TestMoveOnlySessionLeavesGroupEmpty(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "only_session", Deleted: false},
				},
			},
			{
				Name:     "work",
				Sessions: []SessionData{},
			},
		},
		CurrentGroup: 0,
	}

	// Act
	err := MoveSessionToNextGroup(data, 0, false)

	// Assert
	if err != nil {
		t.Fatalf("MoveSessionToNextGroup failed: %v", err)
	}

	// Assert - source group is empty
	if len(data.Groups[0].Sessions) != 0 {
		t.Errorf("Expected 0 sessions in main group, got %d", len(data.Groups[0].Sessions))
	}

	// Assert - target group has the session
	if len(data.Groups[1].Sessions) != 1 {
		t.Errorf("Expected 1 session in work group, got %d", len(data.Groups[1].Sessions))
	}
}

func TestMoveSessionToNextGroupWithWrapAround(t *testing.T) {
	// Arrange - at last group
	data := &GroupsData{
		Groups: []Group{
			{
				Name:     "main",
				Sessions: []SessionData{},
			},
			{
				Name: "work",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
				},
			},
		},
		CurrentGroup: 1,
	}

	// Act - move to next (should wrap to first)
	err := MoveSessionToNextGroup(data, 0, true)

	// Assert
	if err != nil {
		t.Fatalf("MoveSessionToNextGroup failed: %v", err)
	}

	// Assert - session moved to first group (wrapped)
	if len(data.Groups[0].Sessions) != 1 {
		t.Error("Session not moved to first group (wrapped)")
	}

	if len(data.Groups[1].Sessions) != 0 {
		t.Error("Session not removed from last group")
	}
}

func TestMoveSessionToPrevGroupWithWrapAround(t *testing.T) {
	// Arrange - at first group
	data := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
				},
			},
			{
				Name:     "work",
				Sessions: []SessionData{},
			},
		},
		CurrentGroup: 0,
	}

	// Act - move to prev (should wrap to last)
	err := MoveSessionToPrevGroup(data, 0, true)

	// Assert
	if err != nil {
		t.Fatalf("MoveSessionToPrevGroup failed: %v", err)
	}

	// Assert - session moved to last group (wrapped)
	if len(data.Groups[1].Sessions) != 1 {
		t.Error("Session not moved to last group (wrapped)")
	}

	if len(data.Groups[0].Sessions) != 0 {
		t.Error("Session not removed from first group")
	}
}

func TestMoveSessionFromSingleGroupReturnsError(t *testing.T) {
	// Arrange - only one group
	data := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
				},
			},
		},
		CurrentGroup: 0,
	}

	// Act
	err := MoveSessionToNextGroup(data, 0, false)

	// Assert
	if err == nil {
		t.Fatal("Expected error when moving from single group, got nil")
	}
}

func TestMoveSessionWithInvalidIndexReturnsError(t *testing.T) {
	// Arrange
	data := &GroupsData{
		Groups: []Group{
			{
				Name: "main",
				Sessions: []SessionData{
					{Name: "session1", Deleted: false},
				},
			},
			{
				Name:     "work",
				Sessions: []SessionData{},
			},
		},
		CurrentGroup: 0,
	}

	// Act - try to move session at invalid index
	err := MoveSessionToNextGroup(data, 5, false)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid session index, got nil")
	}
}
