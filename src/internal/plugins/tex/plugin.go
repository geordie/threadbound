package tex

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"threadbound/internal/models"
	"threadbound/internal/output"
	"threadbound/internal/urlprocessor"
)

// TeXPlugin implements the OutputPlugin interface for TeX generation
type TeXPlugin struct {
	*output.BasePlugin
}

// NewTeXPlugin creates a new TeX plugin instance
func NewTeXPlugin() *TeXPlugin {
	capabilities := output.PluginCapabilities{
		SupportsImages:      true,
		SupportsAttachments: true,
		SupportsReactions:   true,
		SupportsURLPreviews: true,
		RequiresTemplates:   true,
		SupportsPagination:  true,
	}

	base := output.NewBasePlugin(
		"tex",
		"TeX Document",
		"Generate TeX document that can be compiled to PDF with XeLaTeX",
		"tex",
		capabilities,
	)

	return &TeXPlugin{
		BasePlugin: base,
	}
}

// Generate creates a TeX document
func (p *TeXPlugin) Generate(ctx *output.GenerationContext) ([]byte, error) {
	// Create template manager
	tm := output.NewTemplateManager(ctx.Config.TemplateDir)

	// Load all required templates
	if err := tm.LoadTemplates(p.GetRequiredTemplates()); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Process URLs if enabled
	if ctx.Config.IncludePreviews {
		if err := p.processURLs(ctx); err != nil {
			fmt.Printf("⚠️  Warning: URL processing failed: %v\n", err)
		}
	}

	// Generate the book content
	bookContent, err := p.generateBook(ctx, tm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate book: %w", err)
	}

	return []byte(bookContent), nil
}

// processURLs finds and processes all URLs in messages
func (p *TeXPlugin) processURLs(ctx *output.GenerationContext) error {
	// Create a database connection for URL processing
	db, err := sql.Open("sqlite3", ctx.Config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	urlProcessor := urlprocessor.New(ctx.Config, db)
	processedURLs := make(map[string]bool)

	fmt.Printf("🔗 Processing URLs using existing iMessage preview data...\n")

	// Process each message that might have URL previews
	for _, msg := range ctx.Messages {
		if msg.Text != nil {
			urls := urlProcessor.FindURLsInText(*msg.Text)
			if len(urls) > 0 {
				// Extract existing preview data from this message
				messagePreviews := urlProcessor.ProcessMessageForURLPreviews(int64(msg.ID))
				for url, thumbnail := range messagePreviews {
					if !processedURLs[url] {
						ctx.URLThumbnails[url] = thumbnail
						processedURLs[url] = true
						if thumbnail.Success {
							fmt.Printf("✅ Found existing preview for: %s (title: %s)\n", url, thumbnail.Title)
						} else {
							fmt.Printf("⚠️  No preview data found for: %s\n", url)
						}
					}
				}

				// For URLs without existing preview data, try the fallback method
				for _, url := range urls {
					if !processedURLs[url] {
						thumbnail := urlProcessor.ProcessURL(url)
						ctx.URLThumbnails[url] = thumbnail
						processedURLs[url] = true
						if thumbnail.Success {
							fmt.Printf("✅ Generated fallback thumbnail for: %s\n", url)
						} else {
							fmt.Printf("⚠️  Failed to generate thumbnail for: %s\n", url)
						}
					}
				}
			}
		}
	}

	fmt.Printf("🔗 Processed %d unique URLs\n", len(ctx.URLThumbnails))
	return nil
}

// generateBook creates the complete TeX book content
func (p *TeXPlugin) generateBook(ctx *output.GenerationContext, tm *output.TemplateManager) (string, error) {
	// Read the main book template as a raw string (book.tex uses %%PLACEHOLDERS%% not Go templates)
	templatePath := filepath.Join(ctx.Config.TemplateDir, "book.tex")
	templateBytes, err := readTemplateFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read book.tex: %w", err)
	}

	// Generate each component
	variables := p.generateVariables(ctx)
	titlePage := p.generateTitlePage(ctx)
	copyrightPage := p.generateCopyrightPage(ctx)
	content := p.generateContent(ctx, tm)

	// Replace placeholders in template
	result := string(templateBytes)
	result = strings.ReplaceAll(result, "%%VARIABLES%%", variables)
	result = strings.ReplaceAll(result, "%%TITLE_PAGE%%", titlePage)
	result = strings.ReplaceAll(result, "%%COPYRIGHT_PAGE%%", copyrightPage)
	result = strings.ReplaceAll(result, "%%CONTENT%%", content)

	return result, nil
}

