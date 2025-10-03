package book

import (
	"fmt"
	"os"
	"os/exec"

	"threadbound/internal/latex"
	"threadbound/internal/models"
)

// PDFBuilder handles PDF generation using XeLaTeX
type PDFBuilder struct {
	config        *models.BookConfig
	latexBuilder *latex.Builder
}

// NewPDFBuilder creates a new PDF builder
func NewPDFBuilder(config *models.BookConfig) *PDFBuilder {
	return &PDFBuilder{
		config:        config,
		latexBuilder: latex.NewBuilder(config),
	}
}

// BuildPDF converts TeX to PDF using XeLaTeX
func (p *PDFBuilder) BuildPDF(inputFile, outputFile string) error {
	return p.latexBuilder.BuildPDF(inputFile, outputFile)
}

// GetPDFInfo returns information about the generated PDF
func (p *PDFBuilder) GetPDFInfo(pdfPath string) (*models.PDFInfo, error) {
	if _, err := os.Stat(pdfPath); err != nil {
		return nil, fmt.Errorf("PDF file not found: %s", pdfPath)
	}

	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		return nil, err
	}

	info := &models.PDFInfo{
		FilePath:   pdfPath,
		FileSize:   fileInfo.Size(),
		CreatedAt:  fileInfo.ModTime(),
		PageWidth:  p.config.PageWidth,
		PageHeight: p.config.PageHeight,
	}

	return info, nil
}

// PreviewCommand returns the command to open the PDF for preview
func (p *PDFBuilder) PreviewCommand(pdfPath string) string {
	// macOS
	if _, err := exec.LookPath("open"); err == nil {
		return fmt.Sprintf("open %s", pdfPath)
	}

	// Linux
	if _, err := exec.LookPath("xdg-open"); err == nil {
		return fmt.Sprintf("xdg-open %s", pdfPath)
	}

	// Windows (if running under WSL or similar)
	if _, err := exec.LookPath("cmd.exe"); err == nil {
		return fmt.Sprintf("cmd.exe /c start %s", pdfPath)
	}

	return fmt.Sprintf("Please open %s with your PDF viewer", pdfPath)
}