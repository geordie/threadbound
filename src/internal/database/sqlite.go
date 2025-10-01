package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"threadbound/internal/models"
	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetConnection returns the underlying sql.DB connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// GetMessages retrieves all messages ordered by date, excluding reactions
func (db *DB) GetMessages() ([]models.Message, error) {
	query := `
		SELECT
			m.ROWID, m.guid, m.text, m.date, m.date_read, m.date_delivered,
			m.is_from_me, m.is_delivered, m.is_read, m.handle_id,
			m.cache_has_attachments, m.subject, m.is_audio_message,
			m.associated_message_guid, m.associated_message_type, m.item_type
		FROM message m
		WHERE m.associated_message_guid IS NULL
		ORDER BY m.date ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID, &msg.GUID, &msg.Text, &msg.Date, &msg.DateRead, &msg.DateDelivered,
			&msg.IsFromMe, &msg.IsDelivered, &msg.IsRead, &msg.HandleID,
			&msg.HasAttachments, &msg.Subject, &msg.IsAudioMessage,
			&msg.AssociatedMessageGUID, &msg.AssociatedMessageType, &msg.ItemType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Convert Apple's timestamp to Go time
		// Apple uses seconds since January 1, 2001
		appleEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		msg.FormattedDate = appleEpoch.Add(time.Duration(msg.Date) * time.Nanosecond)

		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// GetAttachmentsForMessage retrieves attachments for a specific message
func (db *DB) GetAttachmentsForMessage(messageID int) ([]models.Attachment, error) {
	query := `
		SELECT
			a.ROWID, a.guid, a.filename, a.uti, a.mime_type,
			a.total_bytes, a.is_sticker, a.is_outgoing
		FROM attachment a
		JOIN message_attachment_join maj ON a.ROWID = maj.attachment_id
		WHERE maj.message_id = ?
		ORDER BY a.ROWID ASC
	`

	rows, err := db.conn.Query(query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments: %w", err)
	}
	defer rows.Close()

	var attachments []models.Attachment
	for rows.Next() {
		var att models.Attachment
		err := rows.Scan(
			&att.ID, &att.GUID, &att.Filename, &att.UTI, &att.MimeType,
			&att.TotalBytes, &att.IsSticker, &att.IsOutgoing,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, att)
	}

	return attachments, rows.Err()
}

// GetHandles retrieves all contact handles
func (db *DB) GetHandles(contactNames map[string]string) (map[int]models.Handle, error) {
	query := `
		SELECT ROWID, service, id, country
		FROM handle
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query handles: %w", err)
	}
	defer rows.Close()

	handles := make(map[int]models.Handle)
	for rows.Next() {
		var handle models.Handle
		err := rows.Scan(&handle.ID, &handle.Service, &handle.Contact, &handle.Country)
		if err != nil {
			return nil, fmt.Errorf("failed to scan handle: %w", err)
		}

		// Check if there's a custom name mapping for this contact
		if contactNames != nil {
			if customName, exists := contactNames[handle.Contact]; exists {
				handle.DisplayName = customName
			} else {
				// Use default display name logic
				handle.DisplayName = handle.Contact
				if handle.Service == "iMessage" {
					// Extract name from email if it's an email
					if len(handle.Contact) > 0 && handle.Contact[0] != '+' {
						handle.DisplayName = handle.Contact
					}
				}
			}
		} else {
			// Use default display name logic
			handle.DisplayName = handle.Contact
			if handle.Service == "iMessage" {
				// Extract name from email if it's an email
				if len(handle.Contact) > 0 && handle.Contact[0] != '+' {
					handle.DisplayName = handle.Contact
				}
			}
		}

		handles[handle.ID] = handle
	}

	return handles, rows.Err()
}

// GetReactions retrieves all reactions keyed by the original message GUID
func (db *DB) GetReactions(handles map[int]models.Handle) (map[string][]models.Reaction, error) {
	query := `
		SELECT
			m.associated_message_guid, m.associated_message_type, m.date,
			m.handle_id, m.is_from_me
		FROM message m
		WHERE m.associated_message_guid IS NOT NULL
		ORDER BY m.date ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query reactions: %w", err)
	}
	defer rows.Close()

	reactions := make(map[string][]models.Reaction)
	for rows.Next() {
		var associatedGUID string
		var reactionType int
		var date int64
		var handleID *int
		var isFromMe bool

		err := rows.Scan(&associatedGUID, &reactionType, &date, &handleID, &isFromMe)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reaction: %w", err)
		}

		// Extract the actual UUID from the associated_message_guid field
		// Format is like "p:0/BB33CAD3-02F1-4226-9AB8-3D0BF5A9D1E1"
		originalGUID := associatedGUID
		if len(associatedGUID) > 3 && (associatedGUID[:3] == "p:0" || associatedGUID[:3] == "p:1") {
			slashIndex := strings.Index(associatedGUID, "/")
			if slashIndex != -1 && slashIndex+1 < len(associatedGUID) {
				originalGUID = associatedGUID[slashIndex+1:]
			}
		}

		// Convert Apple's timestamp to Go time
		appleEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		timestamp := appleEpoch.Add(time.Duration(date) * time.Nanosecond)

		// Determine sender name
		var senderName string
		if isFromMe {
			senderName = "Me"
		} else {
			if handleID != nil {
				if handle, exists := handles[*handleID]; exists {
					senderName = handle.DisplayName
				} else {
					senderName = "Unknown"
				}
			} else {
				senderName = "Unknown"
			}
		}

		reaction := models.Reaction{
			Type:          reactionType,
			SenderName:    senderName,
			Timestamp:     timestamp,
			ReactionEmoji: reactionTypeToEmoji(reactionType),
		}

		reactions[originalGUID] = append(reactions[originalGUID], reaction)
	}

	return reactions, rows.Err()
}

// reactionTypeToEmoji converts iMessage reaction types to Unicode emoji
func reactionTypeToEmoji(reactionType int) string {
	switch reactionType {
	case 2000:
		return "❤️" // Love
	case 2001:
		return "👍" // Like/Thumbs up
	case 2002:
		return "👎" // Dislike/Thumbs down
	case 2003:
		return "😂" // Laugh/Ha ha
	case 2004:
		return "‼️" // Emphasize/!!
	case 2005:
		return "❤️" // Love (alternative)
	case 2006:
		return "❓" // Question/?
	default:
		return "❤️" // Default fallback
	}
}

// GetChatInfo retrieves basic chat information
func (db *DB) GetChatInfo() (*models.Handle, error) {
	query := `
		SELECT COUNT(*) as message_count
		FROM message
		WHERE text IS NOT NULL AND text != ''
	`

	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat info: %w", err)
	}

	// For now, return basic info - could be enhanced
	return &models.Handle{
		DisplayName: fmt.Sprintf("Group Chat (%d messages)", count),
	}, nil
}