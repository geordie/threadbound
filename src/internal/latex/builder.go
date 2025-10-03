package latex

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"threadbound/internal/models"
)

// Builder handles PDF generation using XeLaTeX
type Builder struct {
	config *models.BookConfig
}

// NewBuilder creates a new XeLaTeX builder
func NewBuilder(config *models.BookConfig) *Builder {
	return &Builder{config: config}
}

// BuildPDF converts TeX to PDF using XeLaTeX
func (b *Builder) BuildPDF(inputFile, outputFile string) error {
	// Check if XeLaTeX is available
	if err := b.checkXeLaTeX(); err != nil {
		return err
	}

	// Check if input file exists
	if _, err := os.Stat(inputFile); err != nil {
		return fmt.Errorf("input file not found: %s", inputFile)
	}

	fmt.Printf("🔨 Building PDF with XeLaTeX...\n")
	fmt.Printf("📄 Input: %s\n", inputFile)
	fmt.Printf("📖 Output: %s\n", outputFile)
	if b.config != nil {
		fmt.Printf("📐 Page Size: %s x %s\n", b.config.PageWidth, b.config.PageHeight)
	}

	// Get output directory and base filename
	outputDir := filepath.Dir(outputFile)
	baseFilename := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	// Clean up XeLaTeX temporary files after completion
	defer b.cleanupXeLaTeXFiles(filepath.Join(outputDir, baseFilename))

	// Run XeLaTeX multiple times for TOC and cross-references
	// Pass 1: Generate .aux files
	fmt.Printf("🔄 XeLaTeX pass 1/3...\n")
	if err := b.runXeLaTeX(inputFile, outputDir); err != nil {
		return fmt.Errorf("xelatex pass 1 failed: %w", err)
	}

	// Pass 2: Read .aux and generate TOC
	fmt.Printf("🔄 XeLaTeX pass 2/3...\n")
	if err := b.runXeLaTeX(inputFile, outputDir); err != nil {
		return fmt.Errorf("xelatex pass 2 failed: %w", err)
	}

	// Pass 3: Finalize page numbers in TOC
	fmt.Printf("🔄 XeLaTeX pass 3/3...\n")
	if err := b.runXeLaTeX(inputFile, outputDir); err != nil {
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
func (b *Builder) runXeLaTeX(inputFile, outputDir string) error {
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
func (b *Builder) checkXeLaTeX() error {
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

// cleanupXeLaTeXFiles removes temporary files created by XeLaTeX
func (b *Builder) cleanupXeLaTeXFiles(baseFilename string) {
	// List of common XeLaTeX temporary file extensions
	tempExtensions := []string{
		".aux",         // Auxiliary file for cross-references
		".log",         // Log file
		".toc",         // Table of contents
		".out",         // PDF outline/bookmarks
		".lof",         // List of figures
		".lot",         // List of tables
		".fls",         // File list
		".fdb_latexmk", // Latexmk database
	}

	// Remove each temporary file
	for _, ext := range tempExtensions {
		tempFile := baseFilename + ext
		os.Remove(tempFile) // Ignore errors for cleanup
	}
}
