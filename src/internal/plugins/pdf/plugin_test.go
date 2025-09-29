package pdf

import (
	"testing"

	"threadbound/internal/models"
)

func TestPDFPlugin(t *testing.T) {
	plugin := NewPDFPlugin()

	// Test plugin properties
	if plugin.ID() != "pdf" {
		t.Errorf("Expected ID 'pdf', got '%s'", plugin.ID())
	}
	if plugin.Name() != "PDF Book" {
		t.Errorf("Expected name 'PDF Book', got '%s'", plugin.Name())
	}
	if plugin.FileExtension() != "pdf" {
		t.Errorf("Expected extension 'pdf', got '%s'", plugin.FileExtension())
	}

	// Test capabilities
	caps := plugin.GetCapabilities()
	if !caps.SupportsImages {
		t.Error("PDF plugin should support images")
	}
	if !caps.SupportsReactions {
		t.Error("PDF plugin should support reactions")
	}
	if !caps.RequiresTemplates {
		t.Error("PDF plugin should require templates")
	}

	// Test required templates
	templates := plugin.GetRequiredTemplates()
	expectedTemplates := []string{
		"book.tex",
		"sent-message.tex",
		"received-message.tex",
		"title-page.tex",
		"copyright-page.tex",
		"page-structure.tex",
		"yaml-header.yml",
		"image-attachment.tex",
		"image-placeholder.tex",
		"attachment.tex",
	}

	if len(templates) != len(expectedTemplates) {
		t.Errorf("Expected %d templates, got %d", len(expectedTemplates), len(templates))
	}

	for i, expected := range expectedTemplates {
		if i >= len(templates) || templates[i] != expected {
			t.Errorf("Expected template '%s' at index %d", expected, i)
		}
	}
}

func TestPDFPluginValidateConfig(t *testing.T) {
	plugin := NewPDFPlugin()

	// Test valid config
	config := &models.BookConfig{
		Title:       "Test Book",
		TemplateDir: "templates",
		PageWidth:   "8.5in",
		PageHeight:  "11in",
	}

	err := plugin.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}

	// Test config without template dir
	configNoTemplate := &models.BookConfig{
		Title: "Test Book",
	}

	err = plugin.ValidateConfig(configNoTemplate)
	if err == nil {
		t.Error("Expected error for config without template directory")
	}

	// Test config gets default page dimensions
	configNoDimensions := &models.BookConfig{
		Title:       "Test Book",
		TemplateDir: "templates",
	}

	err = plugin.ValidateConfig(configNoDimensions)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if configNoDimensions.PageWidth != "5.5in" {
		t.Errorf("Expected default page width '5.5in', got '%s'", configNoDimensions.PageWidth)
	}
	if configNoDimensions.PageHeight != "8.5in" {
		t.Errorf("Expected default page height '8.5in', got '%s'", configNoDimensions.PageHeight)
	}
}

func TestPDFPluginGenerate(t *testing.T) {
	plugin := NewPDFPlugin()

	// Create a minimal context for testing
	// Note: This test will fail without proper templates and database
	// We'll test that the validation catches missing template directory
	config := &models.BookConfig{
		Title:        "Test Book",
		DatabasePath: "nonexistent.db",
		TemplateDir:  "", // Empty template dir should cause validation error
	}

	// Test that validation catches the empty template directory
	err := plugin.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for empty template directory")
	}

	// Don't test actual generation as it requires templates and database
	// The integration test in the main test already covers that
}