// readTemplateFile reads a template file as raw bytes
func readTemplateFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// generateVariables creates LaTeX variable definitions
func (p *TeXPlugin) generateVariables(ctx *output.GenerationContext) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("\\newcommand{\\booktitle}{%s}\n", p.escapeLaTeX(ctx.Config.Title)))
	if ctx.Config.Author != "" {
		builder.WriteString(fmt.Sprintf("\\newcommand{\\bookauthor}{%s}\n", p.escapeLaTeX(ctx.Config.Author)))
	}
	builder.WriteString(fmt.Sprintf("\\newcommand{\\bookdate}{%s}\n", time.Now().Format("January 2, 2006")))
	builder.WriteString(fmt.Sprintf("\\newcommand{\\bookyear}{%d}\n", time.Now().Year()))

	return builder.String()
}

// generateTitlePage creates the title page content
func (p *TeXPlugin) generateTitlePage(ctx *output.GenerationContext) string {
	var builder strings.Builder

	builder.WriteString("\\begin{titlepage}\n")
	builder.WriteString("\\centering\n")
	builder.WriteString("\\vspace*{2cm}\n\n")
	builder.WriteString("{\\Huge\\bfseries \\booktitle}\n\n")
	builder.WriteString("\\vspace{1cm}\n\n")

	if ctx.Config.Author != "" {
		builder.WriteString("{\\Large \\bookauthor}\n\n")
		builder.WriteString("\\vspace{1cm}\n\n")
	}

	builder.WriteString("{\\large \\bookdate}\n\n")
	builder.WriteString("\\vfill\n")
	builder.WriteString("\\end{titlepage}\n")

	return builder.String()
}

// generateCopyrightPage creates the copyright page content
func (p *TeXPlugin) generateCopyrightPage(ctx *output.GenerationContext) string {
	var builder strings.Builder

	builder.WriteString("\\newpage\n\n")
	builder.WriteString("\\thispagestyle{empty}\n\n")
	builder.WriteString("\\vspace*{\\fill}\n\n")
	builder.WriteString("\\begin{flushleft}\n")

	if ctx.Config.Author != "" {
		builder.WriteString("© \\bookyear\\ \\bookauthor\n")
	} else {
		builder.WriteString("© \\bookyear\n")
	}

	builder.WriteString("\\\\[0.5cm]\n\n")
	builder.WriteString("This book contains personal messages and conversations.\n")
	builder.WriteString("All rights reserved. No part of this publication may be reproduced,\n")
	builder.WriteString("distributed, or transmitted in any form or by any means without\n")
	builder.WriteString("the prior written permission of the copyright holder.\n\n")
	builder.WriteString("Generated using threadbound.\n")
	builder.WriteString("\\end{flushleft}\n\n")
	builder.WriteString("\\newpage\n")

	return builder.String()
}

// generateContent creates the main message content
func (p *TeXPlugin) generateContent(ctx *output.GenerationContext, tm *output.TemplateManager) string {
	var builder strings.Builder
	p.writeMessages(&builder, ctx, tm)
	return builder.String()
}

