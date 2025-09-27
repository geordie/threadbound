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

// loadMessageTemplates loads the message bubble templates
func (g *Generator) loadMessageTemplates() {
	// Load sent message template
	sentContent, err := ioutil.ReadFile("templates/sent-message.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load sent message template: %v", err))
	}
	g.sentMessageTemplate, err = template.New("sent-message").Parse(string(sentContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse sent message template: %v", err))
	}

	// Load received message template
	receivedContent, err := ioutil.ReadFile("templates/received-message.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load received message template: %v", err))
	}
	g.receivedMessageTemplate, err = template.New("received-message").Parse(string(receivedContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse received message template: %v", err))
	}

	// Load title page template
	titleContent, err := ioutil.ReadFile("templates/title-page.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load title page template: %v", err))
	}
	g.titlePageTemplate, err = template.New("title-page").Parse(string(titleContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse title page template: %v", err))
	}

	// Load copyright page template
	copyrightContent, err := ioutil.ReadFile("templates/copyright-page.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load copyright page template: %v", err))
	}
	g.copyrightPageTemplate, err = template.New("copyright-page").Parse(string(copyrightContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse copyright page template: %v", err))
	}

	// Load page structure template
	pageStructureContent, err := ioutil.ReadFile("templates/page-structure.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load page structure template: %v", err))
	}
	g.pageStructureTemplate, err = template.New("page-structure").Parse(string(pageStructureContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse page structure template: %v", err))
	}

	// Load YAML header template
	yamlHeaderContent, err := ioutil.ReadFile("templates/yaml-header.yml")
	if err != nil {
		panic(fmt.Sprintf("failed to load YAML header template: %v", err))
	}
	g.yamlHeaderTemplate, err = template.New("yaml-header").Parse(string(yamlHeaderContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse YAML header template: %v", err))
	}

	// Load image attachment template
	imageAttachmentContent, err := ioutil.ReadFile("templates/image-attachment.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load image attachment template: %v", err))
	}
	g.imageAttachmentTemplate, err = template.New("image-attachment").Parse(string(imageAttachmentContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse image attachment template: %v", err))
	}

	// Load image placeholder template
	imagePlaceholderContent, err := ioutil.ReadFile("templates/image-placeholder.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load image placeholder template: %v", err))
	}
	g.imagePlaceholderTemplate, err = template.New("image-placeholder").Parse(string(imagePlaceholderContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse image placeholder template: %v", err))
	}

	// Load attachment template
	attachmentContent, err := ioutil.ReadFile("templates/attachment.tex")
	if err != nil {
		panic(fmt.Sprintf("failed to load attachment template: %v", err))
	}
	g.attachmentTemplate, err = template.New("attachment").Parse(string(attachmentContent))
	if err != nil {
		panic(fmt.Sprintf("failed to parse attachment template: %v", err))
	}
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

	// Table of contents and page structure
	g.writePageStructure(&builder)

	// Group messages by date for better organization
	g.writeMessages(&builder, messages, handles)

	return builder.String()
}

// writeFrontmatter writes the YAML frontmatter using template
func (g *Generator) writeFrontmatter(builder *strings.Builder) {
	// Use template - fail if not available
	if g.yamlHeaderTemplate == nil {
		panic("YAML header template not loaded")
	}

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

	var buf bytes.Buffer
	if err := g.yamlHeaderTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute YAML header template: %v", err))
	}

	builder.WriteString(buf.String())
	builder.WriteString("\n\n")
}

// writeTitlePage writes the title page using template
func (g *Generator) writeTitlePage(builder *strings.Builder) {
	// Use template - fail if not available
	if g.titlePageTemplate == nil {
		panic("title page template not loaded")
	}

	data := struct {
		Title  string
		Author string
		Date   string
	}{
		Title:  g.config.Title,
		Author: g.config.Author,
		Date:   time.Now().Format("January 2, 2006"),
	}

	var buf bytes.Buffer
	if err := g.titlePageTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute title page template: %v", err))
	}

	builder.WriteString(buf.String())
	builder.WriteString("\n\n")
}

