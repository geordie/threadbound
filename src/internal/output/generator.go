package output

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"threadbound/internal/models"
)

// Generator coordinates the output generation process using plugins
type Generator struct {
	registry *Registry
}

// New creates a new output generator
func New() *Generator {
	return &Generator{
		registry: globalRegistry,
	}
}

// Generate produces output using the specified plugin
func (g *Generator) Generate(pluginID string, ctx *GenerationContext) ([]byte, string, error) {
	plugin, err := g.registry.Get(pluginID)
	if err != nil {
		return nil, "", &PluginError{
			PluginID: pluginID,
			Message:  "plugin not found",
			Cause:    err,
		}
	}

	// Validate plugin configuration
	if err := plugin.ValidateConfig(ctx.Config); err != nil {
		return nil, "", &PluginError{
			PluginID: pluginID,
			Message:  "configuration validation failed",
			Cause:    err,
		}
	}

	// Generate output
	data, err := plugin.Generate(ctx)
	if err != nil {
		return nil, "", &PluginError{
			PluginID: pluginID,
			Message:  "generation failed",
			Cause:    err,
		}
	}

	// Determine output filename
	filename := g.generateFilename(ctx.Config.OutputPath, plugin.FileExtension())

	return data, filename, nil
}

// generateFilename creates an appropriate filename for the given plugin
func (g *Generator) generateFilename(basePath, extension string) string {
	// If basePath already has the correct extension, use it as-is
	if strings.HasSuffix(basePath, "."+extension) {
		return basePath
	}

	// Remove existing extension and add the plugin's extension
	base := strings.TrimSuffix(basePath, filepath.Ext(basePath))
	return base + "." + extension
}

// CreateContext creates a GenerationContext from the provided data
func CreateContext(messages []models.Message, handles map[int]models.Handle,
	reactions map[string][]models.Reaction, config *models.BookConfig,
	stats *models.BookStats) *GenerationContext {

	return &GenerationContext{
		Messages:      messages,
		Handles:       handles,
		Reactions:     reactions,
		Config:        config,
		URLThumbnails: make(map[string]*URLThumbnail), // Will be populated later
		Stats:         stats,
	}
}

// GetTemplateData creates common template data from the generation context
func (ctx *GenerationContext) GetTemplateData() *TemplateData {
	return &TemplateData{
		Title:      ctx.Config.Title,
		Author:     ctx.Config.Author,
		Date:       time.Now().Format("January 2, 2006"),
		PageWidth:  ctx.Config.PageWidth,
		PageHeight: ctx.Config.PageHeight,
		Stats:      ctx.Stats,
	}
}

// CreateMessageTemplateData creates template data for a specific message
func CreateMessageTemplateData(msg models.Message, senderName string, timeStr string,
	showSender, showTimestamp bool, reactions []models.Reaction) *MessageTemplateData {

	text := ""
	if msg.Text != nil {
		text = *msg.Text
	}

	return &MessageTemplateData{
		Text:          text,
		Timestamp:     timeStr,
		Sender:        senderName,
		IsFromMe:      msg.IsFromMe,
		ShowSender:    showSender,
		ShowTimestamp: showTimestamp,
		Reactions:     reactions,
		Attachments:   msg.Attachments,
		HasURL:        false, // Will be set by URL processing
		URLPreviews:   make([]*URLThumbnail, 0),
	}
}

// ValidatePluginExists checks if a plugin ID exists and returns a helpful error if not
func ValidatePluginExists(pluginID string) error {
	if !Exists(pluginID) {
		availableIDs := GetIDs()
		if len(availableIDs) > 0 {
			return fmt.Errorf("unknown output format '%s'. Available formats: %s",
				pluginID, strings.Join(availableIDs, ", "))
		}
		return fmt.Errorf("unknown output format '%s' and no plugins are registered", pluginID)
	}
	return nil
}