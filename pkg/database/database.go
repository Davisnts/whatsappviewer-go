package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// Database represents an open WhatsApp SQLite database
type Database struct {
	db       *sql.DB
	path     string
	isModern bool
}

// OpenDatabase opens a WhatsApp database and detects whether it is a modern or legacy schema
func OpenDatabase(path string) (*Database, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("database file not found: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	wdb := &Database{
		db:   db,
		path: path,
	}

	// Detect schema based on which table actually contains data
	var modernTable, legacyTable int
	db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='message'").Scan(&modernTable)
	db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'").Scan(&legacyTable)

	var modernRows, legacyRows int
	if modernTable > 0 {
		db.QueryRow("SELECT COUNT(*) FROM message").Scan(&modernRows)
	}
	if legacyTable > 0 {
		db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&legacyRows)
	}

	if modernRows > 0 {
		wdb.isModern = true
	} else if legacyRows > 0 {
		wdb.isModern = false
	} else if modernTable > 0 {
		wdb.isModern = true
	} else {
		wdb.isModern = false
	}

	return wdb, nil
}

func (d *Database) hasColumn(tableName, colName string) bool {
	rows, err := d.db.Query(fmt.Sprintf("PRAGMA table_info('%s')", tableName))
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
			if strings.EqualFold(name, colName) {
				return true
			}
		}
	}
	return false
}

// Close closes the underlying SQLite database
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// IsModern returns true if the database uses the modern WhatsApp 2020+ schema
func (d *Database) IsModern() bool {
	return d.isModern
}

// GetChats returns all conversations in the database
func (d *Database) GetChats(contacts map[string]string) ([]Chat, error) {
	if d.isModern {
		return d.getChatsModern(contacts)
	}
	return d.getChatsLegacy(contacts)
}

func (d *Database) getChatsModern(contacts map[string]string) ([]Chat, error) {
	query := `
		SELECT j.raw_string, COALESCE(c.subject, ''), COALESCE(c.created_timestamp, 0), COALESCE(MAX(m.timestamp), 0)
		FROM chat c
		JOIN jid j ON j._id = c.jid_row_id
		LEFT JOIN message m ON m.chat_row_id = c._id
		WHERE c.hidden = 0
		GROUP BY j.raw_string, c.subject, c.created_timestamp
		ORDER BY MAX(m.timestamp) DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query modern chats: %w", err)
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var c Chat
		if err := rows.Scan(&c.JID, &c.Subject, &c.CreatedTimestamp, &c.LastMessageTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan chat row: %w", err)
		}
		if name, ok := contacts[c.JID]; ok {
			c.DisplayName = name
		}
		fromMe, total, _ := d.getCountsModern(c.JID)
		c.FromMeCount = fromMe
		c.TotalCount = total
		c.MessagesCount = total
		chats = append(chats, c)
	}

	return chats, nil
}

func (d *Database) getCountsModern(jid string) (fromMe, total int, err error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN m.from_me = 1 THEN 1 ELSE 0 END), 0)
		FROM message m
		JOIN chat c ON c._id = m.chat_row_id
		JOIN jid j ON j._id = c.jid_row_id
		WHERE j.raw_string = ?
	`
	err = d.db.QueryRow(query, jid).Scan(&total, &fromMe)
	return fromMe, total, err
}

func (d *Database) getChatsLegacy(contacts map[string]string) ([]Chat, error) {
	// Try chat_view first
	var hasChatView int
	d.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table', 'view') AND name='chat_view'").Scan(&hasChatView)

	var query string
	if hasChatView > 0 {
		query = `
			SELECT cv.raw_string_jid, COALESCE(cv.subject, ''), COALESCE(cv.created_timestamp, 0), COALESCE(MAX(m.timestamp), 0)
			FROM chat_view cv
			LEFT JOIN messages m ON m.key_remote_jid = cv.raw_string_jid
			WHERE cv.hidden = 0
			GROUP BY cv.raw_string_jid, cv.subject, cv.created_timestamp
			ORDER BY MAX(m.timestamp) DESC
		`
	} else {
		query = `
			SELECT key_remote_jid, '', 0, COALESCE(MAX(timestamp), 0)
			FROM messages
			WHERE key_remote_jid IS NOT NULL AND key_remote_jid != ''
			GROUP BY key_remote_jid
			ORDER BY MAX(timestamp) DESC
		`
	}

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query legacy chats: %w", err)
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var c Chat
		if err := rows.Scan(&c.JID, &c.Subject, &c.CreatedTimestamp, &c.LastMessageTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan legacy chat: %w", err)
		}
		if name, ok := contacts[c.JID]; ok {
			c.DisplayName = name
		}
		fromMe, total, _ := d.getCountsLegacy(c.JID)
		c.FromMeCount = fromMe
		c.TotalCount = total
		c.MessagesCount = total
		chats = append(chats, c)
	}

	return chats, nil
}

