package book

import (
	"fmt"
	"os"

	"imessages-book/internal/attachments"
	"imessages-book/internal/database"
	"imessages-book/internal/markdown"
	"imessages-book/internal/models"
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

// Generate creates the markdown book
func (b *Builder) Generate() error {
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

	// Process attachments for messages that have them
	fmt.Println("📎 Processing attachments...")
	err = b.processAttachments(messages)
	if err != nil {
		return fmt.Errorf("failed to process attachments: %w", err)
	}

	// Generate markdown
	fmt.Println("📝 Generating markdown...")
	generator := markdown.New(b.config)
	markdownContent := generator.GenerateBook(messages, handles)

	// Write to file
	err = os.WriteFile(b.config.OutputPath, []byte(markdownContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	fmt.Printf("✅ Generated book: %s\n", b.config.OutputPath)
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