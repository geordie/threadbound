package text

import (
	"strings"
	"testing"
	"time"

	"threadbound/internal/models"
	"threadbound/internal/output"
)

func TestTextPlugin(t *testing.T) {
	plugin := NewTextPlugin()

	// Test plugin properties
	if plugin.ID() != "txt" {
		t.Errorf("Expected ID 'txt', got '%s'", plugin.ID())
	}
	if plugin.Name() != "Plain Text" {
		t.Errorf("Expected name 'Plain Text', got '%s'", plugin.Name())
	}
	if plugin.FileExtension() != "txt" {
		t.Errorf("Expected extension 'txt', got '%s'", plugin.FileExtension())
	}

	// Test capabilities
	caps := plugin.GetCapabilities()
	if caps.SupportsImages {
		t.Error("Text plugin should not support images")
	}
	if !caps.SupportsAttachments {
		t.Error("Text plugin should support attachments")
	}
	if !caps.SupportsReactions {
		t.Error("Text plugin should support reactions")
	}
	if !caps.RequiresTemplates {
		t.Error("Text plugin should require templates")
	}
	if caps.SupportsPagination {
		t.Error("Text plugin should not support pagination")
	}

	// Test required templates
	templates := plugin.GetRequiredTemplates()
	expectedTemplates := []string{"header.txt", "date-separator.txt", "message.txt"}
	if len(templates) != len(expectedTemplates) {
		t.Errorf("Expected %d templates, got %d", len(expectedTemplates), len(templates))
	}
	for i, expected := range expectedTemplates {
		if templates[i] != expected {
			t.Errorf("Expected template '%s', got '%s'", expected, templates[i])
		}
	}
}

func TestTextPluginValidateConfig(t *testing.T) {
	plugin := NewTextPlugin()

	// Test valid config
	config := &models.BookConfig{
		Title:  "Test Book",
		Author: "Test Author",
	}

	err := plugin.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}

	// Test empty title gets default
	configEmptyTitle := &models.BookConfig{}
	err = plugin.ValidateConfig(configEmptyTitle)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if configEmptyTitle.Title != "Untitled Book" {
		t.Errorf("Expected default title 'Untitled Book', got '%s'", configEmptyTitle.Title)
	}
}

func TestTextPluginGenerate(t *testing.T) {
	plugin := NewTextPlugin()

	// Create test data
	testTime := time.Date(2023, 9, 15, 10, 30, 0, 0, time.UTC)
	messages := []models.Message{
		{
			ID:            1,
			GUID:          "msg1",
			Text:          stringPtr("Hello world!"),
			IsFromMe:      true,
			FormattedDate: testTime,
		},
		{
			ID:            2,
			GUID:          "msg2",
			Text:          stringPtr("Hi there!"),
			IsFromMe:      false,
			HandleID:      intPtr(1),
			FormattedDate: testTime.Add(time.Minute),
		},
	}

	handles := map[int]models.Handle{
		1: {ID: 1, DisplayName: "Test User"},
	}

	reactions := map[string][]models.Reaction{
		"msg1": {
			{SenderName: "Test User", ReactionEmoji: "👍"},
		},
	}

	config := &models.BookConfig{
		Title:       "Test Text Book",
		Author:      "Test Author",
		TemplateDir: "",
	}

	stats := &models.BookStats{
		TotalMessages: 2,
		TextMessages:  2,
		TotalContacts: 1,
		StartDate:     testTime,
		EndDate:       testTime.Add(time.Minute),
	}

	ctx := &output.GenerationContext{
		Messages:      messages,
		Handles:       handles,
		Reactions:     reactions,
		Config:        config,
		Stats:         stats,
		URLThumbnails: make(map[string]*output.URLThumbnail),
	}

	// Generate text
	data, err := plugin.Generate(ctx)
	if err != nil {
		t.Fatalf("Failed to generate text: %v", err)
	}

	text := string(data)

	// Test that text contains expected elements
	if !strings.Contains(text, "=== Test Text Book ===") {
		t.Error("Text should contain book title")
	}
	if !strings.Contains(text, "Test Author") {
		t.Error("Text should contain author name")
	}
	if !strings.Contains(text, "Messages: 2") {
		t.Error("Text should contain message count")
	}
	if !strings.Contains(text, "Text Messages: 2") {
		t.Error("Text should contain text message count")
	}
	if !strings.Contains(text, "Contacts: 1") {
		t.Error("Text should contain contact count")
	}
	if !strings.Contains(text, "Hello world!") {
		t.Error("Text should contain first message")
	}
	if !strings.Contains(text, "Hi there!") {
		t.Error("Text should contain second message")
	}
	if !strings.Contains(text, "Me:") {
		t.Error("Text should contain 'Me' as sender for sent messages")
	}
	if !strings.Contains(text, "Test User:") {
		t.Error("Text should contain sender name for received messages")
	}
	if !strings.Contains(text, "👍") {
		t.Error("Text should contain reaction emoji")
	}
	if !strings.Contains(text, "---") {
		t.Error("Text should contain date separator")
	}

	// Test that timestamps are present
	if !strings.Contains(text, "[") || !strings.Contains(text, "]") {
		t.Error("Text should contain timestamps in brackets")
	}
}

