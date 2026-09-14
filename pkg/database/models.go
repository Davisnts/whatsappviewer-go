package database

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"
)

// Media types corresponding to WhatsApp internal types
const (
	MediaText         = 0
	MediaImage        = 1
	MediaAudio        = 2
	MediaVideo        = 3
	MediaContact      = 4
	MediaLocation     = 5
	MediaCall         = 8
	MediaFile         = 9
	MediaGif          = 13
	MediaLiveLocation = 16
)

// Chat represents a conversation thread in WhatsApp
type Chat struct {
	JID                  string `json:"jid"`
	DisplayName          string `json:"displayName"`
	Subject              string `json:"subject"`
	CreatedTimestamp     int64  `json:"createdTimestamp"`
	LastMessageTimestamp int64  `json:"lastMessageTimestamp"`
	MessagesCount        int    `json:"messagesCount"`
	FromMeCount          int    `json:"fromMeCount"`
	TotalCount           int    `json:"totalCount"`
}

// Name returns the best available name for this chat
func (c Chat) Name() string {
	if c.DisplayName != "" {
		return c.DisplayName
	}
	if c.Subject != "" {
		return c.Subject
	}
	return c.JID
}

// Message represents an individual message
type Message struct {
	ID                         string   `json:"id"`
	ChatRowID                  string   `json:"chatRowId"`
	FromMe                     bool     `json:"fromMe"`
	Status                     int      `json:"status"`
	Data                       string   `json:"data"`
	Timestamp                  int64    `json:"timestamp"` // milliseconds since unix epoch
	MediaURL                   string   `json:"mediaUrl,omitempty"`
	MediaMimeType              string   `json:"mediaMimeType,omitempty"`
	MediaWhatsappType          int      `json:"mediaWhatsappType"`
	MediaSize                  int64    `json:"mediaSize,omitempty"`
	MediaName                  string   `json:"mediaName,omitempty"`
	MediaCaption               string   `json:"mediaCaption,omitempty"`
	MediaDuration              int      `json:"mediaDuration,omitempty"`
	Latitude                   float64  `json:"latitude,omitempty"`
	Longitude                  float64  `json:"longitude,omitempty"`
	Thumbnail                  []byte   `json:"-"`
	QuotedThumbnail            []byte   `json:"-"`
	RemoteResource             string   `json:"remoteResource,omitempty"`
	RemoteResourceDisplayName string   `json:"remoteResourceDisplayName,omitempty"`
	QuotedMessageID            string   `json:"quotedMessageId,omitempty"`
	QuotedMessage              *Message `json:"quotedMessage,omitempty"`
	IsLink                     bool     `json:"isLink"`
}

// Time converts timestamp in milliseconds to time.Time
func (m Message) Time() time.Time {
	return time.UnixMilli(m.Timestamp)
}

// FormattedTimestamp returns standard date time string
func (m Message) FormattedTimestamp() string {
	if m.Timestamp == 0 {
		return ""
	}
	return m.Time().Format("02/01/2006 15:04:05")
}

// FormattedISO returns RFC3339 timestamp
func (m Message) FormattedISO() string {
	if m.Timestamp == 0 {
		return ""
	}
	return m.Time().UTC().Format(time.RFC3339)
}

// FormattedTime returns only HH:MM
func (m Message) FormattedTime() string {
	if m.Timestamp == 0 {
		return ""
	}
	return m.Time().Format("15:04")
}

// FormattedDate returns only DD/MM/YYYY
func (m Message) FormattedDate() string {
	if m.Timestamp == 0 {
		return ""
	}
	return m.Time().Format("02/01/2006")
}

// CleanThumbnail extracts valid JPEG or PNG image bytes from raw thumbnail blobs.
// WhatsApp often prepends a 27-byte Java serialization header (X'aced0005...') before the JPEG.
func CleanThumbnail(raw []byte) []byte {
	if len(raw) == 0 {
		return nil
	}

	// Look for JPEG magic header 0xFF 0xD8
	idx := bytes.Index(raw, []byte{0xFF, 0xD8})
	if idx >= 0 {
		return raw[idx:]
	}

	// Look for PNG magic header 0x89 PNG
	idx = bytes.Index(raw, []byte{0x89, 'P', 'N', 'G'})
	if idx >= 0 {
		return raw[idx:]
	}

	// Look for WebP magic header RIFF....WEBP
	if len(raw) > 12 && string(raw[:4]) == "RIFF" && string(raw[8:12]) == "WEBP" {
		return raw
	}

	return raw
}

// ThumbnailBase64 returns thumbnail encoded as data URL
func (m Message) ThumbnailBase64() string {
	cleaned := CleanThumbnail(m.Thumbnail)
	if len(cleaned) == 0 {
		return ""
	}
	mime := "image/jpeg"
	if bytes.HasPrefix(cleaned, []byte{0x89, 'P', 'N', 'G'}) {
		mime = "image/png"
	} else if len(cleaned) > 12 && string(cleaned[:4]) == "RIFF" && string(cleaned[8:12]) == "WEBP" {
		mime = "image/webp"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(cleaned))
}

// QuotedThumbnailBase64 returns quoted thumbnail encoded as data URL
func (m Message) QuotedThumbnailBase64() string {
	cleaned := CleanThumbnail(m.QuotedThumbnail)
	if len(cleaned) == 0 {
		return ""
	}
	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(cleaned))
}

// HasThumbnail checks if message has any thumbnail
func (m Message) HasThumbnail() bool {
	return len(CleanThumbnail(m.Thumbnail)) > 0
}
