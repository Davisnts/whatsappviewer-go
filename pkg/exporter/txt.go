package exporter

import (
	"fmt"
	"io"
	"os"

	"whatsapp-viewer/pkg/database"
)

// ExportTxt writes a chat to a plain text file matching WhatsApp Viewer's format
func ExportTxt(chat database.Chat, messages []database.Message, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create text export file: %w", err)
	}
	defer f.Close()

	return WriteTxt(chat, messages, f)
}

// WriteTxt formats and writes a chat to any io.Writer
func WriteTxt(chat database.Chat, messages []database.Message, w io.Writer) error {
	header := chat.JID
	if chat.Subject != "" {
		header += "; " + chat.Subject
	}
	if _, err := fmt.Fprintf(w, "%s\n\n", header); err != nil {
		return err
	}

	for _, msg := range messages {
		sender := "he: "
		if msg.FromMe {
			sender = "ME: "
		}

		quoteStr := ""
		if msg.QuotedMessage != nil {
			quoteStr = fmt.Sprintf("[ Quote from %s: %s ]; ", msg.QuotedMessage.FormattedTimestamp(), msg.QuotedMessage.ID)
		}

		content := formatTxtContent(msg)

		line := fmt.Sprintf("%s; %s%s%s\n", msg.FormattedTimestamp(), sender, quoteStr, content)
		if _, err := fmt.Fprint(w, line); err != nil {
			return err
		}
	}

	return nil
}

func formatTxtContent(msg database.Message) string {
	switch msg.MediaWhatsappType {
	case database.MediaText:
		if msg.IsLink {
			if msg.MediaCaption != "" {
				return fmt.Sprintf("[ Link: %s; %s ]", msg.MediaCaption, msg.Data)
			}
			return fmt.Sprintf("[ Link: %s ]", msg.Data)
		}
		return msg.Data

	case database.MediaImage:
		name := msg.MediaName
		if name == "" {
			name = "image"
		}
		if msg.MediaCaption != "" {
			return fmt.Sprintf("[ Image: %s; %s ]", name, msg.MediaCaption)
		}
		return fmt.Sprintf("[ Image: %s ]", name)

	case database.MediaAudio:
		name := msg.MediaName
		if name == "" {
			name = "audio"
		}
		return fmt.Sprintf("[ Audio: %ds; %s ]", msg.MediaDuration, name)

	case database.MediaVideo:
		name := msg.MediaName
		if name == "" {
			name = "video"
		}
		if msg.MediaCaption != "" {
			return fmt.Sprintf("[ Video: %s; %s ]", name, msg.MediaCaption)
		}
		return fmt.Sprintf("[ Video: %s ]", name)

	case database.MediaContact:
		return "[ Contact ]"

	case database.MediaLocation, database.MediaLiveLocation:
		return fmt.Sprintf("[ Location: %f,%f ]", msg.Latitude, msg.Longitude)

	case database.MediaGif:
		name := msg.MediaName
		if name == "" {
			name = "gif"
		}
		if msg.MediaCaption != "" {
			return fmt.Sprintf("[ GIF: %s; %s ]", name, msg.MediaCaption)
		}
		return fmt.Sprintf("[ GIF: %s ]", name)

	case database.MediaCall:
		return "[ Call ]"

	case database.MediaFile:
		name := msg.MediaName
		if name == "" {
			name = "file"
		}
		return fmt.Sprintf("[ File: %s ]", name)

	default:
		if msg.Data != "" {
			return msg.Data
		}
		return "[ Media ]"
	}
}
