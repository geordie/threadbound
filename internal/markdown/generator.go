package markdown

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"imessages-book/internal/models"
)

// Generator handles markdown generation
type Generator struct {
	config *models.BookConfig
}

// New creates a new markdown generator
func New(config *models.BookConfig) *Generator {
	return &Generator{config: config}
}

// GenerateBook creates the complete markdown book
func (g *Generator) GenerateBook(messages []models.Message, handles map[int]models.Handle) string {
	var builder strings.Builder

	// YAML frontmatter
	g.writeFrontmatter(&builder)

	// Title page
	g.writeTitlePage(&builder)

	// Copyright page
	g.writeCopyrightPage(&builder)

	// Table of contents placeholder
	builder.WriteString("\\newpage\n\n# Table of Contents\n\n")
	builder.WriteString("\\tableofcontents\n\n")

	// Main content
	builder.WriteString("\\newpage\n\n# Messages\n\n")

	// Group messages by date for better organization
	g.writeMessages(&builder, messages, handles)

	return builder.String()
}

// writeFrontmatter writes the YAML frontmatter
func (g *Generator) writeFrontmatter(builder *strings.Builder) {
	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("title: \"%s\"\n", g.config.Title))
	if g.config.Author != "" {
		builder.WriteString(fmt.Sprintf("author: \"%s\"\n", g.config.Author))
	}
	builder.WriteString(fmt.Sprintf("date: \"%s\"\n", time.Now().Format("January 2, 2006")))
	builder.WriteString("documentclass: book\n")
	builder.WriteString("fontsize: 10pt\n")
	builder.WriteString("geometry:\n")
	builder.WriteString(fmt.Sprintf("  - paperwidth=%s\n", g.config.PageWidth))
	builder.WriteString(fmt.Sprintf("  - paperheight=%s\n", g.config.PageHeight))
	builder.WriteString("  - margin=0.5in\n")
	builder.WriteString("mainfont: \"SF Pro Text\"\n")
	builder.WriteString("monofont: \"SF Mono\"\n")
	builder.WriteString("---\n\n")
}

// writeTitlePage writes the title page
func (g *Generator) writeTitlePage(builder *strings.Builder) {
	builder.WriteString("\\begin{titlepage}\n")
	builder.WriteString("\\centering\n")
	builder.WriteString("\\vspace*{2cm}\n\n")
	builder.WriteString(fmt.Sprintf("{\\Huge\\bfseries %s}\n\n", g.config.Title))
	builder.WriteString("\\vspace{1cm}\n\n")
	if g.config.Author != "" {
		builder.WriteString(fmt.Sprintf("{\\Large %s}\n\n", g.config.Author))
		builder.WriteString("\\vspace{1cm}\n\n")
	}
	builder.WriteString(fmt.Sprintf("{\\large %s}\n\n", time.Now().Format("January 2, 2006")))
	builder.WriteString("\\vfill\n")
	builder.WriteString("\\end{titlepage}\n\n")
}

// writeCopyrightPage writes the copyright page
func (g *Generator) writeCopyrightPage(builder *strings.Builder) {
	builder.WriteString("\\newpage\n\n")
	builder.WriteString("\\thispagestyle{empty}\n\n")
	builder.WriteString("\\vspace*{\\fill}\n\n")
	builder.WriteString("\\begin{flushleft}\n")
	builder.WriteString(fmt.Sprintf("© %d", time.Now().Year()))
	if g.config.Author != "" {
		builder.WriteString(fmt.Sprintf(" %s", g.config.Author))
	}
	builder.WriteString("\\\\[0.5cm]\n\n")
	builder.WriteString("This book contains personal messages and conversations.\n")
	builder.WriteString("All rights reserved. No part of this publication may be reproduced,\n")
	builder.WriteString("distributed, or transmitted in any form or by any means without\n")
	builder.WriteString("the prior written permission of the copyright holder.\n\n")
	builder.WriteString("Generated using imessages-book tool.\n")
	builder.WriteString("\\end{flushleft}\n\n")
}

