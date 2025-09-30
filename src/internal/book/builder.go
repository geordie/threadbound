package book

import (
	"fmt"
	"os"

	"threadbound/internal/attachments"
	"threadbound/internal/database"
	"threadbound/internal/models"
	"threadbound/internal/output"
	_ "threadbound/internal/plugins" // Import to register plugins
)

// Builder orchestrates the book generation process
type Builder struct {
	config *models.BookConfig
	db     *database.DB
}

// New creates a new book builder
func New(config *models.BookConfig) (*Builder, error) {
	db, err := database.New(config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Builder{
		config: config,
		db:     db,
	}, nil
}

// Close closes the database connection
func (b *Builder) Close() error {
	return b.db.Close()
}

// Generate creates the book using the default output format (TeX)
func (b *Builder) Generate() error {
	return b.GenerateWithFormat("tex")
}

// GenerateWithFormat creates the book using the specified output plugin
func (b *Builder) GenerateWithFormat(format string) error {
	fmt.Println("📱 Extracting messages from database...")

	// Get all messages
	messages, err := b.db.GetMessages()
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	if len(messages) == 0 {
		return fmt.Errorf("no messages found in database")
	}

	fmt.Printf("✅ Found %d messages\n", len(messages))

	// Get handles (contacts)
	handles, err := b.db.GetHandles()
	if err != nil {
		return fmt.Errorf("failed to get handles: %w", err)
	}

	fmt.Printf("👥 Found %d contacts\n", len(handles))

	// Get reactions
	fmt.Println("👍 Loading message reactions...")
	reactions, err := b.db.GetReactions(handles)
	if err != nil {
		return fmt.Errorf("failed to get reactions: %w", err)
	}

	fmt.Printf("❤️ Found reactions for %d messages\n", len(reactions))

	// Process attachments for messages that have them
	fmt.Println("📎 Processing attachments...")
	err = b.processAttachments(messages)
	if err != nil {
		return fmt.Errorf("failed to process attachments: %w", err)
	}

	// Get book statistics
	stats, err := b.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	// Create generation context
	ctx := output.CreateContext(messages, handles, reactions, b.config, stats)

	// Generate using plugin system
	fmt.Printf("📝 Generating %s output...\n", format)
	generator := output.New()
	data, filename, err := generator.Generate(format, ctx)
	if err != nil {
		return fmt.Errorf("failed to generate %s: %w", format, err)
	}

	// Write to file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("✅ Generated book: %s\n", filename)
	return nil
}

// processAttachments loads attachment data for messages
func (b *Builder) processAttachments(messages []models.Message) error {
	processor := attachments.New(b.config)
	attachmentCount := 0
	imageCount := 0

	for i := range messages {
		if !messages[i].HasAttachments {
			continue
		}

		attachmentList, err := b.db.GetAttachmentsForMessage(messages[i].ID)
		if err != nil {
			return fmt.Errorf("failed to get attachments for message %d: %w", messages[i].ID, err)
		}

		// Process each attachment
		for j := range attachmentList {
			att := &attachmentList[j]

			// Try to find the attachment file
			err := processor.ProcessAttachment(att)
			if err != nil {
				fmt.Printf("⚠️  %v\n", err)
				continue
			}

			attachmentCount++

			// If it's an image, process it for the book
			if processor.IsImageFile(att) && b.config.IncludeImages {
				err := processor.ProcessImage(att)
				if err != nil {
					fmt.Printf("⚠️  Failed to process image %s: %v\n", *att.Filename, err)
				} else {
					imageCount++
				}
			}
		}

		messages[i].Attachments = attachmentList
	}

	fmt.Printf("✅ Processed %d attachments (%d images)\n", attachmentCount, imageCount)
	return nil
}

// GetStats returns statistics about the messages
func (b *Builder) GetStats() (*models.BookStats, error) {
	messages, err := b.db.GetMessages()
	if err != nil {
		return nil, err
	}

	handles, err := b.db.GetHandles()
	if err != nil {
		return nil, err
	}

	stats := &models.BookStats{
		TotalMessages:    len(messages),
		TotalContacts:    len(handles),
		AttachmentCount:  0,
	}

	// Count messages with text
	textMessages := 0
	for _, msg := range messages {
		if msg.Text != nil && *msg.Text != "" {
			textMessages++
		}
		if msg.HasAttachments {
			stats.AttachmentCount++
		}
	}

	stats.TextMessages = textMessages

	// Find date range
	if len(messages) > 0 {
		stats.StartDate = messages[0].FormattedDate
		stats.EndDate = messages[len(messages)-1].FormattedDate
	}

	return stats, nil
}