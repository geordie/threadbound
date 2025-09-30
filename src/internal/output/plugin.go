package output

import (
	"threadbound/internal/models"
)

// PluginCapabilities defines what features a plugin supports
type PluginCapabilities struct {
	SupportsImages     bool   // Can handle image attachments
	SupportsAttachments bool   // Can handle non-image attachments
	SupportsReactions  bool   // Can display message reactions
	SupportsURLPreviews bool   // Can display URL previews
	RequiresTemplates  bool   // Needs template files to function
	SupportsPagination bool   // Can handle page breaks/pagination
}

// OutputPlugin defines the interface that all output format plugins must implement
type OutputPlugin interface {
	// Name returns the human-readable name of the plugin
	Name() string

	// ID returns a unique identifier for the plugin (used in CLI and config)
	ID() string

	// FileExtension returns the file extension for this output format (without dot)
	FileExtension() string

	// Description returns a brief description of what this plugin generates
	Description() string

	// GetCapabilities returns what features this plugin supports
	GetCapabilities() PluginCapabilities

	// Generate processes the message data and returns the formatted output
	Generate(ctx *GenerationContext) ([]byte, error)

	// ValidateConfig checks if the plugin configuration is valid
	ValidateConfig(config *models.BookConfig) error

	// GetRequiredTemplates returns a list of template files this plugin needs
	GetRequiredTemplates() []string
}

// GenerationContext contains all the data and configuration needed for output generation
type GenerationContext struct {
	Messages      []models.Message
	Handles       map[int]models.Handle
	Reactions     map[string][]models.Reaction
	Config        *models.BookConfig
	URLThumbnails map[string]*URLThumbnail
	Stats         *models.BookStats
}

// URLThumbnail represents a processed URL preview
type URLThumbnail struct {
	URL           string
	Title         string
	Description   string
	ThumbnailPath string // Path to thumbnail image
	ImagePath     string // Alias for ThumbnailPath (deprecated)
	Success       bool
	Error         string
}

// PluginError represents an error that occurred during plugin execution
type PluginError struct {
	PluginID string
	Message  string
	Cause    error
}

func (e *PluginError) Error() string {
	if e.Cause != nil {
		return e.PluginID + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.PluginID + ": " + e.Message
}

// TemplateData provides common data structures that plugins can use for templating
type TemplateData struct {
	Title      string
	Author     string
	Date       string
	PageWidth  string
	PageHeight string
	Stats      *models.BookStats
}

// MessageTemplateData provides message-specific data for templating
type MessageTemplateData struct {
	Text          string
	Timestamp     string
	Sender        string
	IsFromMe      bool
	ShowSender    bool
	ShowTimestamp bool
	Reactions     []models.Reaction
	Attachments   []models.Attachment
	HasURL        bool
	URLPreviews   []*URLThumbnail
}