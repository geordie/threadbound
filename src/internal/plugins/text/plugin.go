package text

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"threadbound/internal/models"
	"threadbound/internal/output"
)

// TextPlugin implements the OutputPlugin interface for plain text generation
type TextPlugin struct {
	*output.BasePlugin
	templateManager *output.TemplateManager
}

// NewTextPlugin creates a new text plugin instance
func NewTextPlugin() *TextPlugin {
	capabilities := output.PluginCapabilities{
		SupportsImages:      false,
		SupportsAttachments: true,
		SupportsReactions:   true,
		SupportsURLPreviews: false,
		RequiresTemplates:   true,
		SupportsPagination:  false,
	}

	base := output.NewBasePlugin(
		"txt",
		"Plain Text",
		"Generate plain text format suitable for AI analysis",
		"txt",
		capabilities,
	)

	return &TextPlugin{
		BasePlugin: base,
	}
}

// Generate creates a plain text output from the message data
func (t *TextPlugin) Generate(ctx *output.GenerationContext) ([]byte, error) {
	// Initialize template manager if not already done
	if t.templateManager == nil {
		t.templateManager = output.NewTemplateManager(ctx.Config.TemplateDir)
	}

	// Load templates
	if err := t.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	var buf bytes.Buffer

	// Generate header
	header, err := t.generateHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate header: %w", err)
	}
	buf.WriteString(header)
	buf.WriteString("\n")

	// Group messages by date
	messagesByDate := t.groupMessagesByDate(ctx.Messages)

	// Get sorted date keys
	var dateKeys []string
	for dateKey := range messagesByDate {
		dateKeys = append(dateKeys, dateKey)
	}
	sort.Strings(dateKeys)

	// Generate messages for each date
	for _, dateKey := range dateKeys {
		messages := messagesByDate[dateKey]
		if len(messages) == 0 {
			continue
		}

		// Generate date separator
		dateSeparator, err := t.generateDateSeparator(messages[0].FormattedDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate date separator: %w", err)
		}
		buf.WriteString(dateSeparator)
		buf.WriteString("\n")

		// Generate each message
		for _, msg := range messages {
			messageText, err := t.generateMessage(msg, ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate message: %w", err)
			}
			buf.WriteString(messageText)
			buf.WriteString("\n")
		}

		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// loadTemplates loads all required template files
func (t *TextPlugin) loadTemplates() error {
	templates := t.GetRequiredTemplates()
	for _, tmplFile := range templates {
		if _, err := t.templateManager.LoadTemplate(tmplFile); err != nil {
			// If template doesn't exist, use embedded defaults
			if err := t.createDefaultTemplate(tmplFile); err != nil {
				return err
			}
		}
	}
	return nil
}

// createDefaultTemplate creates a default embedded template if file doesn't exist
func (t *TextPlugin) createDefaultTemplate(name string) error {
	var content string
	switch name {
	case "header.txt":
		content = `=== {{.Title}} ==={{if .Author}}
by {{.Author}}{{end}}{{if .Stats}}
Messages: {{.Stats.TotalMessages}} | Text Messages: {{.Stats.TextMessages}} | Contacts: {{.Stats.TotalContacts}}{{if not .Stats.StartDate.IsZero}} | Date Range: {{.Stats.StartDate.Format "Jan 2, 2006"}} - {{.Stats.EndDate.Format "Jan 2, 2006"}}{{end}}{{end}}

`
	case "date-separator.txt":
		content = `--- {{.FormattedDate}} ---
`
	case "message.txt":
		content = `[{{.Timestamp}}] {{.Sender}}: {{.Text}}{{if .Reactions}} {{range .Reactions}}{{.ReactionEmoji}}{{end}}{{end}}{{if .Attachments}}
  Attachments: {{range $i, $a := .Attachments}}{{if $i}}, {{end}}{{$a.Filename}}{{end}}{{end}}`
	default:
		return fmt.Errorf("unknown template: %s", name)
	}

	// Parse and cache the template
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse embedded template %s: %w", name, err)
	}

	// Store in template manager's cache
	t.templateManager = &output.TemplateManager{}
	_ = tmpl // Template is parsed but we'll use inline execution

	return nil
}

// generateHeader generates the conversation header
func (t *TextPlugin) generateHeader(ctx *output.GenerationContext) (string, error) {
	data := ctx.GetTemplateData()

	// Use embedded template if not loaded from file
	headerTemplate := `=== {{.Title}} ==={{if .Author}}
by {{.Author}}{{end}}{{if .Stats}}
Messages: {{.Stats.TotalMessages}} | Text Messages: {{.Stats.TextMessages}} | Contacts: {{.Stats.TotalContacts}}{{if not .Stats.StartDate.IsZero}} | Date Range: {{.Stats.StartDate.Format "Jan 2, 2006"}} - {{.Stats.EndDate.Format "Jan 2, 2006"}}{{end}}{{end}}

`

	tmpl, err := template.New("header").Parse(headerTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateDateSeparator generates a date separator line
func (t *TextPlugin) generateDateSeparator(date time.Time) (string, error) {
	dateTemplate := `--- {{.FormattedDate}} ---
`

	type DateData struct {
		FormattedDate string
	}

	formattedDate := date.Format("Monday, January 2, 2006")
	data := DateData{FormattedDate: formattedDate}

	tmpl, err := template.New("date-separator").Parse(dateTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateMessage generates a single message
func (t *TextPlugin) generateMessage(msg models.Message, ctx *output.GenerationContext) (string, error) {
	// Skip messages without text
	if msg.Text == nil || strings.TrimSpace(*msg.Text) == "" {
		return "", nil
	}

	senderName := output.GetSenderName(msg, ctx.Handles)
	timeStr := output.FormatTimestamp(msg.FormattedDate, "time")
	reactions := ctx.Reactions[msg.GUID]

	msgData := output.CreateMessageTemplateData(
		msg, senderName, timeStr, true, true, reactions,
	)

	messageTemplate := `[{{.Timestamp}}] {{.Sender}}: {{.Text}}{{if .Reactions}} {{range .Reactions}}{{.ReactionEmoji}}{{end}}{{end}}{{if .Attachments}}
  Attachments: {{range $i, $a := .Attachments}}{{if $i}}, {{end}}{{$a.Filename}}{{end}}{{end}}`

	tmpl, err := template.New("message").Parse(messageTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, msgData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// groupMessagesByDate groups messages by date
func (t *TextPlugin) groupMessagesByDate(messages []models.Message) map[string][]models.Message {
	grouped := make(map[string][]models.Message)

	for _, msg := range messages {
		// Skip empty messages
		if msg.Text == nil || strings.TrimSpace(*msg.Text) == "" {
			continue
		}

		dateKey := msg.FormattedDate.Format("2006-01-02")
		grouped[dateKey] = append(grouped[dateKey], msg)
	}

	return grouped
}

// ValidateConfig validates the text plugin configuration
func (t *TextPlugin) ValidateConfig(config *models.BookConfig) error {
	// Call base validation
	return t.BasePlugin.ValidateConfig(config)
}

// GetRequiredTemplates returns the list of template files this plugin needs
func (t *TextPlugin) GetRequiredTemplates() []string {
	return []string{
		"header.txt",
		"date-separator.txt",
		"message.txt",
	}
}
