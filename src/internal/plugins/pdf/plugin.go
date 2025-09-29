package pdf

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"threadbound/internal/markdown"
	"threadbound/internal/models"
	"threadbound/internal/output"
)

// PDFPlugin implements the OutputPlugin interface for PDF generation via Pandoc
type PDFPlugin struct {
	*output.BasePlugin
}

// NewPDFPlugin creates a new PDF plugin instance
func NewPDFPlugin() *PDFPlugin {
	capabilities := output.PluginCapabilities{
		SupportsImages:      true,
		SupportsAttachments: true,
		SupportsReactions:   true,
		SupportsURLPreviews: true,
		RequiresTemplates:   true,
		SupportsPagination:  true,
	}

	base := output.NewBasePlugin(
		"pdf",
		"PDF Book",
		"Generate PDF book using Pandoc with LaTeX templates",
		"pdf",
		capabilities,
	)

	return &PDFPlugin{
		BasePlugin: base,
	}
}

// Generate creates a PDF by first generating markdown then converting with Pandoc
func (p *PDFPlugin) Generate(ctx *output.GenerationContext) ([]byte, error) {
	// First generate markdown content
	markdownContent, err := p.generateMarkdown(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate markdown: %w", err)
	}

	// Write markdown to temporary file
	tempMarkdownPath := "temp_book.md"
	if err := writeToFile(tempMarkdownPath, markdownContent); err != nil {
		return nil, fmt.Errorf("failed to write temporary markdown: %w", err)
	}
	defer removeFile(tempMarkdownPath)

	// Convert markdown to PDF using Pandoc
	pdfData, err := p.convertToPDF(tempMarkdownPath, ctx.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to PDF: %w", err)
	}

	return pdfData, nil
}

// generateMarkdown creates the markdown content using the existing generator
func (p *PDFPlugin) generateMarkdown(ctx *output.GenerationContext) ([]byte, error) {
	// Create a temporary database connection for the markdown generator
	// This is a temporary solution until we refactor the URL processor
	db, err := sql.Open("sqlite3", ctx.Config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create markdown generator
	generator := markdown.New(ctx.Config, db)

	// Convert reactions to the format expected by the generator
	reactions := make(map[string][]models.Reaction)
	for guid, reactionList := range ctx.Reactions {
		reactions[guid] = reactionList
	}

	// Generate the markdown content
	markdownContent := generator.GenerateBook(ctx.Messages, ctx.Handles, reactions)

	return []byte(markdownContent), nil
}

// ValidateConfig validates the PDF plugin configuration
func (p *PDFPlugin) ValidateConfig(config *models.BookConfig) error {
	// Call base validation first
	if err := p.BasePlugin.ValidateConfig(config); err != nil {
		return err
	}

	// Check for required template directory
	if config.TemplateDir == "" {
		return fmt.Errorf("template directory is required for PDF generation")
	}

	// Validate page dimensions
	if config.PageWidth == "" {
		config.PageWidth = "5.5in"
	}
	if config.PageHeight == "" {
		config.PageHeight = "8.5in"
	}

	return nil
}

// GetRequiredTemplates returns the list of template files needed for PDF generation
func (p *PDFPlugin) GetRequiredTemplates() []string {
	return []string{
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
}

// convertToPDF converts the markdown file to PDF using Pandoc
func (p *PDFPlugin) convertToPDF(markdownPath string, config *models.BookConfig) ([]byte, error) {
	// Check if Pandoc is available
	if err := checkPandoc(); err != nil {
		return nil, err
	}

	// Prepare template path
	templatePath := filepath.Join(config.TemplateDir, "book.tex")
	if !fileExists(templatePath) {
		return nil, fmt.Errorf("template not found: %s", templatePath)
	}

	// Generate temporary PDF file
	tempPDFPath := strings.TrimSuffix(markdownPath, ".md") + ".pdf"
	defer removeFile(tempPDFPath)

	// Build Pandoc command
	args := []string{
		markdownPath,
		"--from", "markdown",
		"--to", "pdf",
		"--template", templatePath,
		"--pdf-engine", "xelatex",
		"--variable", fmt.Sprintf("geometry:paperwidth=%s", config.PageWidth),
		"--variable", fmt.Sprintf("geometry:paperheight=%s", config.PageHeight),
		"--variable", "geometry:margin=0.5in",
		"--table-of-contents",
		"--number-sections",
		"--standalone",
		"--output", tempPDFPath,
		"--verbose",
	}

	// Execute Pandoc
	if err := runPandoc(args); err != nil {
		return nil, err
	}

	// Read the generated PDF file
	pdfData, err := readFile(tempPDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	return pdfData, nil
}