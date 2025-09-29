package book

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

func TestSectionNumbersAreRemoved(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "book_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test markdown file with date sections
	testMarkdown := `---
title: "Test Book"
author: "Test Author"
date: "September 28, 2025"
---

# Test Book

## Monday, July 28, 2025

Test message 1

## Tuesday, July 29, 2025

Test message 2

## Wednesday, July 30, 2025

Test message 3
`

	markdownPath := filepath.Join(tempDir, "test.md")
	err = os.WriteFile(markdownPath, []byte(testMarkdown), 0644)
	if err != nil {
		t.Fatalf("Failed to write test markdown: %v", err)
	}

	// Build PDF using the same command as the main application
	pdfPath := filepath.Join(tempDir, "test.pdf")
	cmd := exec.Command("pandoc",
		"-f", "markdown",
		"-t", "latex",
		"--template=templates/book.tex",
		"--pdf-engine=xelatex",
		"-o", pdfPath,
		markdownPath)

	// Set working directory to project root so template can be found
	cmd.Dir = "../.."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Pandoc output: %s", string(output))
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	// Generate LaTeX intermediate to check for section numbering
	latexPath := filepath.Join(tempDir, "test.tex")
	cmd = exec.Command("pandoc",
		"-f", "markdown",
		"-t", "latex",
		"--template=templates/book.tex",
		"-o", latexPath,
		markdownPath)

	cmd.Dir = "../.."
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Pandoc output: %s", string(output))
		t.Fatalf("Failed to generate LaTeX: %v", err)
	}

	// Read the generated LaTeX file
	latexContent, err := os.ReadFile(latexPath)
	if err != nil {
		t.Fatalf("Failed to read generated LaTeX: %v", err)
	}

	latexStr := string(latexContent)

	// Check that the template doesn't include section numbering format
	// Look for \thesection in titleformat which would add numbers
	sectionNumberingPattern := regexp.MustCompile(`\\titleformat\{\\section\}.*\\thesection`)
	if sectionNumberingPattern.MatchString(latexStr) {
		t.Errorf("Found \\thesection in section titleformat - this will add section numbers")
	}

	// Also check subsection numbering
	subsectionNumberingPattern := regexp.MustCompile(`\\titleformat\{\\subsection\}.*\\thesubsection`)
	if subsectionNumberingPattern.MatchString(latexStr) {
		t.Errorf("Found \\thesubsection in subsection titleformat - this will add subsection numbers")
	}

	// Verify that sections are formatted without numbers
	// Look for the correct titleformat pattern (allowing for whitespace)
	expectedSectionFormat := regexp.MustCompile(`\\titleformat\{\\section\}\s*\s*\{[^}]*\}\s*\{\}\s*\{0em\}\s*\{\}`)
	if !expectedSectionFormat.MatchString(latexStr) {
		t.Errorf("Section titleformat is not configured to remove numbering")
	}

	t.Logf("✅ Section numbering test passed - no numbered sections found")
}