// writeMessages writes all messages in conversation format
func (p *TeXPlugin) writeMessages(builder *strings.Builder, ctx *output.GenerationContext, tm *output.TemplateManager) {
	var lastDate string
	var lastMonth string
	var lastSender string
	var lastTimestamp string

	for _, msg := range ctx.Messages {
		// Skip empty messages
		if msg.Text == nil || strings.TrimSpace(*msg.Text) == "" {
			continue
		}

		// Add month chapter header if month changed
		currentMonth := msg.FormattedDate.Format("January 2006")
		if currentMonth != lastMonth {
			builder.WriteString(fmt.Sprintf("\n\\chapter{%s}\n\n", p.escapeLaTeX(currentMonth)))
			lastMonth = currentMonth
		}

		// Add date section header if day changed
		currentDate := msg.FormattedDate.Format("Monday, January 2, 2006")
		if currentDate != lastDate {
			builder.WriteString(fmt.Sprintf("\n\\section{%s}\n\n", p.escapeLaTeX(currentDate)))
			lastDate = currentDate
			lastSender = ""
			lastTimestamp = ""
		}

		// Determine sender
		senderName := output.GetSenderName(msg, ctx.Handles)

		// Check if we should show sender name (when it changes)
		showSender := (senderName != lastSender)
		if showSender {
			lastSender = senderName
		}

		// Format time
		timeStr := msg.FormattedDate.Format("3:04 PM")

		// Check if we should show timestamp (when sender changes or timestamp changes)
		showTimestamp := showSender || (timeStr != lastTimestamp)
		if showTimestamp {
			lastTimestamp = timeStr
		}

		// Get reactions for this message
		messageReactions := ctx.Reactions[msg.GUID]

		// Write message content
		p.writeMessageBubble(builder, ctx, tm, msg, *msg.Text, timeStr, senderName, showSender, showTimestamp, messageReactions)

		// Add attachments if any
		if msg.HasAttachments && ctx.Config.IncludeImages {
			p.writeAttachments(builder, tm, msg.Attachments)
		}

		builder.WriteString("\n")
	}
}

// writeMessageBubble formats a single message as a conversation bubble
func (p *TeXPlugin) writeMessageBubble(builder *strings.Builder, ctx *output.GenerationContext, tm *output.TemplateManager,
	msg models.Message, text, timeStr, senderName string, showSender, showTimestamp bool, reactions []models.Reaction) {

	// Process text for URLs
	processedText := text
	if ctx.URLThumbnails != nil && len(ctx.URLThumbnails) > 0 {
		processedText = p.replaceURLsWithImages(text, ctx.URLThumbnails)
	}

	// Escape LaTeX special characters
	escapedText := p.escapeLaTeX(processedText)

	// Replace newlines with line breaks
	escapedText = strings.ReplaceAll(escapedText, "\n", "  \n")

	if msg.IsFromMe {
		p.writeSentMessage(builder, tm, escapedText, timeStr, reactions)
	} else {
		p.writeReceivedMessage(builder, tm, escapedText, timeStr, senderName, showSender, showTimestamp, reactions)
	}
}

// writeSentMessage formats a message sent by the user
func (p *TeXPlugin) writeSentMessage(builder *strings.Builder, tm *output.TemplateManager, text, timeStr string, reactions []models.Reaction) {
	data := struct {
		Text      string
		Timestamp string
		Reactions []models.Reaction
	}{
		Text:      text,
		Timestamp: timeStr,
		Reactions: reactions,
	}

	result, err := tm.ExecuteTemplate("sent-message.tex", data)
	if err != nil {
		// Fallback to simple format
		builder.WriteString(fmt.Sprintf("\\sentmessage{%s}{%s}\n", text, timeStr))
	} else {
		builder.WriteString(result)
	}
	builder.WriteString("\n\n")
}

// writeReceivedMessage formats a message received from others
func (p *TeXPlugin) writeReceivedMessage(builder *strings.Builder, tm *output.TemplateManager, text, timeStr, senderName string,
	showSender, showTimestamp bool, reactions []models.Reaction) {

	data := struct {
		Text          string
		Timestamp     string
		Sender        string
		ShowSender    bool
		ShowTimestamp bool
		Reactions     []models.Reaction
	}{
		Text:          text,
		Timestamp:     timeStr,
		Sender:        senderName,
		ShowSender:    showSender,
		ShowTimestamp: showTimestamp,
		Reactions:     reactions,
	}

	result, err := tm.ExecuteTemplate("received-message.tex", data)
	if err != nil {
		// Fallback to simple format
		builder.WriteString(fmt.Sprintf("\\receivedmessage{%s}{%s}{%s}\n", senderName, text, timeStr))
	} else {
		builder.WriteString(result)
	}
	builder.WriteString("\n\n")
}