func (d *Database) getCountsLegacy(jid string) (fromMe, total int, err error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN key_from_me = 1 THEN 1 ELSE 0 END), 0)
		FROM messages
		WHERE key_remote_jid = ?
	`
	err = d.db.QueryRow(query, jid).Scan(&total, &fromMe)
	return fromMe, total, err
}

// GetMessages retrieves messages for a given chat JID, populating quoted messages and contact names
func (d *Database) GetMessages(jid string, contacts map[string]string) ([]Message, error) {
	if d.isModern {
		return d.getMessagesModern(jid, contacts)
	}
	return d.getMessagesLegacy(jid, contacts)
}

func (d *Database) getMessagesModern(jid string, contacts map[string]string) ([]Message, error) {
	captionCol := "''"
	if d.hasColumn("message_media", "media_caption") {
		captionCol = "COALESCE(message_media.media_caption, '')"
	}

	query := fmt.Sprintf(`
		SELECT
			message.key_id,
			CAST(message.chat_row_id AS TEXT),
			COALESCE(message.from_me, 0),
			COALESCE(message.status, 0),
			COALESCE(message.text_data, ''),
			COALESCE(message.timestamp, 0),
			COALESCE(message_media.message_url, ''),
			COALESCE(message_media.mime_type, ''),
			COALESCE(message.message_type, 0),
			COALESCE(message_media.file_length, 0),
			COALESCE(message_media.media_name, ''),
			%s,
			COALESCE(message_media.media_duration, 0),`, captionCol) + `
			COALESCE(message_location.latitude, 0.0),
			COALESCE(message_location.longitude, 0.0),
			message_thumbnail.thumbnail,
			message_quoted_media.thumbnail,
			COALESCE(message_quoted.key_id, ''),
			COALESCE(message_link._id, 0),
			COALESCE(sender_jid.user, '')
		FROM message
		LEFT JOIN message_quoted ON message._id = message_quoted.message_row_id
		LEFT JOIN message_quoted_media ON message_quoted.message_row_id = message_quoted_media.message_row_id
		LEFT JOIN message_link ON message._id = message_link.message_row_id
		LEFT JOIN message_media ON message._id = message_media.message_row_id
		LEFT JOIN message_location ON message._id = message_location.message_row_id
		LEFT JOIN message_thumbnail ON message._id = message_thumbnail.message_row_id
		JOIN chat ON chat._id = message.chat_row_id
		JOIN jid ON jid._id = chat.jid_row_id
		LEFT JOIN jid AS sender_jid ON sender_jid._id = message.sender_jid_row_id
		WHERE jid.raw_string = ?
		ORDER BY message.timestamp ASC
	`

	rows, err := d.db.Query(query, jid)
	if err != nil {
		return nil, fmt.Errorf("failed to query modern messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	msgMap := make(map[string]*Message)

	for rows.Next() {
		var m Message
		var fromMeInt, linkId int
		var thumb, quotedThumb []byte

		err := rows.Scan(
			&m.ID,
			&m.ChatRowID,
			&fromMeInt,
			&m.Status,
			&m.Data,
			&m.Timestamp,
			&m.MediaURL,
			&m.MediaMimeType,
			&m.MediaWhatsappType,
			&m.MediaSize,
			&m.MediaName,
			&m.MediaCaption,
			&m.MediaDuration,
			&m.Latitude,
			&m.Longitude,
			&thumb,
			&quotedThumb,
			&m.QuotedMessageID,
			&linkId,
			&m.RemoteResource,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modern message: %w", err)
		}

		m.FromMe = fromMeInt == 1
		m.IsLink = linkId > 0
		m.Thumbnail = thumb
		m.QuotedThumbnail = quotedThumb

		if m.RemoteResource != "" {
			if name, ok := contacts[m.RemoteResource]; ok {
				m.RemoteResourceDisplayName = name
			} else if name, ok := contacts[m.RemoteResource+"@s.whatsapp.net"]; ok {
				m.RemoteResourceDisplayName = name
			}
		}

		messages = append(messages, m)
		msgMap[m.ID] = &messages[len(messages)-1]
	}

	// Link quotes
	for i := range messages {
		if messages[i].QuotedMessageID != "" {
			if q, ok := msgMap[messages[i].QuotedMessageID]; ok {
				messages[i].QuotedMessage = q
			}
		}
	}

	return messages, nil
}

func (d *Database) getMessagesLegacy(jid string, contacts map[string]string) ([]Message, error) {
	query := `
		SELECT
			messages.key_id,
			COALESCE(messages.key_remote_jid, ''),
			COALESCE(messages.key_from_me, 0),
			COALESCE(messages.status, 0),
			COALESCE(messages.data, ''),
			COALESCE(messages.timestamp, 0),
			COALESCE(messages.media_url, ''),
			COALESCE(messages.media_mime_type, ''),
			COALESCE(messages.media_wa_type, 0),
			COALESCE(messages.media_size, 0),
			COALESCE(messages.media_name, ''),
			COALESCE(messages.media_caption, ''),
			COALESCE(messages.media_duration, 0),
			COALESCE(messages.latitude, 0.0),
			COALESCE(messages.longitude, 0.0),
			messages.thumb_image,
			COALESCE(messages.remote_resource, ''),
			COALESCE(message_thumbnails.thumbnail, messages.raw_data),
			COALESCE(messages_quotes.key_id, ''),
			COALESCE(messages_links._id, 0)
		FROM messages
		LEFT JOIN message_thumbnails ON messages.key_id = message_thumbnails.key_id
		LEFT JOIN messages_quotes ON messages.quoted_row_id > 0 AND messages.quoted_row_id = messages_quotes._id
		LEFT JOIN messages_links ON messages._id = messages_links.message_row_id
		WHERE messages.key_remote_jid = ?
		ORDER BY messages.timestamp ASC
	`

	rows, err := d.db.Query(query, jid)
	if err != nil {
		return nil, fmt.Errorf("failed to query legacy messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	msgMap := make(map[string]*Message)

	for rows.Next() {
		var m Message
		var fromMeInt, linkId int
		var thumbImage, thumbData []byte

		err := rows.Scan(
			&m.ID,
			&m.ChatRowID,
			&fromMeInt,
			&m.Status,
			&m.Data,
			&m.Timestamp,
			&m.MediaURL,
			&m.MediaMimeType,
			&m.MediaWhatsappType,
			&m.MediaSize,
			&m.MediaName,
			&m.MediaCaption,
			&m.MediaDuration,
			&m.Latitude,
			&m.Longitude,
			&thumbImage,
			&m.RemoteResource,
			&thumbData,
			&m.QuotedMessageID,
			&linkId,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan legacy message: %w", err)
		}

		m.FromMe = fromMeInt == 1
		m.IsLink = linkId > 0

		if len(thumbData) > 0 {
			m.Thumbnail = thumbData
		} else {
			m.Thumbnail = thumbImage
		}

		if m.RemoteResource != "" {
			if name, ok := contacts[m.RemoteResource]; ok {
				m.RemoteResourceDisplayName = name
			} else if name, ok := contacts[m.RemoteResource+"@s.whatsapp.net"]; ok {
				m.RemoteResourceDisplayName = name
			}
		}

		messages = append(messages, m)
		msgMap[m.ID] = &messages[len(messages)-1]
	}

	// Link quotes
	for i := range messages {
		if messages[i].QuotedMessageID != "" {
			if q, ok := msgMap[messages[i].QuotedMessageID]; ok {
				messages[i].QuotedMessage = q
			}
		}
	}

	return messages, nil
}