func TestTextPluginGenerateWithAttachments(t *testing.T) {
	plugin := NewTextPlugin()

	// Create test data with attachment
	testTime := time.Date(2023, 9, 15, 10, 30, 0, 0, time.UTC)
	messages := []models.Message{
		{
			ID:            1,
			GUID:          "msg1",
			Text:          stringPtr("Check out this file"),
			IsFromMe:      true,
			FormattedDate: testTime,
			Attachments: []models.Attachment{
				{Filename: stringPtr("test.pdf")},
				{Filename: stringPtr("image.jpg")},
			},
		},
	}

	handles := map[int]models.Handle{}
	reactions := map[string][]models.Reaction{}
	config := &models.BookConfig{
		Title:       "Test",
		TemplateDir: "",
	}
	stats := &models.BookStats{}

	ctx := &output.GenerationContext{
		Messages:      messages,
		Handles:       handles,
		Reactions:     reactions,
		Config:        config,
		Stats:         stats,
		URLThumbnails: make(map[string]*output.URLThumbnail),
	}

	// Generate text
	data, err := plugin.Generate(ctx)
	if err != nil {
		t.Fatalf("Failed to generate text: %v", err)
	}

	text := string(data)

	// Test that attachments are listed
	if !strings.Contains(text, "Attachments:") {
		t.Error("Text should contain 'Attachments:' label")
	}
	if !strings.Contains(text, "test.pdf") {
		t.Error("Text should contain first attachment filename")
	}
	if !strings.Contains(text, "image.jpg") {
		t.Error("Text should contain second attachment filename")
	}
}

func TestTextPluginGroupMessagesByDate(t *testing.T) {
	plugin := NewTextPlugin()

	// Create test data with messages on different dates
	testTime1 := time.Date(2023, 9, 15, 10, 30, 0, 0, time.UTC)
	testTime2 := time.Date(2023, 9, 16, 14, 45, 0, 0, time.UTC)

	messages := []models.Message{
		{
			ID:            1,
			Text:          stringPtr("Message on day 1"),
			FormattedDate: testTime1,
		},
		{
			ID:            2,
			Text:          stringPtr("Another message on day 1"),
			FormattedDate: testTime1.Add(time.Hour),
		},
		{
			ID:            3,
			Text:          stringPtr("Message on day 2"),
			FormattedDate: testTime2,
		},
	}

	grouped := plugin.groupMessagesByDate(messages)

	if len(grouped) != 2 {
		t.Errorf("Expected 2 date groups, got %d", len(grouped))
	}

	date1Key := testTime1.Format("2006-01-02")
	date2Key := testTime2.Format("2006-01-02")

	if msgs, exists := grouped[date1Key]; !exists || len(msgs) != 2 {
		t.Errorf("Expected 2 messages for date %s, got %d", date1Key, len(msgs))
	}

	if msgs, exists := grouped[date2Key]; !exists || len(msgs) != 1 {
		t.Errorf("Expected 1 message for date %s, got %d", date2Key, len(msgs))
	}
}

func TestTextPluginSkipsEmptyMessages(t *testing.T) {
	plugin := NewTextPlugin()

	// Create test data with empty messages
	testTime := time.Date(2023, 9, 15, 10, 30, 0, 0, time.UTC)
	messages := []models.Message{
		{
			ID:            1,
			Text:          stringPtr("Valid message"),
			FormattedDate: testTime,
		},
		{
			ID:            2,
			Text:          nil, // Empty message
			FormattedDate: testTime,
		},
		{
			ID:            3,
			Text:          stringPtr("   "), // Whitespace only
			FormattedDate: testTime,
		},
	}

	grouped := plugin.groupMessagesByDate(messages)

	dateKey := testTime.Format("2006-01-02")
	if msgs, exists := grouped[dateKey]; !exists || len(msgs) != 1 {
		t.Errorf("Expected 1 message after filtering, got %d", len(msgs))
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
