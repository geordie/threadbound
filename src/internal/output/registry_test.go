package output

import (
	"testing"

	"threadbound/internal/models"
)

func TestRegistry(t *testing.T) {
	// Create a new registry for testing
	registry := NewRegistry()

	// Test empty registry
	if len(registry.GetIDs()) != 0 {
		t.Errorf("Expected empty registry, got %d plugins", len(registry.GetIDs()))
	}

	// Create a mock plugin
	mockPlugin := &MockPlugin{
		id:          "test",
		name:        "Test Plugin",
		description: "Test plugin for unit tests",
		extension:   "test",
	}

	// Test registration
	err := registry.Register(mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(mockPlugin)
	if err == nil {
		t.Error("Expected error when registering duplicate plugin")
	}

	// Test retrieval
	plugin, err := registry.Get("test")
	if err != nil {
		t.Fatalf("Failed to retrieve plugin: %v", err)
	}
	if plugin.ID() != "test" {
		t.Errorf("Expected plugin ID 'test', got '%s'", plugin.ID())
	}

	// Test non-existent plugin
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when retrieving non-existent plugin")
	}

	// Test exists
	if !registry.Exists("test") {
		t.Error("Expected plugin 'test' to exist")
	}
	if registry.Exists("nonexistent") {
		t.Error("Expected plugin 'nonexistent' to not exist")
	}

	// Test list
	plugins := registry.List()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	// Test format list
	formatList := registry.FormatList()
	if formatList == "" {
		t.Error("Expected non-empty format list")
	}
}

// MockPlugin is a test implementation of OutputPlugin
type MockPlugin struct {
	id          string
	name        string
	description string
	extension   string
}

func (m *MockPlugin) ID() string                           { return m.id }
func (m *MockPlugin) Name() string                         { return m.name }
func (m *MockPlugin) Description() string                  { return m.description }
func (m *MockPlugin) FileExtension() string                { return m.extension }
func (m *MockPlugin) GetCapabilities() PluginCapabilities  { return PluginCapabilities{} }
func (m *MockPlugin) Generate(ctx *GenerationContext) ([]byte, error) { return []byte("test"), nil }
func (m *MockPlugin) ValidateConfig(config *models.BookConfig) error { return nil }
func (m *MockPlugin) GetRequiredTemplates() []string           { return []string{} }