// writeCopyrightPage writes the copyright page using template
func (g *Generator) writeCopyrightPage(builder *strings.Builder) {
	// Use template - fail if not available
	if g.copyrightPageTemplate == nil {
		panic("copyright page template not loaded")
	}

	data := struct {
		Year   int
		Author string
	}{
		Year:   time.Now().Year(),
		Author: g.config.Author,
	}

	var buf bytes.Buffer
	if err := g.copyrightPageTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute copyright page template: %v", err))
	}

	builder.WriteString(buf.String())
	builder.WriteString("\n\n")
}

// writePageStructure writes the table of contents and page structure using template
func (g *Generator) writePageStructure(builder *strings.Builder) {
	// Use template - fail if not available
	if g.pageStructureTemplate == nil {
		panic("page structure template not loaded")
	}

	var buf bytes.Buffer
	if err := g.pageStructureTemplate.Execute(&buf, nil); err != nil {
		panic(fmt.Sprintf("failed to execute page structure template: %v", err))
	}

	builder.WriteString(buf.String())
	builder.WriteString("\n\n")
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
	if isFromMe {
		g.writeSentMessageBubble(builder, text, timeStr)
	} else {
		g.writeReceivedMessageBubble(builder, text, timeStr)
	}
}

// writeSentMessageBubble formats a message sent by the user (right-aligned, blue)
func (g *Generator) writeSentMessageBubble(builder *strings.Builder, text string, timeStr string) {
	// Escape LaTeX special characters
	escapedText := g.escapeLaTeX(text)

	// Replace newlines with line breaks
	escapedText = strings.ReplaceAll(escapedText, "\n", "  \n")

	// Use template - fail if not available
	if g.sentMessageTemplate == nil {
		panic("sent message template not loaded")
	}

	data := struct {
		Text      string
		Timestamp string
	}{
		Text:      escapedText,
		Timestamp: timeStr,
	}

	var buf bytes.Buffer
	if err := g.sentMessageTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute sent message template: %v", err))
	}

	builder.WriteString(buf.String())
	builder.WriteString("\n\n")
}

// writeReceivedMessageBubble formats a message received from others (left-aligned, gray)
func (g *Generator) writeReceivedMessageBubble(builder *strings.Builder, text string, timeStr string) {
	// Escape LaTeX special characters
	escapedText := g.escapeLaTeX(text)

	// Replace newlines with line breaks
	escapedText = strings.ReplaceAll(escapedText, "\n", "  \n")

	// Use template - fail if not available
	if g.receivedMessageTemplate == nil {
		panic("received message template not loaded")
	}

	data := struct {
		Text      string
		Timestamp string
	}{
		Text:      escapedText,
		Timestamp: timeStr,
	}

	var buf bytes.Buffer
	if err := g.receivedMessageTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute received message template: %v", err))
	}

	builder.WriteString(buf.String())
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
	if g.imageAttachmentTemplate == nil {
		panic("image attachment template not loaded")
	}

	data := struct {
		Filename string
		Path     string
	}{
		Filename: filename,
		Path:     path,
	}

	var buf bytes.Buffer
	if err := g.imageAttachmentTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute image attachment template: %v", err))
	}

	builder.WriteString(buf.String())
	builder.WriteString("\n\n")
}

// writeImagePlaceholder writes an image placeholder using template
func (g *Generator) writeImagePlaceholder(builder *strings.Builder, filename string) {
	if g.imagePlaceholderTemplate == nil {
		panic("image placeholder template not loaded")
	}

	data := struct {
		Filename string
	}{
		Filename: filename,
	}

	var buf bytes.Buffer
	if err := g.imagePlaceholderTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute image placeholder template: %v", err))
	}

	builder.WriteString(buf.String())
	builder.WriteString("\n\n")
}

// writeAttachment writes a non-image attachment using template
func (g *Generator) writeAttachment(builder *strings.Builder, filename string) {
	if g.attachmentTemplate == nil {
		panic("attachment template not loaded")
	}

	data := struct {
		Filename string
	}{
		Filename: filename,
	}

	var buf bytes.Buffer
	if err := g.attachmentTemplate.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("failed to execute attachment template: %v", err))
	}

	builder.WriteString(buf.String())
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