// replaceURLsWithImages replaces URLs with LaTeX image commands
func (p *TeXPlugin) replaceURLsWithImages(text string, thumbnails map[string]*output.URLThumbnail) string {
	urlRegex := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)

	return urlRegex.ReplaceAllStringFunc(text, func(url string) string {
		cleanURL := strings.TrimRight(url, ".,;!?)")

		if thumbnail, exists := thumbnails[cleanURL]; exists && thumbnail.Success && thumbnail.ThumbnailPath != "" {
			return fmt.Sprintf("\\messageimage{%s}", thumbnail.ThumbnailPath)
		}

		return url
	})
}

// writeAttachments adds attachment references to the output
func (p *TeXPlugin) writeAttachments(builder *strings.Builder, tm *output.TemplateManager, attachments []models.Attachment) {
	for _, att := range attachments {
		if att.Filename != nil {
			filename := *att.Filename
			ext := strings.ToLower(filepath.Ext(filename))

			// Handle images
			if p.isImageFile(ext) {
				if att.ProcessedPath != "" {
					p.writeImageAttachment(builder, tm, filename, att.ProcessedPath)
				} else {
					p.writeImagePlaceholder(builder, tm, filename)
				}
			} else {
				// Handle other file types
				p.writeAttachment(builder, tm, filename)
			}
		}
	}
}

// writeImageAttachment writes an image attachment
func (p *TeXPlugin) writeImageAttachment(builder *strings.Builder, tm *output.TemplateManager, filename, path string) {
	data := struct {
		Filename string
		Path     string
	}{
		Filename: filename,
		Path:     path,
	}

	result, err := tm.ExecuteTemplate("image-attachment.tex", data)
	if err != nil {
		builder.WriteString(fmt.Sprintf("\\includegraphics[width=0.8\\textwidth]{%s}\n", path))
	} else {
		builder.WriteString(result)
	}
	builder.WriteString("\n\n")
}

// writeImagePlaceholder writes an image placeholder
func (p *TeXPlugin) writeImagePlaceholder(builder *strings.Builder, tm *output.TemplateManager, filename string) {
	data := struct {
		Filename string
	}{
		Filename: filename,
	}

	result, err := tm.ExecuteTemplate("image-placeholder.tex", data)
	if err != nil {
		builder.WriteString(fmt.Sprintf("\\textit{[Image: %s]}\n", filename))
	} else {
		builder.WriteString(result)
	}
	builder.WriteString("\n\n")
}

// writeAttachment writes a non-image attachment
func (p *TeXPlugin) writeAttachment(builder *strings.Builder, tm *output.TemplateManager, filename string) {
	data := struct {
		Filename string
	}{
		Filename: filename,
	}

	result, err := tm.ExecuteTemplate("attachment.tex", data)
	if err != nil {
		builder.WriteString(fmt.Sprintf("\\textit{[Attachment: %s]}\n", filename))
	} else {
		builder.WriteString(result)
	}
	builder.WriteString("\n\n")
}

// escapeLaTeX escapes special LaTeX characters while preserving image commands
func (p *TeXPlugin) escapeLaTeX(text string) string {
	// First, protect image commands by temporarily replacing them
	imageCommands := make(map[string]string)
	imageRegex := regexp.MustCompile(`\\messageimage\{[^}]+\}`)
	matches := imageRegex.FindAllString(text, -1)

	for i, match := range matches {
		placeholder := fmt.Sprintf("IMAGECOMMAND%d", i)
		imageCommands[placeholder] = match
		text = strings.ReplaceAll(text, match, placeholder)
	}

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

	// Restore protected image commands
	for placeholder, imageCommand := range imageCommands {
		text = strings.ReplaceAll(text, placeholder, imageCommand)
	}

	return text
}

// isImageFile checks if the file extension indicates an image
func (p *TeXPlugin) isImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".heic"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// ValidateConfig validates the TeX plugin configuration
func (p *TeXPlugin) ValidateConfig(config *models.BookConfig) error {
	// Call base validation first
	if err := p.BasePlugin.ValidateConfig(config); err != nil {
		return err
	}

	// Check for required template directory
	if config.TemplateDir == "" {
		return fmt.Errorf("template directory is required for TeX generation")
	}

	return nil
}

// GetRequiredTemplates returns the list of template files needed for TeX generation
func (p *TeXPlugin) GetRequiredTemplates() []string {
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