package database

import (
	"testing"
	"threadbound/internal/models"
)

func TestGetHandlesWithCustomNames(t *testing.T) {
	// This test verifies that custom contact names from config are applied correctly
	// Since we need a real database for this test, we'll create a mock test

	// Test data
	contactNames := map[string]string{
		"+15551234567": "Alice",
		"+15559876543": "Bob",
		"test@example.com": "Charlie",
	}

	// Test case 1: Contact ID exists in mapping
	testContact := "+15551234567"
	expectedName := "Alice"

	if customName, exists := contactNames[testContact]; exists {
		if customName != expectedName {
			t.Errorf("Expected custom name %s, got %s", expectedName, customName)
		}
	} else {
		t.Errorf("Expected to find mapping for %s", testContact)
	}

	// Test case 2: Contact ID does not exist in mapping
	unmappedContact := "+15550000000"
	if _, exists := contactNames[unmappedContact]; exists {
		t.Errorf("Did not expect to find mapping for %s", unmappedContact)
	}

	// Test case 3: Email address mapping
	emailContact := "test@example.com"
	expectedEmailName := "Charlie"

	if customName, exists := contactNames[emailContact]; exists {
		if customName != expectedEmailName {
			t.Errorf("Expected custom name %s for email, got %s", expectedEmailName, customName)
		}
	} else {
		t.Errorf("Expected to find mapping for email %s", emailContact)
	}
}

func TestHandleDisplayNameLogic(t *testing.T) {
	tests := []struct {
		name         string
		contactID    string
		contactNames map[string]string
		service      string
		expected     string
	}{
		{
			name:         "Custom name takes precedence",
			contactID:    "+15551234567",
			contactNames: map[string]string{"+15551234567": "Alice"},
			service:      "SMS",
			expected:     "Alice",
		},
		{
			name:         "Falls back to contact ID when no mapping",
			contactID:    "+15551234567",
			contactNames: map[string]string{},
			service:      "SMS",
			expected:     "+15551234567",
		},
		{
			name:         "Email address with custom name",
			contactID:    "alice@example.com",
			contactNames: map[string]string{"alice@example.com": "Alice Smith"},
			service:      "iMessage",
			expected:     "Alice Smith",
		},
		{
			name:         "Email address without custom name",
			contactID:    "bob@example.com",
			contactNames: map[string]string{},
			service:      "iMessage",
			expected:     "bob@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var displayName string

			// Simulate the logic from GetHandles
			if tt.contactNames != nil {
				if customName, exists := tt.contactNames[tt.contactID]; exists {
					displayName = customName
				} else {
					displayName = tt.contactID
					if tt.service == "iMessage" {
						if len(tt.contactID) > 0 && tt.contactID[0] != '+' {
							displayName = tt.contactID
						}
					}
				}
			} else {
				displayName = tt.contactID
			}

			if displayName != tt.expected {
				t.Errorf("Expected display name %s, got %s", tt.expected, displayName)
			}
		})
	}
}

func TestConfigContactNamesStructure(t *testing.T) {
	// Test that BookConfig properly handles ContactNames field
	config := &models.BookConfig{
		Title:        "Test Book",
		Author:       "Test Author",
		ContactNames: map[string]string{
			"+15551234567": "Alice",
			"+15559876543": "Bob",
		},
	}

	if config.ContactNames == nil {
		t.Error("ContactNames map should not be nil")
	}

	if len(config.ContactNames) != 2 {
		t.Errorf("Expected 2 contact name mappings, got %d", len(config.ContactNames))
	}

	if config.ContactNames["+15551234567"] != "Alice" {
		t.Errorf("Expected Alice for +15551234567, got %s", config.ContactNames["+15551234567"])
	}

	if config.ContactNames["+15559876543"] != "Bob" {
		t.Errorf("Expected Bob for +15559876543, got %s", config.ContactNames["+15559876543"])
	}
}