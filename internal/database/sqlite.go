package database

import (
	"database/sql"
	"fmt"
	"time"

	"imessages-book/internal/models"
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

// GetMessages retrieves all messages ordered by date
func (db *DB) GetMessages() ([]models.Message, error) {
	query := `
		SELECT
			m.ROWID, m.guid, m.text, m.date, m.date_read, m.date_delivered,
			m.is_from_me, m.is_delivered, m.is_read, m.handle_id,
			m.cache_has_attachments, m.subject, m.is_audio_message
		FROM message m
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
func (db *DB) GetHandles() (map[int]models.Handle, error) {
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

		// Simple display name logic - could be enhanced
		handle.DisplayName = handle.Contact
		if handle.Service == "iMessage" {
			// Extract name from email if it's an email
			if len(handle.Contact) > 0 && handle.Contact[0] != '+' {
				handle.DisplayName = handle.Contact
			}
		}

		handles[handle.ID] = handle
	}

	return handles, rows.Err()
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