package markdown

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"imessages-book/internal/models"
)

// Generator handles markdown generation
type Generator struct {
	config                    *models.BookConfig
	sentMessageTemplate       *template.Template
	receivedMessageTemplate   *template.Template
	titlePageTemplate         *template.Template
	copyrightPageTemplate     *template.Template
	pageStructureTemplate     *template.Template
	yamlHeaderTemplate        *template.Template
	imageAttachmentTemplate   *template.Template
	imagePlaceholderTemplate  *template.Template
	attachmentTemplate        *template.Template
}

// New creates a new markdown generator
func New(config *models.BookConfig) *Generator {
	g := &Generator{config: config}
	g.loadMessageTemplates()
	return g
}

// loadTemplate loads and parses a single template file
func (g *Generator) loadTemplate(filename, templateName string) *template.Template {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("failed to load %s template: %v", templateName, err))
	}

	tmpl, err := template.New(templateName).Parse(string(content))
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s template: %v", templateName, err))
	}

	return tmpl
}

// executeTemplate executes a template with data and returns the result
func (g *Generator) executeTemplate(tmpl *template.Template, templateName string, data interface{}) string {
	if tmpl == nil {
		panic(fmt.Sprintf("%s template not loaded", templateName))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute %s template: %v", templateName, err))
	}

	return buf.String()
}

// loadMessageTemplates loads all templates
func (g *Generator) loadMessageTemplates() {
	g.sentMessageTemplate = g.loadTemplate("templates/sent-message.tex", "sent-message")
	g.receivedMessageTemplate = g.loadTemplate("templates/received-message.tex", "received-message")
	g.titlePageTemplate = g.loadTemplate("templates/title-page.tex", "title-page")
	g.copyrightPageTemplate = g.loadTemplate("templates/copyright-page.tex", "copyright-page")
	g.pageStructureTemplate = g.loadTemplate("templates/page-structure.tex", "page-structure")
	g.yamlHeaderTemplate = g.loadTemplate("templates/yaml-header.yml", "yaml-header")
	g.imageAttachmentTemplate = g.loadTemplate("templates/image-attachment.tex", "image-attachment")
	g.imagePlaceholderTemplate = g.loadTemplate("templates/image-placeholder.tex", "image-placeholder")
	g.attachmentTemplate = g.loadTemplate("templates/attachment.tex", "attachment")
}

// GenerateBook creates the complete markdown book
func (g *Generator) GenerateBook(messages []models.Message, handles map[int]models.Handle, reactions map[string][]models.Reaction) string {
	var builder strings.Builder

	// YAML frontmatter
	g.writeFrontmatter(&builder)

	// Title page
	g.writeTitlePage(&builder)

	// Copyright page
	g.writeCopyrightPage(&builder)

	// Table of contents and page structure
	g.writePageStructure(&builder)

	// Group messages by date for better organization
	g.writeMessages(&builder, messages, handles, reactions)

	return builder.String()
}

