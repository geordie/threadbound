package models

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Message represents an iMessage from the database
type Message struct {
	ID                    int       `db:"ROWID"`
	GUID                  string    `db:"guid"`
	Text                  *string   `db:"text"`
	Date                  int64     `db:"date"`
	DateRead              *int64    `db:"date_read"`
	DateDelivered         *int64    `db:"date_delivered"`
	IsFromMe              bool      `db:"is_from_me"`
	IsDelivered           bool      `db:"is_delivered"`
	IsRead                bool      `db:"is_read"`
	HandleID              *int      `db:"handle_id"`
	HasAttachments        bool      `db:"cache_has_attachments"`
	Subject               *string   `db:"subject"`
	IsAudioMessage        bool      `db:"is_audio_message"`
	AssociatedMessageGUID *string   `db:"associated_message_guid"`
	AssociatedMessageType int       `db:"associated_message_type"`
	ItemType              int       `db:"item_type"`

	// Computed fields
	FormattedDate   time.Time
	SenderName      string
	Attachments     []Attachment
	Reactions       []Reaction
}

// Attachment represents a file attachment
type Attachment struct {
	ID          int     `db:"ROWID"`
	GUID        string  `db:"guid"`
	Filename    *string `db:"filename"`
	UTI         *string `db:"uti"`
	MimeType    *string `db:"mime_type"`
	TotalBytes  int64   `db:"total_bytes"`
	IsSticker   bool    `db:"is_sticker"`
	IsOutgoing  bool    `db:"is_outgoing"`

	// Computed fields
	LocalPath   string
	ProcessedPath string
}

// Reaction represents a message reaction/tapback
type Reaction struct {
	Type          int
	SenderName    string
	Timestamp     time.Time
	ReactionEmoji string
}

// Handle represents a contact/phone number
type Handle struct {
	ID      int    `db:"ROWID"`
	Service string `db:"service"`
	Contact string `db:"id"`
	Country string `db:"country"`

	// Computed fields
	DisplayName string
}

// BookConfig holds configuration for book generation
type BookConfig struct {
	Title           string `yaml:"title"`
	Author          string `yaml:"author"`
	DatabasePath    string `yaml:"database_path"`
	AttachmentsPath string `yaml:"attachments_path"`
	OutputPath      string `yaml:"output_path"`
	TemplateDir     string `yaml:"template_dir"`
	IncludeImages   bool   `yaml:"include_images"`
	IncludePreviews bool   `yaml:"include_previews"`
	PageWidth       string `yaml:"page_width"`
	PageHeight      string `yaml:"page_height"`
}

// LoadConfigFromFile loads configuration from a YAML file
func LoadConfigFromFile(configPath string) (*BookConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config BookConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &config, nil
}

// GetDefaultConfig returns a BookConfig with default values
func GetDefaultConfig() *BookConfig {
	return &BookConfig{
		Title:           "Our Messages",
		Author:          "",
		DatabasePath:    "chat.db",
		AttachmentsPath: "Attachments",
		OutputPath:      "book.md",
		TemplateDir:     "templates",
		IncludeImages:   true,
		IncludePreviews: true,
		PageWidth:       "5.5in",
		PageHeight:      "8.5in",
	}
}

// BookStats holds statistics about the book content
type BookStats struct {
	TotalMessages   int
	TextMessages    int
	TotalContacts   int
	AttachmentCount int
	StartDate       time.Time
	EndDate         time.Time
}

// PDFInfo holds information about a generated PDF
type PDFInfo struct {
	FilePath   string
	FileSize   int64
	CreatedAt  time.Time
	PageWidth  string
	PageHeight string
}