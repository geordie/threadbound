package book

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"threadbound/internal/models"
)

// PDFBuilder handles PDF generation using XeLaTeX
type PDFBuilder struct {
	config *models.BookConfig
}

// NewPDFBuilder creates a new PDF builder
func NewPDFBuilder(config *models.BookConfig) *PDFBuilder {
	return &PDFBuilder{config: config}
}

// BuildPDF converts TeX to PDF using XeLaTeX
func (p *PDFBuilder) BuildPDF(inputFile, outputFile string) error {
	// Check if XeLaTeX is available
	if err := p.checkXeLaTeX(); err != nil {
		return err
	}

	// Check if input file exists
	if _, err := os.Stat(inputFile); err != nil {
		return fmt.Errorf("input file not found: %s", inputFile)
	}

	fmt.Printf("🔨 Building PDF with XeLaTeX...\n")
	fmt.Printf("📄 Input: %s\n", inputFile)
	fmt.Printf("📖 Output: %s\n", outputFile)
	fmt.Printf("📐 Page Size: %s x %s\n", p.config.PageWidth, p.config.PageHeight)

	// Get output directory and base filename
	outputDir := filepath.Dir(outputFile)
	baseFilename := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	// Run XeLaTeX multiple times for TOC and cross-references
	// Pass 1: Generate .aux files
	fmt.Printf("🔄 XeLaTeX pass 1/3...\n")
	if err := p.runXeLaTeX(inputFile, outputDir); err != nil {
		return fmt.Errorf("xelatex pass 1 failed: %w", err)
	}

	// Pass 2: Read .aux and generate TOC
	fmt.Printf("🔄 XeLaTeX pass 2/3...\n")
	if err := p.runXeLaTeX(inputFile, outputDir); err != nil {
		return fmt.Errorf("xelatex pass 2 failed: %w", err)
	}

	// Pass 3: Finalize page numbers in TOC
	fmt.Printf("🔄 XeLaTeX pass 3/3...\n")
	if err := p.runXeLaTeX(inputFile, outputDir); err != nil {
		return fmt.Errorf("xelatex pass 3 failed: %w", err)
	}

	// Move the generated PDF to the desired output location
	generatedPDF := filepath.Join(outputDir, baseFilename+".pdf")
	if generatedPDF != outputFile {
		if err := os.Rename(generatedPDF, outputFile); err != nil {
			return fmt.Errorf("failed to move PDF to output location: %w", err)
		}
	}

	// Check if output file was created
	if _, err := os.Stat(outputFile); err != nil {
		return fmt.Errorf("PDF was not created: %s", outputFile)
	}

	fmt.Printf("✅ PDF generated successfully: %s\n", outputFile)
	return nil
}

// runXeLaTeX executes a single XeLaTeX compilation pass
func (p *PDFBuilder) runXeLaTeX(inputFile, outputDir string) error {
	args := []string{
		"-interaction=nonstopmode",
		"-output-directory=" + outputDir,
		inputFile,
	}

	cmd := exec.Command("xelatex", args...)
	cmd.Dir = "."

	// Capture output
	output, err := cmd.CombinedOutput()

	// XeLaTeX may return an error even on success (warnings treated as errors)
	// Check if PDF was actually created
	baseFilename := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	pdfPath := filepath.Join(outputDir, baseFilename+".pdf")
	pdfExists := false
	if _, statErr := os.Stat(pdfPath); statErr == nil {
		pdfExists = true
	}

	if err != nil && !pdfExists {
		fmt.Printf("❌ XeLaTeX failed with error: %v\n", err)
		fmt.Printf("Output:\n%s\n", string(output))
		return fmt.Errorf("xelatex failed: %w", err)
	}

	if err != nil && pdfExists {
		fmt.Printf("⚠️  XeLaTeX completed with warnings (likely font/emoji issues)\n")
	}

	return nil
}

// checkXeLaTeX verifies that XeLaTeX is installed and available
func (p *PDFBuilder) checkXeLaTeX() error {
	cmd := exec.Command("xelatex", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("xelatex not found - please install XeLaTeX (part of TeX Live or MiKTeX) to generate PDFs")
	}

	// Parse version for informational purposes
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		fmt.Printf("📋 Using %s\n", strings.TrimSpace(lines[0]))
	}

	return nil
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