// writeFrontmatter writes the YAML frontmatter using template
func (g *Generator) writeFrontmatter(builder *strings.Builder) {
	data := struct {
		Title      string
		Author     string
		Date       string
		PageWidth  string
		PageHeight string
	}{
		Title:      g.config.Title,
		Author:     g.config.Author,
		Date:       time.Now().Format("January 2, 2006"),
		PageWidth:  g.config.PageWidth,
		PageHeight: g.config.PageHeight,
	}

	result := g.executeTemplate(g.yamlHeaderTemplate, "YAML header", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
}

// writeTitlePage writes the title page using template
func (g *Generator) writeTitlePage(builder *strings.Builder) {
	data := struct {
		Title  string
		Author string
		Date   string
	}{
		Title:  g.config.Title,
		Author: g.config.Author,
		Date:   time.Now().Format("January 2, 2006"),
	}

	result := g.executeTemplate(g.titlePageTemplate, "title page", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
}

// writeCopyrightPage writes the copyright page using template
func (g *Generator) writeCopyrightPage(builder *strings.Builder) {
	data := struct {
		Year   int
		Author string
	}{
		Year:   time.Now().Year(),
		Author: g.config.Author,
	}

	result := g.executeTemplate(g.copyrightPageTemplate, "copyright page", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
}

// writePageStructure writes the table of contents and page structure using template
func (g *Generator) writePageStructure(builder *strings.Builder) {
	result := g.executeTemplate(g.pageStructureTemplate, "page structure", nil)
	builder.WriteString(result)
	builder.WriteString("\n\n")
}

// writeMessages writes all messages in conversation format
func (g *Generator) writeMessages(builder *strings.Builder, messages []models.Message, handles map[int]models.Handle, reactions map[string][]models.Reaction) {
	var lastDate string
	var lastSender string
	var lastTimestamp string


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
			lastTimestamp = "" // Reset timestamp tracking for new day
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
		messageReactions := reactions[msg.GUID]


		// Write message content in conversation style
		g.writeMessageBubble(builder, *msg.Text, msg.IsFromMe, timeStr, senderName, showSender, showTimestamp, messageReactions)

		// Add attachments if any
		if msg.HasAttachments && g.config.IncludeImages {
			g.writeAttachments(builder, msg.Attachments)
		}

		builder.WriteString("\n")
	}
}

// writeMessageBubble formats a single message as a conversation bubble
func (g *Generator) writeMessageBubble(builder *strings.Builder, text string, isFromMe bool, timeStr string, senderName string, showSender bool, showTimestamp bool, reactions []models.Reaction) {
	if isFromMe {
		g.writeSentMessageBubble(builder, text, timeStr, reactions)
	} else {
		g.writeReceivedMessageBubble(builder, text, timeStr, senderName, showSender, showTimestamp, reactions)
	}
}

// writeSentMessageBubble formats a message sent by the user (right-aligned, blue)
func (g *Generator) writeSentMessageBubble(builder *strings.Builder, text string, timeStr string, reactions []models.Reaction) {
	// Escape LaTeX special characters
	escapedText := g.escapeLaTeX(text)

	// Replace newlines with line breaks
	escapedText = strings.ReplaceAll(escapedText, "\n", "  \n")


	data := struct {
		Text      string
		Timestamp string
		Reactions []models.Reaction
	}{
		Text:      escapedText,
		Timestamp: timeStr,
		Reactions: reactions,
	}

	result := g.executeTemplate(g.sentMessageTemplate, "sent message", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
}

// writeReceivedMessageBubble formats a message received from others (left-aligned, gray)
func (g *Generator) writeReceivedMessageBubble(builder *strings.Builder, text string, timeStr string, senderName string, showSender bool, showTimestamp bool, reactions []models.Reaction) {
	// Escape LaTeX special characters
	escapedText := g.escapeLaTeX(text)

	// Replace newlines with line breaks
	escapedText = strings.ReplaceAll(escapedText, "\n", "  \n")


	data := struct {
		Text          string
		Timestamp     string
		Sender        string
		ShowSender    bool
		ShowTimestamp bool
		Reactions     []models.Reaction
	}{
		Text:          escapedText,
		Timestamp:     timeStr,
		Sender:        senderName,
		ShowSender:    showSender,
		ShowTimestamp: showTimestamp,
		Reactions:     reactions,
	}

	result := g.executeTemplate(g.receivedMessageTemplate, "received message", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
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

// writeAttachments adds attachment references to the markdown using templates
func (g *Generator) writeAttachments(builder *strings.Builder, attachments []models.Attachment) {
	for _, att := range attachments {
		if att.Filename != nil {
			filename := *att.Filename
			ext := strings.ToLower(filepath.Ext(filename))

			// Handle images
			if isImageFile(ext) {
				if att.ProcessedPath != "" {
					g.writeImageAttachment(builder, filename, att.ProcessedPath)
				} else {
					g.writeImagePlaceholder(builder, filename)
				}
			} else {
				// Handle other file types
				g.writeAttachment(builder, filename)
			}
		}
	}
}

// writeImageAttachment writes an image attachment with path using template
func (g *Generator) writeImageAttachment(builder *strings.Builder, filename, path string) {
	data := struct {
		Filename string
		Path     string
	}{
		Filename: filename,
		Path:     path,
	}

	result := g.executeTemplate(g.imageAttachmentTemplate, "image attachment", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
}

// writeImagePlaceholder writes an image placeholder using template
func (g *Generator) writeImagePlaceholder(builder *strings.Builder, filename string) {
	data := struct {
		Filename string
	}{
		Filename: filename,
	}

	result := g.executeTemplate(g.imagePlaceholderTemplate, "image placeholder", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
}

// writeAttachment writes a non-image attachment using template
func (g *Generator) writeAttachment(builder *strings.Builder, filename string) {
	data := struct {
		Filename string
	}{
		Filename: filename,
	}

	result := g.executeTemplate(g.attachmentTemplate, "attachment", data)
	builder.WriteString(result)
	builder.WriteString("\n\n")
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