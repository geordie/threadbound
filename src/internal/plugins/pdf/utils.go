package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeToFile writes data to a file
func writeToFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// readFile reads data from a file
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// removeFile removes a file if it exists
func removeFile(path string) {
	os.Remove(path) // Ignore errors for cleanup
}

// cleanupXeLaTeXFiles removes temporary files created by XeLaTeX
func cleanupXeLaTeXFiles(texPath string) {
	// Get the base filename without extension
	baseFilename := strings.TrimSuffix(texPath, ".tex")

	// List of common XeLaTeX temporary file extensions
	tempExtensions := []string{
		".aux",  // Auxiliary file for cross-references
		".log",  // Log file
		".toc",  // Table of contents
		".out",  // PDF outline/bookmarks
		".lof",  // List of figures
		".lot",  // List of tables
		".fls",  // File list
		".fdb_latexmk", // Latexmk database
	}

	// Remove each temporary file
	for _, ext := range tempExtensions {
		tempFile := baseFilename + ext
		removeFile(tempFile)
	}
}

// checkXeLaTeX verifies that XeLaTeX is installed and available
func checkXeLaTeX() error {
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

// runXeLaTeX executes a single XeLaTeX compilation pass
func runXeLaTeX(inputFile, outputDir string) error {
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
	baseFilename := strings.TrimSuffix(inputFile, ".tex")
	pdfPath := baseFilename + ".pdf"
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