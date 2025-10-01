package html

import (
	"fmt"
	"html/template"
	"strings"

	"threadbound/internal/models"
	"threadbound/internal/output"
)

// HTMLPlugin implements the OutputPlugin interface for HTML generation
type HTMLPlugin struct {
	*output.BasePlugin
}

// NewHTMLPlugin creates a new HTML plugin instance
func NewHTMLPlugin() *HTMLPlugin {
	capabilities := output.PluginCapabilities{
		SupportsImages:      true,
		SupportsAttachments: true,
		SupportsReactions:   true,
		SupportsURLPreviews: true,
		RequiresTemplates:   false,
		SupportsPagination:  false,
	}

	base := output.NewBasePlugin(
		"html",
		"HTML Book",
		"Generate HTML book with responsive design",
		"html",
		capabilities,
	)

	return &HTMLPlugin{
		BasePlugin: base,
	}
}

// Generate creates an HTML book from the message data
func (h *HTMLPlugin) Generate(ctx *output.GenerationContext) ([]byte, error) {
	templateData := h.prepareTemplateData(ctx)

	htmlContent, err := h.generateHTML(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}

	return []byte(htmlContent), nil
}

// HTMLTemplateData contains all data needed for HTML generation
type HTMLTemplateData struct {
	*output.TemplateData
	MessagesByDate map[string][]MessageData
}

// MessageData represents a message for HTML templating
type MessageData struct {
	*output.MessageTemplateData
	FormattedDate string
	DateKey       string
}

// prepareTemplateData organizes the data for HTML templating
func (h *HTMLPlugin) prepareTemplateData(ctx *output.GenerationContext) *HTMLTemplateData {
	baseData := ctx.GetTemplateData()

	// Group messages by date
	messagesByDate := make(map[string][]MessageData)

	for _, msg := range ctx.Messages {
		if msg.Text == nil || strings.TrimSpace(*msg.Text) == "" {
			continue
		}

		dateKey := msg.FormattedDate.Format("2006-01-02")
		senderName := output.GetSenderNameWithConfig(msg, ctx.Handles, ctx.Config)
		timeStr := output.FormatTimestamp(msg.FormattedDate, "time")

		// Get reactions for this message
		reactions := ctx.Reactions[msg.GUID]

		msgData := MessageData{
			MessageTemplateData: output.CreateMessageTemplateData(
				msg, senderName, timeStr, true, true, reactions,
			),
			FormattedDate: msg.FormattedDate.Format("January 2, 2006"),
			DateKey:       dateKey,
		}

		messagesByDate[dateKey] = append(messagesByDate[dateKey], msgData)
	}

	return &HTMLTemplateData{
		TemplateData:   baseData,
		MessagesByDate: messagesByDate,
	}
}

// generateHTML creates the HTML content using embedded templates
func (h *HTMLPlugin) generateHTML(data *HTMLTemplateData) (string, error) {
	tmpl := template.New("book")

	// Define the main template
	mainTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 40px; text-align: center; }
        .header h1 { margin: 0; font-size: 2.5em; }
        .header p { margin: 10px 0 0 0; opacity: 0.9; }
        .content { padding: 20px; }
        .date-section { margin: 30px 0; }
        .date-header { font-size: 1.2em; font-weight: bold; color: #333; margin-bottom: 15px; padding-bottom: 5px; border-bottom: 2px solid #eee; }
        .message { margin: 10px 0; display: flex; }
        .message.from-me { justify-content: flex-end; }
        .message-bubble { max-width: 70%; padding: 12px 16px; border-radius: 18px; position: relative; }
        .message.from-me .message-bubble { background: #007AFF; color: white; }
        .message:not(.from-me) .message-bubble { background: #E5E5EA; color: black; }
        .message-meta { font-size: 0.8em; opacity: 0.7; margin-top: 4px; }
        .reactions { margin-top: 8px; }
        .reaction { display: inline-block; background: rgba(0,0,0,0.1); padding: 2px 6px; border-radius: 10px; font-size: 0.8em; margin-right: 4px; }
        .attachments { margin-top: 8px; }
        .attachment { padding: 8px; background: rgba(0,0,0,0.05); border-radius: 8px; margin: 4px 0; }
        .stats { background: #f8f9fa; padding: 20px; margin: 20px 0; border-radius: 8px; }
        .stats h3 { margin-top: 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
            {{if .Author}}<p>by {{.Author}}</p>{{end}}
            <p>Generated on {{.Date}}</p>
        </div>

        {{if .Stats}}
        <div class="stats">
            <h3>📊 Book Statistics</h3>
            <p><strong>Messages:</strong> {{.Stats.TotalMessages}} ({{.Stats.TextMessages}} with text)</p>
            <p><strong>Contacts:</strong> {{.Stats.TotalContacts}}</p>
            <p><strong>Attachments:</strong> {{.Stats.AttachmentCount}}</p>
        </div>
        {{end}}

        <div class="content">
            {{range $dateKey, $messages := .MessagesByDate}}
            <div class="date-section">
                <div class="date-header">{{(index $messages 0).FormattedDate}}</div>
                {{range $messages}}
                <div class="message{{if .IsFromMe}} from-me{{end}}">
                    <div class="message-bubble">
                        {{.Text}}
                        <div class="message-meta">
                            {{if not .IsFromMe}}{{.Sender}} • {{end}}{{.Timestamp}}
                        </div>
                        {{if .Reactions}}
                        <div class="reactions">
                            {{range .Reactions}}
                            <span class="reaction">{{.ReactionEmoji}} {{.SenderName}}</span>
                            {{end}}
                        </div>
                        {{end}}
                        {{if .Attachments}}
                        <div class="attachments">
                            {{range .Attachments}}
                            <div class="attachment">📎 {{.Filename}}</div>
                            {{end}}
                        </div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>`

	tmpl, err := tmpl.Parse(mainTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ValidateConfig validates the HTML plugin configuration
func (h *HTMLPlugin) ValidateConfig(config *models.BookConfig) error {
	// Call base validation
	return h.BasePlugin.ValidateConfig(config)
}