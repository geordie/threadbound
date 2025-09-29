package html

import (
	"strings"
	"testing"
	"time"

	"threadbound/internal/models"
	"threadbound/internal/output"
)

func TestHTMLPlugin(t *testing.T) {
	plugin := NewHTMLPlugin()

	// Test plugin properties
	if plugin.ID() != "html" {
		t.Errorf("Expected ID 'html', got '%s'", plugin.ID())
	}
	if plugin.Name() != "HTML Book" {
		t.Errorf("Expected name 'HTML Book', got '%s'", plugin.Name())
	}
	if plugin.FileExtension() != "html" {
		t.Errorf("Expected extension 'html', got '%s'", plugin.FileExtension())
	}

	// Test capabilities
	caps := plugin.GetCapabilities()
	if !caps.SupportsImages {
		t.Error("HTML plugin should support images")
	}
	if !caps.SupportsReactions {
		t.Error("HTML plugin should support reactions")
	}
	if caps.RequiresTemplates {
		t.Error("HTML plugin should not require templates")
	}
	if caps.SupportsPagination {
		t.Error("HTML plugin should not support pagination")
	}

	// Test required templates (should be empty)
	templates := plugin.GetRequiredTemplates()
	if len(templates) != 0 {
		t.Errorf("Expected 0 templates, got %d", len(templates))
	}
}

func TestHTMLPluginValidateConfig(t *testing.T) {
	plugin := NewHTMLPlugin()

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

func TestHTMLPluginGenerate(t *testing.T) {
	plugin := NewHTMLPlugin()

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
		Title:  "Test HTML Book",
		Author: "Test Author",
	}

	stats := &models.BookStats{
		TotalMessages: 2,
		TextMessages:  2,
		TotalContacts: 1,
	}

	ctx := &output.GenerationContext{
		Messages:  messages,
		Handles:   handles,
		Reactions: reactions,
		Config:    config,
		Stats:     stats,
	}

	// Generate HTML
	data, err := plugin.Generate(ctx)
	if err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	html := string(data)

	// Test that HTML contains expected elements
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE declaration")
	}
	if !strings.Contains(html, "Test HTML Book") {
		t.Error("HTML should contain book title")
	}
	if !strings.Contains(html, "Test Author") {
		t.Error("HTML should contain author name")
	}
	if !strings.Contains(html, "Hello world!") {
		t.Error("HTML should contain first message")
	}
	if !strings.Contains(html, "Hi there!") {
		t.Error("HTML should contain second message")
	}
	if !strings.Contains(html, "Test User") {
		t.Error("HTML should contain sender name")
	}
	if !strings.Contains(html, "from-me") {
		t.Error("HTML should contain 'from-me' class for sent messages")
	}
	if !strings.Contains(html, "👍") {
		t.Error("HTML should contain reaction emoji")
	}

	// Test CSS is embedded
	if !strings.Contains(html, "<style>") {
		t.Error("HTML should contain embedded CSS")
	}
	if !strings.Contains(html, "message-bubble") {
		t.Error("HTML should contain message bubble styles")
	}
}

func TestHTMLPluginPrepareTemplateData(t *testing.T) {
	plugin := NewHTMLPlugin()

	// Create test data
	testTime := time.Date(2023, 9, 15, 10, 30, 0, 0, time.UTC)
	messages := []models.Message{
		{
			ID:            1,
			GUID:          "msg1",
			Text:          stringPtr("Test message"),
			IsFromMe:      true,
			FormattedDate: testTime,
		},
	}

	handles := map[int]models.Handle{}
	reactions := map[string][]models.Reaction{}
	config := &models.BookConfig{Title: "Test"}
	stats := &models.BookStats{}

	ctx := &output.GenerationContext{
		Messages:  messages,
		Handles:   handles,
		Reactions: reactions,
		Config:    config,
		Stats:     stats,
	}

	templateData := plugin.prepareTemplateData(ctx)

	if templateData.Title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", templateData.Title)
	}

	if len(templateData.MessagesByDate) == 0 {
		t.Error("Expected messages to be grouped by date")
	}

	dateKey := testTime.Format("2006-01-02")
	if _, exists := templateData.MessagesByDate[dateKey]; !exists {
		t.Errorf("Expected messages for date %s", dateKey)
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}