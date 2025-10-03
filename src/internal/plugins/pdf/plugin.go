package pdf

import (
	"fmt"

	"threadbound/internal/latex"
	"threadbound/internal/models"
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

	// Generate temporary PDF path
	tempPDFPath := "temp_book.pdf"
	defer removeFile(tempPDFPath)

	// Convert TeX to PDF using XeLaTeX builder
	latexBuilder := latex.NewBuilder(ctx.Config)
	if err := latexBuilder.BuildPDF(tempTexPath, tempPDFPath); err != nil {
		return nil, fmt.Errorf("failed to convert to PDF: %w", err)
	}

	// Read the generated PDF file
	pdfData, err := readFile(tempPDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	return pdfData, nil
}

// generateTeX creates the TeX content using the TeX plugin
func (p *PDFPlugin) generateTeX(ctx *output.GenerationContext) ([]byte, error) {
	// Get the TeX plugin from the registry
	registry := output.GetGlobalRegistry()
	texPlugin, err := registry.Get("tex")
	if err != nil {
		return nil, fmt.Errorf("failed to get TeX plugin: %w", err)
	}

	// Generate the TeX content using the TeX plugin
	texContent, err := texPlugin.Generate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TeX: %w", err)
	}

	return texContent, nil
}

// ValidateConfig validates the PDF plugin configuration
func (p *PDFPlugin) ValidateConfig(config *models.BookConfig) error {
	// Call base validation first
	if err := p.BasePlugin.ValidateConfig(config); err != nil {
		return err
	}

	// Template directory is optional (templates are embedded in binary)
	// It's only needed if user wants custom templates

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