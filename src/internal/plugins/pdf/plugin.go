package pdf

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"threadbound/internal/models"
	"threadbound/internal/texgen"
	"threadbound/internal/output"
)

// PDFPlugin implements the OutputPlugin interface for PDF generation via XeLaTeX
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

// Generate creates a PDF by first generating TeX then converting with XeLaTeX
func (p *PDFPlugin) Generate(ctx *output.GenerationContext) ([]byte, error) {
	// First generate TeX content
	texContent, err := p.generateTeX(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TeX: %w", err)
	}

	// Write TeX to temporary file
	tempTexPath := "temp_book.tex"
	if err := writeToFile(tempTexPath, texContent); err != nil {
		return nil, fmt.Errorf("failed to write temporary TeX: %w", err)
	}
	defer removeFile(tempTexPath)

	// Convert TeX to PDF using XeLaTeX
	pdfData, err := p.convertToPDF(tempTexPath, ctx.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to PDF: %w", err)
	}

	return pdfData, nil
}

// generateTeX creates the TeX content using the existing generator
func (p *PDFPlugin) generateTeX(ctx *output.GenerationContext) ([]byte, error) {
	// Create a temporary database connection for the TeX generator
	// This is a temporary solution until we refactor the URL processor
	db, err := sql.Open("sqlite3", ctx.Config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create TeX generator
	generator := texgen.New(ctx.Config, db)

	// Convert reactions to the format expected by the generator
	reactions := make(map[string][]models.Reaction)
	for guid, reactionList := range ctx.Reactions {
		reactions[guid] = reactionList
	}

	// Generate the TeX content
	texContent := generator.GenerateBook(ctx.Messages, ctx.Handles, reactions)

	return []byte(texContent), nil
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

// convertToPDF converts the TeX file to PDF using XeLaTeX
func (p *PDFPlugin) convertToPDF(texPath string, config *models.BookConfig) ([]byte, error) {
	// Check if XeLaTeX is available
	if err := checkXeLaTeX(); err != nil {
		return nil, err
	}

	// Generate temporary PDF file
	tempPDFPath := strings.TrimSuffix(texPath, ".tex") + ".pdf"
	defer removeFile(tempPDFPath)

	// Get the directory for temporary files
	outputDir := filepath.Dir(texPath)

	// Run XeLaTeX multiple times for TOC and cross-references
	for i := 1; i <= 3; i++ {
		if err := runXeLaTeX(texPath, outputDir); err != nil {
			return nil, fmt.Errorf("xelatex pass %d failed: %w", i, err)
		}
	}

	// Read the generated PDF file
	pdfData, err := readFile(tempPDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	return pdfData, nil
}