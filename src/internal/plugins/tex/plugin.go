package tex

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"threadbound/internal/models"
	"threadbound/internal/output"
	"threadbound/internal/texgen"
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
	// Create a temporary database connection for the TeX generator
	db, err := sql.Open("sqlite3", ctx.Config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create TeX generator
	generator := texgen.New(ctx.Config, db)

	// Convert reactions to the format expected by the generator
	reactions := make(map[string][]models.Reaction)
	for guid, reactionList := range ctx.Reactions {
		reactions[guid] = reactionList
	}

	// Generate the TeX content
	texContent := generator.GenerateBook(ctx.Messages, ctx.Handles, reactions)

	return []byte(texContent), nil
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