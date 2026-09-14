package exporter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"whatsapp-viewer/pkg/database"
)

// JSONChat represents the JSON export structure
type JSONChat struct {
	Key         string        `json:"key"`
	Subject     string        `json:"subject,omitempty"`
	ContactName string        `json:"contactName,omitempty"`
	Messages    []JSONMessage `json:"messages"`
}

// JSONMessage represents an individual message in JSON export
type JSONMessage struct {
	Timestamp                 string   `json:"timestamp"`
	ID                        string   `json:"id"`
	FromMe                    bool     `json:"fromMe"`
	RemoteResource            string   `json:"remoteResource,omitempty"`
	RemoteResourceDisplayName string   `json:"remoteResourceDisplayName,omitempty"`
	QuotedTimestamp           string   `json:"quotedTimestamp,omitempty"`
	QuotedMessageID           string   `json:"quotedMessageId,omitempty"`
	Type                      string   `json:"type"`
	Text                      string   `json:"text,omitempty"`
	Image                     string   `json:"image,omitempty"` // base64 encoded
	Filename                  string   `json:"filename,omitempty"`
	Caption                   string   `json:"caption,omitempty"`
	Duration                  int      `json:"duration,omitempty"`
	Latitude                  *float64 `json:"latitude,omitempty"`
	Longitude                 *float64 `json:"longitude,omitempty"`
	Link                      string   `json:"link,omitempty"`
}

// ExportJSON writes a chat to a JSON file
func ExportJSON(chat database.Chat, messages []database.Message, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create JSON export file: %w", err)
	}
	defer f.Close()

	return WriteJSON(chat, messages, f)
}

// WriteJSON formats and writes a chat to any io.Writer
func WriteJSON(chat database.Chat, messages []database.Message, w io.Writer) error {
	jsonChat := JSONChat{
		Key:         chat.JID,
		Subject:     chat.Subject,
		ContactName: chat.DisplayName,
		Messages:    make([]JSONMessage, 0, len(messages)),
	}

	for _, msg := range messages {
		jm := JSONMessage{
			Timestamp:                 msg.FormattedISO(),
			ID:                        msg.ID,
			FromMe:                    msg.FromMe,
			RemoteResource:            msg.RemoteResource,
			RemoteResourceDisplayName: msg.RemoteResourceDisplayName,
		}

		if msg.QuotedMessage != nil {
			jm.QuotedTimestamp = msg.QuotedMessage.FormattedISO()
			jm.QuotedMessageID = msg.QuotedMessage.ID
		}

		thumb := database.CleanThumbnail(msg.Thumbnail)
		if len(thumb) > 0 {
			jm.Image = base64.StdEncoding.EncodeToString(thumb)
		}

		if msg.MediaName != "" {
			jm.Filename = msg.MediaName
		}
		if msg.MediaCaption != "" {
			jm.Caption = msg.MediaCaption
		}

		switch msg.MediaWhatsappType {
		case database.MediaText:
			if msg.IsLink {
				jm.Type = "link"
				jm.Link = msg.Data
			} else {
				jm.Type = "text"
				jm.Text = msg.Data
			}
		case database.MediaImage:
			jm.Type = "image"
		case database.MediaAudio:
			jm.Type = "audio"
			jm.Duration = msg.MediaDuration
		case database.MediaVideo:
			jm.Type = "video"
		case database.MediaContact:
			jm.Type = "contact"
		case database.MediaLocation, database.MediaLiveLocation:
			jm.Type = "location"
			lat, lon := msg.Latitude, msg.Longitude
			jm.Latitude = &lat
			jm.Longitude = &lon
		case database.MediaGif:
			jm.Type = "gif"
		case database.MediaCall:
			jm.Type = "call"
		case database.MediaFile:
			jm.Type = "file"
		default:
			jm.Type = "unknown"
			jm.Text = msg.Data
		}

		jsonChat.Messages = append(jsonChat.Messages, jm)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonChat)
}