// writeMessages writes all messages in conversation format
func (g *Generator) writeMessages(builder *strings.Builder, messages []models.Message, handles map[int]models.Handle) {
	var lastDate string
	var lastSender string

	for _, msg := range messages {
		// Skip empty messages
		if msg.Text == nil || strings.TrimSpace(*msg.Text) == "" {
			continue
		}

		// Add date header if day changed
		currentDate := msg.FormattedDate.Format("Monday, January 2, 2006")
		if currentDate != lastDate {
			builder.WriteString(fmt.Sprintf("\n## %s\n\n", currentDate))
			lastDate = currentDate
			lastSender = "" // Reset sender tracking for new day
		}

		// Determine sender
		var senderName string
		if msg.IsFromMe {
			senderName = "Me"
		} else {
			if msg.HandleID != nil {
				if handle, exists := handles[*msg.HandleID]; exists {
					senderName = handle.DisplayName
				} else {
					senderName = "Unknown"
				}
			} else {
				senderName = "Unknown"
			}
		}

		// Only show sender name if it changed
		if senderName != lastSender {
			builder.WriteString(fmt.Sprintf("**%s** ", senderName))
			lastSender = senderName
		}

		// Format time
		timeStr := msg.FormattedDate.Format("3:04 PM")

		// Write message content in conversation style
		g.writeMessageBubble(builder, *msg.Text, msg.IsFromMe, timeStr)

		// Add attachments if any
		if msg.HasAttachments && g.config.IncludeImages {
			g.writeAttachments(builder, msg.Attachments)
		}

		builder.WriteString("\n")
	}
}

// writeMessageBubble formats a single message as a conversation bubble
func (g *Generator) writeMessageBubble(builder *strings.Builder, text string, isFromMe bool, timeStr string) {
	// Escape LaTeX special characters instead of HTML
	escapedText := g.escapeLaTeX(text)

	// Replace newlines with line breaks
	escapedText = strings.ReplaceAll(escapedText, "\n", "  \n")

	if isFromMe {
		// Right-aligned message (sent)
		builder.WriteString(fmt.Sprintf("\\begin{flushright}\n"))
		builder.WriteString(fmt.Sprintf("\\colorbox{blue!20}{\\parbox{0.7\\textwidth}{%s}}\n", escapedText))
		builder.WriteString(fmt.Sprintf("\\\\[0.1cm]\n"))
		builder.WriteString(fmt.Sprintf("{\\small\\textcolor{gray}{%s}}\n", timeStr))
		builder.WriteString(fmt.Sprintf("\\end{flushright}\n\n"))
	} else {
		// Left-aligned message (received)
		builder.WriteString(fmt.Sprintf("\\colorbox{gray!20}{\\parbox{0.7\\textwidth}{%s}}\n", escapedText))
		builder.WriteString(fmt.Sprintf("\\\\[0.1cm]\n"))
		builder.WriteString(fmt.Sprintf("{\\small\\textcolor{gray}{%s}}\n\n", timeStr))
	}
}

// escapeLaTeX escapes special LaTeX characters
func (g *Generator) escapeLaTeX(text string) string {
	// Replace LaTeX special characters
	text = strings.ReplaceAll(text, "\\", "\\textbackslash{}")
	text = strings.ReplaceAll(text, "{", "\\{")
	text = strings.ReplaceAll(text, "}", "\\}")
	text = strings.ReplaceAll(text, "$", "\\$")
	text = strings.ReplaceAll(text, "&", "\\&")
	text = strings.ReplaceAll(text, "%", "\\%")
	text = strings.ReplaceAll(text, "#", "\\#")
	text = strings.ReplaceAll(text, "^", "\\textasciicircum{}")
	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, "~", "\\textasciitilde{}")
	return text
}

// writeAttachments adds attachment references to the markdown
func (g *Generator) writeAttachments(builder *strings.Builder, attachments []models.Attachment) {
	for _, att := range attachments {
		if att.Filename != nil {
			filename := *att.Filename
			ext := strings.ToLower(filepath.Ext(filename))

			// Handle images
			if isImageFile(ext) {
				if att.ProcessedPath != "" {
					builder.WriteString(fmt.Sprintf("![Image: %s](%s){width=2.5in}\n\n", filename, att.ProcessedPath))
				} else {
					builder.WriteString(fmt.Sprintf("*[Image: %s]*\n\n", filename))
				}
			} else {
				// Handle other file types
				builder.WriteString(fmt.Sprintf("*[Attachment: %s]*\n\n", filename))
			}
		}
	}
}

// isImageFile checks if the file extension indicates an image
func isImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".heic"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}