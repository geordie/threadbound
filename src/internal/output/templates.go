package output

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

// TemplateManager handles loading and executing templates
type TemplateManager struct {
	templateDir string
	templates   map[string]*template.Template
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templateDir string) *TemplateManager {
	return &TemplateManager{
		templateDir: templateDir,
		templates:   make(map[string]*template.Template),
	}
}

// LoadTemplate loads and parses a template file
func (tm *TemplateManager) LoadTemplate(filename string) (*template.Template, error) {
	// Check if template is already loaded
	if tmpl, exists := tm.templates[filename]; exists {
		return tmpl, nil
	}

	// Load template file
	fullPath := filepath.Join(tm.templateDir, filename)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", filename, err)
	}

	// Parse template
	tmpl, err := template.New(filename).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", filename, err)
	}

	// Cache the template
	tm.templates[filename] = tmpl

	return tmpl, nil
}

// ExecuteTemplate executes a template with the given data
func (tm *TemplateManager) ExecuteTemplate(filename string, data interface{}) (string, error) {
	tmpl, err := tm.LoadTemplate(filename)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", filename, err)
	}

	return buf.String(), nil
}

// LoadTemplates loads multiple templates at once
func (tm *TemplateManager) LoadTemplates(filenames []string) error {
	for _, filename := range filenames {
		if _, err := tm.LoadTemplate(filename); err != nil {
			return err
		}
	}
	return nil
}

// MustLoadTemplate loads a template and panics on error (for initialization)
func (tm *TemplateManager) MustLoadTemplate(filename string) *template.Template {
	tmpl, err := tm.LoadTemplate(filename)
	if err != nil {
		panic(err)
	}
	return tmpl
}

// MustExecuteTemplate executes a template and panics on error
func (tm *TemplateManager) MustExecuteTemplate(filename string, data interface{}) string {
	result, err := tm.ExecuteTemplate(filename, data)
	if err != nil {
		panic(err)
	}
	return result
}

// GetTemplate returns a loaded template by name
func (tm *TemplateManager) GetTemplate(filename string) (*template.Template, bool) {
	tmpl, exists := tm.templates[filename]
	return tmpl, exists
}