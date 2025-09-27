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
	config                  *models.BookConfig
	sentMessageTemplate     *template.Template
	receivedMessageTemplate *template.Template
	titlePageTemplate       *template.Template
	copyrightPageTemplate   *template.Template
	pageStructureTemplate   *template.Template
	yamlHeaderTemplate      *template.Template
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

// writeFrontmatter writes the YAML frontmatter using template file
func (g *Generator) writeFrontmatter(builder *strings.Builder) {
	templatePath := "templates/yaml-header.yml"
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		// Fallback to basic header if template file can't be read
		builder.WriteString("---\n")
		builder.WriteString(fmt.Sprintf("title: \"%s\"\n", g.config.Title))
		builder.WriteString(fmt.Sprintf("date: \"%s\"\n", time.Now().Format("January 2, 2006")))
		builder.WriteString("---\n\n")
		return
	}

	tmpl, err := template.New("yaml-header").Parse(string(templateContent))
	if err != nil {
		// Fallback to basic header if template parsing fails
		builder.WriteString("---\n")
		builder.WriteString(fmt.Sprintf("title: \"%s\"\n", g.config.Title))
		builder.WriteString(fmt.Sprintf("date: \"%s\"\n", time.Now().Format("January 2, 2006")))
		builder.WriteString("---\n\n")
		return
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
	if err := tmpl.Execute(&buf, data); err != nil {
		// Fallback to basic header if template execution fails
		builder.WriteString("---\n")
		builder.WriteString(fmt.Sprintf("title: \"%s\"\n", g.config.Title))
		builder.WriteString(fmt.Sprintf("date: \"%s\"\n", time.Now().Format("January 2, 2006")))
		builder.WriteString("---\n\n")
		return
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