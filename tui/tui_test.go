package tui

import (
	"rolo/storage"
	"testing"
)

func TestGetAdjacentGroupNames(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
			{Name: "personal", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act
	prev, current, next := getAdjacentGroupNames(groupsData)

	// Assert
	if prev != "main" {
		t.Errorf("Expected prev 'main', got '%s'", prev)
	}
	if current != "work" {
		t.Errorf("Expected current 'work', got '%s'", current)
	}
	if next != "personal" {
		t.Errorf("Expected next 'personal', got '%s'", next)
	}
}

func TestGetAdjacentGroupNamesFirstGroup(t *testing.T) {
	// Arrange - at first group
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	prev, current, next := getAdjacentGroupNames(groupsData)

	// Assert
	if prev != "" {
		t.Errorf("Expected prev '', got '%s'", prev)
	}
	if current != "main" {
		t.Errorf("Expected current 'main', got '%s'", current)
	}
	if next != "work" {
		t.Errorf("Expected next 'work', got '%s'", next)
	}
}

func TestGetAdjacentGroupNamesLastGroup(t *testing.T) {
	// Arrange - at last group
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act
	prev, current, next := getAdjacentGroupNames(groupsData)

	// Assert
	if prev != "main" {
		t.Errorf("Expected prev 'main', got '%s'", prev)
	}
	if current != "work" {
		t.Errorf("Expected current 'work', got '%s'", current)
	}
	if next != "" {
		t.Errorf("Expected next '', got '%s'", next)
	}
}

func TestGetAdjacentGroupNamesSingleGroup(t *testing.T) {
	// Arrange - only one group
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	prev, current, next := getAdjacentGroupNames(groupsData)

	// Assert
	if prev != "" {
		t.Errorf("Expected prev '', got '%s'", prev)
	}
	if current != "main" {
		t.Errorf("Expected current 'main', got '%s'", current)
	}
	if next != "" {
		t.Errorf("Expected next '', got '%s'", next)
	}
}

func TestGroupSwitchingLogic(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{{Name: "s1", Deleted: false}}},
			{Name: "work", Sessions: []storage.SessionData{{Name: "s2", Deleted: false}}},
			{Name: "personal", Sessions: []storage.SessionData{{Name: "s3", Deleted: false}}},
		},
		CurrentGroup: 0,
	}

	// Act - simulate switching to next group
	storage.NextGroup(groupsData, false)

	// Assert
	if groupsData.CurrentGroup != 1 {
		t.Errorf("Expected current group 1, got %d", groupsData.CurrentGroup)
	}

	// Act - simulate switching to previous group
	storage.PrevGroup(groupsData, false)

	// Assert
	if groupsData.CurrentGroup != 0 {
		t.Errorf("Expected current group 0 after prev, got %d", groupsData.CurrentGroup)
	}
}

func TestCreateGroupAddsToEnd(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := storage.CreateGroup(groupsData, "personal")

	// Assert
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if len(groupsData.Groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groupsData.Groups))
	}

	if groupsData.Groups[2].Name != "personal" {
		t.Errorf("Expected new group 'personal', got '%s'", groupsData.Groups[2].Name)
	}
}

func TestCreateGroupRejectsEmpty(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := storage.CreateGroup(groupsData, "")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty name, got nil")
	}
}

func TestCreateGroupRejectsDuplicate(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := storage.CreateGroup(groupsData, "work")

	// Assert
	if err == nil {
		t.Fatal("Expected error for duplicate name, got nil")
	}
}

func TestRenameGroupChangesName(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "old_name", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act
	err := storage.RenameGroup(groupsData, 1, "new_name")

	// Assert
	if err != nil {
		t.Fatalf("RenameGroup failed: %v", err)
	}

	if groupsData.Groups[1].Name != "new_name" {
		t.Errorf("Expected 'new_name', got '%s'", groupsData.Groups[1].Name)
	}
}

func TestRenameGroupRejectsEmpty(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := storage.RenameGroup(groupsData, 0, "")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty name, got nil")
	}
}

func TestRenameGroupRejectsDuplicate(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act - try to rename to existing name
	err := storage.RenameGroup(groupsData, 1, "main")

	// Assert
	if err == nil {
		t.Fatal("Expected error for duplicate name, got nil")
	}
}

func TestDeleteGroupRemovesGroup(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
			{Name: "work", Sessions: []storage.SessionData{}},
			{Name: "personal", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 1,
	}

	// Act
	err := storage.DeleteGroup(groupsData, 1)

	// Assert
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	if len(groupsData.Groups) != 2 {
		t.Errorf("Expected 2 groups after delete, got %d", len(groupsData.Groups))
	}

	if groupsData.Groups[1].Name != "personal" {
		t.Errorf("Expected second group 'personal', got '%s'", groupsData.Groups[1].Name)
	}
}

func TestDeleteGroupRejectsLastGroup(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{Name: "main", Sessions: []storage.SessionData{}},
		},
		CurrentGroup: 0,
	}

	// Act
	err := storage.DeleteGroup(groupsData, 0)

	// Assert
	if err == nil {
		t.Fatal("Expected error when deleting last group, got nil")
	}
}

func TestMoveSessionToNextGroupInTUI(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{
				Name: "main",
				Sessions: []storage.SessionData{
					{Name: "session1", Deleted: false},
					{Name: "session2", Deleted: false},
				},
			},
			{
				Name:     "work",
				Sessions: []storage.SessionData{},
			},
		},
		CurrentGroup: 0,
	}

	// Act
	err := storage.MoveSessionToNextGroup(groupsData, 0, false)

	// Assert
	if err != nil {
		t.Fatalf("MoveSessionToNextGroup failed: %v", err)
	}

	if len(groupsData.Groups[0].Sessions) != 1 {
		t.Errorf("Expected 1 session in main, got %d", len(groupsData.Groups[0].Sessions))
	}

	if len(groupsData.Groups[1].Sessions) != 1 {
		t.Errorf("Expected 1 session in work, got %d", len(groupsData.Groups[1].Sessions))
	}

	if groupsData.Groups[1].Sessions[0].Name != "session1" {
		t.Error("Session not moved correctly")
	}
}

func TestMoveSessionToPrevGroupInTUI(t *testing.T) {
	// Arrange
	groupsData := &storage.GroupsData{
		Groups: []storage.Group{
			{
				Name:     "main",
				Sessions: []storage.SessionData{},
			},
			{
				Name: "work",
				Sessions: []storage.SessionData{
					{Name: "session1", Deleted: false},
				},
			},
		},
		CurrentGroup: 1,
	}

	// Act
	err := storage.MoveSessionToPrevGroup(groupsData, 0, false)

	// Assert
	if err != nil {
		t.Fatalf("MoveSessionToPrevGroup failed: %v", err)
	}

	if len(groupsData.Groups[1].Sessions) != 0 {
		t.Errorf("Expected 0 sessions in work, got %d", len(groupsData.Groups[1].Sessions))
	}

	if len(groupsData.Groups[0].Sessions) != 1 {
		t.Errorf("Expected 1 session in main, got %d", len(groupsData.Groups[0].Sessions))
	}
}
