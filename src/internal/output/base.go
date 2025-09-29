package output

import (
	"strings"
	"time"

	"threadbound/internal/models"
)

// BasePlugin provides common functionality that plugins can embed
type BasePlugin struct {
	id          string
	name        string
	description string
	extension   string
	capabilities PluginCapabilities
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(id, name, description, extension string, capabilities PluginCapabilities) *BasePlugin {
	return &BasePlugin{
		id:           id,
		name:         name,
		description:  description,
		extension:    extension,
		capabilities: capabilities,
	}
}

// ID returns the plugin ID
func (b *BasePlugin) ID() string {
	return b.id
}

// Name returns the plugin name
func (b *BasePlugin) Name() string {
	return b.name
}

// Description returns the plugin description
func (b *BasePlugin) Description() string {
	return b.description
}

// FileExtension returns the file extension
func (b *BasePlugin) FileExtension() string {
	return b.extension
}

// GetCapabilities returns the plugin capabilities
func (b *BasePlugin) GetCapabilities() PluginCapabilities {
	return b.capabilities
}

// ValidateConfig provides basic configuration validation
func (b *BasePlugin) ValidateConfig(config *models.BookConfig) error {
	// Basic validation - can be overridden by specific plugins
	if config.Title == "" {
		config.Title = "Untitled Book"
	}
	return nil
}

// GetRequiredTemplates returns an empty slice by default
func (b *BasePlugin) GetRequiredTemplates() []string {
	return []string{}
}

// Helper functions for plugins

// GroupMessagesByDate groups messages by date for easier processing
func GroupMessagesByDate(messages []models.Message) map[string][]models.Message {
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

// GroupMessagesByMonth groups messages by month for chapter organization
func GroupMessagesByMonth(messages []models.Message) map[string][]models.Message {
	grouped := make(map[string][]models.Message)

	for _, msg := range messages {
		// Skip empty messages
		if msg.Text == nil || strings.TrimSpace(*msg.Text) == "" {
			continue
		}

		monthKey := msg.FormattedDate.Format("2006-01")
		grouped[monthKey] = append(grouped[monthKey], msg)
	}

	return grouped
}

// FormatTimestamp formats a timestamp for display
func FormatTimestamp(t time.Time, format string) string {
	switch format {
	case "time":
		return t.Format("3:04 PM")
	case "date":
		return t.Format("Monday, January 2, 2006")
	case "month":
		return t.Format("January 2006")
	case "iso":
		return t.Format("2006-01-02T15:04:05Z07:00")
	default:
		return t.Format("January 2, 2006 3:04 PM")
	}
}

// GetSenderName determines the display name for a message sender
func GetSenderName(msg models.Message, handles map[int]models.Handle) string {
	if msg.IsFromMe {
		return "Me"
	}

	if msg.HandleID != nil {
		if handle, exists := handles[*msg.HandleID]; exists {
			return handle.DisplayName
		}
	}

	return "Unknown"
}

// IsImageFile checks if a filename represents an image
func IsImageFile(filename string) bool {
	if filename == "" {
		return false
	}

	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".heic"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}