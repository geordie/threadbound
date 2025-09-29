package markdown

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"threadbound/internal/models"
	"threadbound/internal/urlprocessor"
)

// Generator handles markdown generation
type Generator struct {
	config                    *models.BookConfig
	urlProcessor              *urlprocessor.URLProcessor
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
func New(config *models.BookConfig, db *sql.DB) *Generator {
	g := &Generator{
		config:       config,
		urlProcessor: urlprocessor.New(config, db),
	}
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

	// Process URLs first if enabled
	var urlThumbnails map[string]*urlprocessor.URLThumbnail
	if g.config.IncludePreviews {
		urlThumbnails = g.processAllURLs(messages)
	}

	// YAML frontmatter
	g.writeFrontmatter(&builder, urlThumbnails)

	// Title page
	g.writeTitlePage(&builder)

	// Copyright page
	g.writeCopyrightPage(&builder)

	// Table of contents and page structure
	g.writePageStructure(&builder)

	// Group messages by date for better organization
	g.writeMessages(&builder, messages, handles, reactions, urlThumbnails)

	return builder.String()
}

// processAllURLs finds and processes all URLs in messages using existing iMessage preview data
func (g *Generator) processAllURLs(messages []models.Message) map[string]*urlprocessor.URLThumbnail {
	fmt.Printf("🔗 Processing URLs using existing iMessage preview data...\n")

	urlThumbnails := make(map[string]*urlprocessor.URLThumbnail)
	processedURLs := make(map[string]bool)

	// Process each message that might have URL previews
	for _, msg := range messages {
		if msg.Text != nil {
			urls := g.urlProcessor.FindURLsInText(*msg.Text)
			if len(urls) > 0 {
				// Extract existing preview data from this message
				messagePreviews := g.urlProcessor.ProcessMessageForURLPreviews(int64(msg.ID))
				for url, thumbnail := range messagePreviews {
					if !processedURLs[url] {
						urlThumbnails[url] = thumbnail
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
						thumbnail := g.urlProcessor.ProcessURL(url)
						urlThumbnails[url] = thumbnail
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

	fmt.Printf("🔗 Processed %d unique URLs\n", len(urlThumbnails))
	return urlThumbnails
}

// writeURLSetupFile writes LaTeX commands for URL processing to a file
func (g *Generator) writeURLSetupFile() {
	content := `% Additional commands for URL processing
\newunicodechar{😂}{{\emojifont\symbol{"1F602}}}
\newunicodechar{🤣}{{\emojifont\symbol{"1F923}}}
\newunicodechar{👍}{{\emojifont\symbol{"1F44D}}}
\newunicodechar{🤷}{{\emojifont\symbol{"1F937}}}
\newunicodechar{😀}{{\emojifont\symbol{"1F600}}}
\newunicodechar{⭐}{{\emojifont\symbol{"2B50}}}
\newunicodechar{😊}{{\emojifont\symbol{"1F60A}}}
\newunicodechar{❗}{{\emojifont\symbol{"2757}}}
\newunicodechar{💪}{{\emojifont\symbol{"1F4AA}}}
\newunicodechar{🏼}{}
\newunicodechar{️}{}
\newunicodechar{‍}{}

\newcommand{\messageimage}[1]{%
  \begin{center}
  \includegraphics[width=0.8\textwidth,height=0.4\textheight,keepaspectratio]{#1}
  \end{center}
  \vspace{0.3cm}
}
`

	err := ioutil.WriteFile("templates/url-setup.tex", []byte(content), 0644)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not write URL setup file: %v\n", err)
	}
}

// writeFrontmatter writes the YAML frontmatter using template
func (g *Generator) writeFrontmatter(builder *strings.Builder, urlThumbnails map[string]*urlprocessor.URLThumbnail) {
	// Write URL setup file if needed
	if urlThumbnails != nil && len(urlThumbnails) > 0 {
		g.writeURLSetupFile()
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
func (g *Generator) writeMessages(builder *strings.Builder, messages []models.Message, handles map[int]models.Handle, reactions map[string][]models.Reaction, urlThumbnails map[string]*urlprocessor.URLThumbnail) {
	var lastDate string
	var lastMonth string
	var lastSender string
	var lastTimestamp string


	for _, msg := range messages {
		// Skip empty messages
		if msg.Text == nil || strings.TrimSpace(*msg.Text) == "" {
			continue
		}

		// Add month chapter header if month changed
		currentMonth := msg.FormattedDate.Format("January 2006")
		if currentMonth != lastMonth {
			builder.WriteString(fmt.Sprintf("\n# %s\n\n", currentMonth))
			lastMonth = currentMonth
		}

		// Add date section header if day changed
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
		g.writeMessageBubble(builder, *msg.Text, msg.IsFromMe, timeStr, senderName, showSender, showTimestamp, messageReactions, urlThumbnails)

		// Add attachments if any
		if msg.HasAttachments && g.config.IncludeImages {
			g.writeAttachments(builder, msg.Attachments)
		}

		builder.WriteString("\n")
	}
}

// writeMessageBubble formats a single message as a conversation bubble
func (g *Generator) writeMessageBubble(builder *strings.Builder, text string, isFromMe bool, timeStr string, senderName string, showSender bool, showTimestamp bool, reactions []models.Reaction, urlThumbnails map[string]*urlprocessor.URLThumbnail) {
	if isFromMe {
		g.writeSentMessageBubble(builder, text, timeStr, reactions, urlThumbnails)
	} else {
		g.writeReceivedMessageBubble(builder, text, timeStr, senderName, showSender, showTimestamp, reactions, urlThumbnails)
	}
}

// writeSentMessageBubble formats a message sent by the user (right-aligned, blue)
func (g *Generator) writeSentMessageBubble(builder *strings.Builder, text string, timeStr string, reactions []models.Reaction, urlThumbnails map[string]*urlprocessor.URLThumbnail) {
	// Replace URLs with images if thumbnails available
	processedText := text
	if urlThumbnails != nil {
		processedText = g.urlProcessor.ReplaceURLsWithImages(text, urlThumbnails)
	}

	// Escape LaTeX special characters
	escapedText := g.escapeLaTeX(processedText)

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
func (g *Generator) writeReceivedMessageBubble(builder *strings.Builder, text string, timeStr string, senderName string, showSender bool, showTimestamp bool, reactions []models.Reaction, urlThumbnails map[string]*urlprocessor.URLThumbnail) {
	// Replace URLs with images if thumbnails available
	processedText := text
	if urlThumbnails != nil {
		processedText = g.urlProcessor.ReplaceURLsWithImages(text, urlThumbnails)
	}

	// Escape LaTeX special characters
	escapedText := g.escapeLaTeX(processedText)

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

// escapeLaTeX escapes special LaTeX characters while preserving image commands
func (g *Generator) escapeLaTeX(text string) string {
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