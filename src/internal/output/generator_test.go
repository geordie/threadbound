package output

import (
	"testing"

	"threadbound/internal/models"
)

func TestGenerator(t *testing.T) {
	// Create a test registry and register a mock plugin
	registry := NewRegistry()
	mockPlugin := &MockPlugin{
		id:          "test",
		name:        "Test Plugin",
		description: "Test plugin",
		extension:   "test",
	}
	err := registry.Register(mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register mock plugin: %v", err)
	}

	// Create generator with test registry
	generator := &Generator{registry: registry}

	// Create test context
	config := &models.BookConfig{
		Title:      "Test Book",
		OutputPath: "test_output.md",
	}
	ctx := &GenerationContext{
		Messages:  []models.Message{},
		Handles:   make(map[int]models.Handle),
		Reactions: make(map[string][]models.Reaction),
		Config:    config,
		Stats:     &models.BookStats{},
	}

	// Test generation
	data, filename, err := generator.Generate("test", ctx)
	if err != nil {
		t.Fatalf("Failed to generate output: %v", err)
	}

	if string(data) != "test" {
		t.Errorf("Expected output 'test', got '%s'", string(data))
	}

	expectedFilename := "test_output.test"
	if filename != expectedFilename {
		t.Errorf("Expected filename '%s', got '%s'", expectedFilename, filename)
	}

	// Test non-existent plugin
	_, _, err = generator.Generate("nonexistent", ctx)
	if err == nil {
		t.Error("Expected error when generating with non-existent plugin")
	}
}

func TestGenerateFilename(t *testing.T) {
	generator := &Generator{}

	tests := []struct {
		basePath  string
		extension string
		expected  string
	}{
		{"output.md", "pdf", "output.pdf"},
		{"output.test", "test", "output.test"},
		{"output", "html", "output.html"},
		{"path/to/output.md", "pdf", "path/to/output.pdf"},
	}

	for _, test := range tests {
		result := generator.generateFilename(test.basePath, test.extension)
		if result != test.expected {
			t.Errorf("generateFilename(%s, %s) = %s, expected %s",
				test.basePath, test.extension, result, test.expected)
		}
	}
}

func TestCreateContext(t *testing.T) {
	messages := []models.Message{{ID: 1, GUID: "test"}}
	handles := map[int]models.Handle{1: {ID: 1, DisplayName: "Test"}}
	reactions := map[string][]models.Reaction{"test": {}}
	config := &models.BookConfig{Title: "Test"}
	stats := &models.BookStats{TotalMessages: 1}

	ctx := CreateContext(messages, handles, reactions, config, stats)

	if len(ctx.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(ctx.Messages))
	}
	if len(ctx.Handles) != 1 {
		t.Errorf("Expected 1 handle, got %d", len(ctx.Handles))
	}
	if ctx.Config.Title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", ctx.Config.Title)
	}
	if ctx.Stats.TotalMessages != 1 {
		t.Errorf("Expected 1 total message, got %d", ctx.Stats.TotalMessages)
	}
}

func TestValidatePluginExists(t *testing.T) {
	// This test uses the global registry, so we need to register a plugin
	mockPlugin := &MockPlugin{id: "testvalidate", name: "Test", description: "Test", extension: "test"}
	err := Register(mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register test plugin: %v", err)
	}

	// Test existing plugin
	err = ValidatePluginExists("testvalidate")
	if err != nil {
		t.Errorf("Expected no error for existing plugin, got: %v", err)
	}

	// Test non-existing plugin
	err = ValidatePluginExists("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
}