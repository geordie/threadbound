package models

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ReactionType represents the type of message reaction/tapback
type ReactionType int

const (
	ReactionNone     ReactionType = 0
	ReactionLoved    ReactionType = 2000
	ReactionLiked    ReactionType = 2001
	ReactionUnknown  ReactionType = 2002
	ReactionLaughed  ReactionType = 2003
	ReactionUnknown2 ReactionType = 2004
	ReactionCustom   ReactionType = 2006 // Custom emoji reactions
)

// String returns a human-readable representation of the reaction type
func (rt ReactionType) String() string {
	switch rt {
	case ReactionLoved:
		return "Loved"
	case ReactionLiked:
		return "Liked"
	case ReactionLaughed:
		return "Laughed"
	case ReactionCustom:
		return "Reacted"
	default:
		return "Unknown"
	}
}

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

	// Threading fields
	ReplyToGUID            *string `db:"reply_to_guid"`
	ThreadOriginatorGUID   *string `db:"thread_originator_guid"`
	ThreadOriginatorPart   *string `db:"thread_originator_part"`

	// Computed fields
	FormattedDate   time.Time
	SenderName      string
	Attachments     []Attachment
	Reactions       []Reaction

	// Threading computed fields
	ReplyToMessage     *Message  // Populated when loading thread context
	ThreadReplies      []Message // Messages that reply to this message
	ThreadOriginator   *Message  // The original message that started this thread
	IsReaction         bool      // True if this is a reaction (tapback)
	ReactionType       ReactionType
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

// Message helper methods for threading

// IsReply returns true if this message is a reply to another message
func (m *Message) IsReply() bool {
	return m.ReplyToGUID != nil && *m.ReplyToGUID != ""
}

// IsInThread returns true if this message is part of a thread
func (m *Message) IsInThread() bool {
	return m.ThreadOriginatorGUID != nil && *m.ThreadOriginatorGUID != ""
}

// IsThreadOriginator returns true if this message started a thread
func (m *Message) IsThreadOriginator() bool {
	return m.IsInThread() && *m.ThreadOriginatorGUID == m.GUID
}

// GetReactionType returns the reaction type for this message
func (m *Message) GetReactionType() ReactionType {
	return ReactionType(m.AssociatedMessageType)
}

// IsReactionMessage returns true if this message is a reaction/tapback
func (m *Message) IsReactionMessage() bool {
	return m.AssociatedMessageGUID != nil && m.AssociatedMessageType != 0
}

// GetReactionTarget returns the GUID of the message this reaction targets
func (m *Message) GetReactionTarget() string {
	if m.AssociatedMessageGUID != nil {
		return *m.AssociatedMessageGUID
	}
	return ""
}

// HasReplies returns true if this message has direct replies
func (m *Message) HasReplies() bool {
	return len(m.ThreadReplies) > 0
}

// GetThreadDepth returns how deep this message is in a thread (0 = original, 1 = first reply, etc.)
func (m *Message) GetThreadDepth() int {
	if !m.IsInThread() {
		return 0
	}
	if m.IsThreadOriginator() {
		return 0
	}
	// For now, assume direct replies are depth 1
	// Could be enhanced to calculate actual depth by traversing the chain
	return 1
}