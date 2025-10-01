package texgen

import (
	"strings"
	"testing"
	"time"

	"threadbound/internal/models"
)

// TestTexGeneration tests the new pure TeX generation approach
func TestTexGeneration(t *testing.T) {
	config := &models.BookConfig{
		Title:       "Test Book",
		Author:      "Test Author",
		PageWidth:   "5.5in",
		PageHeight:  "8.5in",
		TemplateDir: "../templates/tex",
	}

	// Create a minimal generator without database (we don't need URL processing for this test)
	gen := &Generator{
		config: config,
	}

	// Test generateVariables
	t.Run("generateVariables", func(t *testing.T) {
		vars := gen.generateVariables()

		if !strings.Contains(vars, "\\newcommand{\\booktitle}{Test Book}") {
			t.Errorf("Expected booktitle command in variables, got: %s", vars)
		}
		if !strings.Contains(vars, "\\newcommand{\\bookauthor}{Test Author}") {
			t.Errorf("Expected bookauthor command in variables, got: %s", vars)
		}
		if !strings.Contains(vars, "\\newcommand{\\bookdate}") {
			t.Errorf("Expected bookdate command in variables, got: %s", vars)
		}
		if !strings.Contains(vars, "\\newcommand{\\bookyear}") {
			t.Errorf("Expected bookyear command in variables, got: %s", vars)
		}
	})

	// Test generateTitlePage
	t.Run("generateTitlePage", func(t *testing.T) {
		titlePage := gen.generateTitlePage()

		if !strings.Contains(titlePage, "\\begin{titlepage}") {
			t.Errorf("Expected titlepage environment in output")
		}
		if !strings.Contains(titlePage, "\\booktitle") {
			t.Errorf("Expected booktitle reference in title page")
		}
		if !strings.Contains(titlePage, "\\bookauthor") {
			t.Errorf("Expected bookauthor reference in title page")
		}
		if !strings.Contains(titlePage, "\\bookdate") {
			t.Errorf("Expected bookdate reference in title page")
		}
		if !strings.Contains(titlePage, "\\end{titlepage}") {
			t.Errorf("Expected closing titlepage tag")
		}
	})

	// Test generateCopyrightPage
	t.Run("generateCopyrightPage", func(t *testing.T) {
		copyrightPage := gen.generateCopyrightPage()

		if !strings.Contains(copyrightPage, "\\begin{flushleft}") {
			t.Errorf("Expected flushleft environment in copyright page")
		}
		if !strings.Contains(copyrightPage, "\\bookyear") {
			t.Errorf("Expected bookyear reference in copyright page")
		}
		if !strings.Contains(copyrightPage, "\\bookauthor") {
			t.Errorf("Expected bookauthor reference in copyright page")
		}
		if !strings.Contains(copyrightPage, "This book contains personal messages") {
			t.Errorf("Expected copyright notice text")
		}
		if !strings.Contains(copyrightPage, "\\end{flushleft}") {
			t.Errorf("Expected closing flushleft tag")
		}
	})

	// Test that chapter and section commands are generated correctly
	t.Run("chapterAndSectionGeneration", func(t *testing.T) {
		// Load templates for this test
		gen.loadMessageTemplates()

		// Create mock data
		baseTime := time.Date(2025, 9, 27, 10, 0, 0, 0, time.UTC)
		text1 := "Test message"

		messages := []models.Message{
			{
				ID:            1,
				GUID:          "MSG-001",
				Text:          &text1,
				Date:          int64(baseTime.UnixNano()),
				FormattedDate: baseTime,
				IsFromMe:      true,
			},
		}

		handles := map[int]models.Handle{}
		reactions := map[string][]models.Reaction{}

		content := gen.generateContent(messages, handles, reactions, nil)

		// Verify chapter command is used for month
		if !strings.Contains(content, "\\chapter{September 2025}") {
			t.Errorf("Expected \\chapter command for month, got: %s", content)
		}

		// Verify section command is used for date
		if !strings.Contains(content, "\\section{Saturday, September 27, 2025}") {
			t.Errorf("Expected \\section command for date, got: %s", content)
		}

		// Verify no markdown headings remain
		if strings.Contains(content, "# September") || strings.Contains(content, "## Friday") {
			t.Errorf("Found markdown headings in content, should be TeX commands: %s", content)
		}
	})
}

// TestTexEscaping tests that LaTeX special characters are properly escaped
func TestTexEscaping(t *testing.T) {
	config := &models.BookConfig{
		Title:      "Test & Special $Chars%",
		Author:     "Test_Author#",
		TemplateDir: "../templates/tex",
	}

	gen := &Generator{
		config: config,
	}

	vars := gen.generateVariables()

	// Verify special characters are escaped
	if !strings.Contains(vars, "Test \\& Special \\$Chars\\%") {
		t.Errorf("Special characters not properly escaped in title: %s", vars)
	}
	if !strings.Contains(vars, "Test\\_Author\\#") {
		t.Errorf("Special characters not properly escaped in author: %s", vars)
	}
}