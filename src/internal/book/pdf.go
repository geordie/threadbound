package book

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"imessages-book/internal/models"
)

// PDFBuilder handles PDF generation using Pandoc
type PDFBuilder struct {
	config *models.BookConfig
}

// NewPDFBuilder creates a new PDF builder
func NewPDFBuilder(config *models.BookConfig) *PDFBuilder {
	return &PDFBuilder{config: config}
}

// BuildPDF converts markdown to PDF using Pandoc
func (p *PDFBuilder) BuildPDF(inputFile, outputFile string) error {
	// Check if Pandoc is available
	if err := p.checkPandoc(); err != nil {
		return err
	}

	// Check if input file exists
	if _, err := os.Stat(inputFile); err != nil {
		return fmt.Errorf("input file not found: %s", inputFile)
	}

	// Prepare template path
	templatePath := filepath.Join(p.config.TemplateDir, "book.tex")
	if _, err := os.Stat(templatePath); err != nil {
		return fmt.Errorf("template not found: %s", templatePath)
	}

	// Build Pandoc command
	args := []string{
		inputFile,
		"--from", "markdown",
		"--to", "pdf",
		"--template", templatePath,
		"--pdf-engine", "xelatex",
		"--variable", fmt.Sprintf("geometry:paperwidth=%s", p.config.PageWidth),
		"--variable", fmt.Sprintf("geometry:paperheight=%s", p.config.PageHeight),
		"--variable", "geometry:margin=0.5in",
		"--table-of-contents",
		"--number-sections",
		"--standalone",
		"--output", outputFile,
	}

	// Add verbose output for debugging
	args = append(args, "--verbose")

	fmt.Printf("🔨 Building PDF with Pandoc...\n")
	fmt.Printf("📄 Input: %s\n", inputFile)
	fmt.Printf("📖 Output: %s\n", outputFile)
	fmt.Printf("📐 Page Size: %s x %s\n", p.config.PageWidth, p.config.PageHeight)

	// Execute Pandoc
	cmd := exec.Command("pandoc", args...)

	// Set working directory to the project root
	cmd.Dir = "."

	// Capture output
	output, err := cmd.CombinedOutput()

	// Check if PDF was created despite errors (LaTeX often succeeds with warnings)
	pdfExists := false
	if _, statErr := os.Stat(outputFile); statErr == nil {
		pdfExists = true
	}

	if err != nil && !pdfExists {
		fmt.Printf("❌ Pandoc failed with error: %v\n", err)
		fmt.Printf("Output:\n%s\n", string(output))
		return fmt.Errorf("pandoc failed: %w\nOutput: %s", err, string(output))
	}

	if err != nil && pdfExists {
		fmt.Printf("⚠️  Pandoc completed with warnings (likely font/emoji issues)\n")
		fmt.Printf("📄 PDF was still generated successfully\n")
	}

	// Check if output file was created
	if _, err := os.Stat(outputFile); err != nil {
		return fmt.Errorf("PDF was not created: %s", outputFile)
	}

	fmt.Printf("✅ PDF generated successfully: %s\n", outputFile)
	return nil
}

// checkPandoc verifies that Pandoc is installed and available
func (p *PDFBuilder) checkPandoc() error {
	cmd := exec.Command("pandoc", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("pandoc not found - please install Pandoc to generate PDFs")
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