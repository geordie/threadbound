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

// checkPandoc verifies that Pandoc is installed and available
func checkPandoc() error {
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

// runPandoc executes pandoc with the given arguments
func runPandoc(args []string) error {
	fmt.Printf("🔨 Building PDF with Pandoc...\n")

	// Execute Pandoc
	cmd := exec.Command("pandoc", args...)
	cmd.Dir = "."

	// Capture output
	output, err := cmd.CombinedOutput()

	// Determine output file from args
	var outputFile string
	for i, arg := range args {
		if arg == "--output" && i+1 < len(args) {
			outputFile = args[i+1]
			break
		}
	}

	// Check if PDF was created despite errors (LaTeX often succeeds with warnings)
	pdfExists := false
	if outputFile != "" {
		if _, statErr := os.Stat(outputFile); statErr == nil {
			pdfExists = true
		}
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
	if outputFile != "" {
		if _, err := os.Stat(outputFile); err != nil {
			return fmt.Errorf("PDF was not created: %s", outputFile)
		}
		fmt.Printf("✅ PDF generated successfully: %s\n", outputFile)
	}

	return nil
}