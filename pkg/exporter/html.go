package exporter

import (
	"fmt"
	"html"
	"io"
	"os"
	"strings"

	"whatsapp-viewer/pkg/database"
)

// ExportHTML writes a chat to a standalone modern HTML file
func ExportHTML(chat database.Chat, messages []database.Message, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create HTML export file: %w", err)
	}
	defer f.Close()

	return WriteHTML(chat, messages, f)
}

// WriteHTML renders chat as a standalone HTML page
func WriteHTML(chat database.Chat, messages []database.Message, w io.Writer) error {
	title := html.EscapeString(chat.Name())

	header := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s - WhatsApp Viewer Export</title>
<style>
  :root {
    --bg-color: #0b141a;
    --chat-bg: #0b141a;
    --msg-incoming: #202c33;
    --msg-outgoing: #005c4b;
    --text-primary: #e9edef;
    --text-secondary: #8696a0;
    --header-bg: #202c33;
    --quote-bg: rgba(0, 0, 0, 0.2);
    --border-quote: #00a884;
    --accent: #00a884;
    --link-color: #53bdeb;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    background-color: var(--bg-color);
    color: var(--text-primary);
    display: flex;
    justify-content: center;
    padding: 20px;
    min-height: 100vh;
  }
  .container {
    width: 100%%;
    max-width: 850px;
    background: var(--chat-bg);
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 4px 20px rgba(0,0,0,0.4);
    display: flex;
    flex-direction: column;
  }
  .chat-header {
    background-color: var(--header-bg);
    padding: 16px 20px;
    display: flex;
    align-items: center;
    border-bottom: 1px solid rgba(255,255,255,0.06);
  }
  .avatar {
    width: 44px;
    height: 44px;
    border-radius: 50%%;
    background-color: #6b7c85;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    font-weight: 600;
    margin-right: 14px;
    flex-shrink: 0;
  }
  .chat-title {
    font-size: 17px;
    font-weight: 600;
    color: var(--text-primary);
  }
  .chat-subtitle {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 2px;
  }
  .messages-container {
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    background: #0b141a url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" opacity="0.03"><rect width="80" height="80" fill="none" stroke="%%23ffffff" stroke-width="1"/></svg>');
  }
  .message-row {
    display: flex;
    width: 100%%;
  }
  .message-row.outgoing { justify-content: flex-end; }
  .message-row.incoming { justify-content: flex-start; }

  .bubble {
    max-width: 75%%;
    padding: 8px 12px 6px;
    border-radius: 8px;
    position: relative;
    font-size: 14.2px;
    line-height: 19px;
    word-break: break-word;
    box-shadow: 0 1px 0.5px rgba(11,20,26,0.13);
  }
  .outgoing .bubble {
    background-color: var(--msg-outgoing);
    border-top-right-radius: 0;
  }
  .incoming .bubble {
    background-color: var(--msg-incoming);
    border-top-left-radius: 0;
  }
  .sender-name {
    font-size: 12.5px;
    font-weight: 600;
    color: #53bdeb;
    margin-bottom: 4px;
  }
  .quote {
    background-color: var(--quote-bg);
    border-left: 4px solid var(--border-quote);
    border-radius: 4px;
    padding: 5px 8px;
    margin-bottom: 6px;
    font-size: 12px;
  }
  .quote-author {
    font-weight: 600;
    color: var(--border-quote);
    margin-bottom: 2px;
  }
  .quote-text {
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .media-image {
    max-width: 100%%;
    border-radius: 6px;
    margin-bottom: 4px;
    display: block;
  }
  .media-caption {
    margin-top: 4px;
  }
  .media-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: rgba(255,255,255,0.08);
    padding: 6px 10px;
    border-radius: 6px;
    margin-bottom: 4px;
    font-size: 13px;
  }
  .meta {
    float: right;
    margin-left: 12px;
    margin-top: 4px;
    font-size: 11px;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .link { color: var(--link-color); text-decoration: none; }
  .link:hover { text-decoration: underline; }
  .day-divider {
    text-align: center;
    margin: 12px 0;
    position: relative;
  }
  .day-badge {
    background-color: #182229;
    color: var(--text-secondary);
    font-size: 11.5px;
    padding: 4px 12px;
    border-radius: 8px;
    display: inline-block;
    box-shadow: 0 1px 2px rgba(0,0,0,0.3);
  }
</style>
</head>
<body>
<div class="container">
  <div class="chat-header">
    <div class="avatar">%s</div>
    <div>
      <div class="chat-title">%s</div>
      <div class="chat-subtitle">%s &bull; %d messages</div>
    </div>
  </div>
  <div class="messages-container">
`,
		title,
		getInitials(chat.Name()),
		title,
		html.EscapeString(chat.JID),
		len(messages),
	)

	if _, err := fmt.Fprint(w, header); err != nil {
		return err
	}

	var lastDate string
	for _, msg := range messages {
		msgDate := msg.FormattedDate()
		if msgDate != "" && msgDate != lastDate {
			lastDate = msgDate
			fmt.Fprintf(w, "    <div class=\"day-divider\"><span class=\"day-badge\">%s</span></div>\n", msgDate)
		}

		rowClass := "incoming"
		if msg.FromMe {
			rowClass = "outgoing"
		}

		fmt.Fprintf(w, "    <div class=\"message-row %s\">\n", rowClass)
		fmt.Fprintf(w, "      <div class=\"bubble\">\n")

		// Sender in groups
		if !msg.FromMe && (msg.RemoteResourceDisplayName != "" || msg.RemoteResource != "") {
			sender := msg.RemoteResourceDisplayName
			if sender == "" {
				sender = msg.RemoteResource
			}
			fmt.Fprintf(w, "        <div class=\"sender-name\">%s</div>\n", html.EscapeString(sender))
		}

		// Quoted Message
		if msg.QuotedMessage != nil {
			fmt.Fprintf(w, "        <div class=\"quote\">\n")
			fmt.Fprintf(w, "          <div class=\"quote-author\">%s</div>\n", html.EscapeString(getMessageSenderName(*msg.QuotedMessage)))
			quoteText := msg.QuotedMessage.Data
			if quoteText == "" {
				quoteText = formatTxtContent(*msg.QuotedMessage)
			}
			fmt.Fprintf(w, "          <div class=\"quote-text\">%s</div>\n", html.EscapeString(quoteText))
			fmt.Fprintf(w, "        </div>\n")
		}

		// Media / Text Body
		renderHTMLMessageBody(w, msg)

		// Timestamp
		fmt.Fprintf(w, "        <span class=\"meta\">%s</span>\n", msg.FormattedTime())
		fmt.Fprintf(w, "      </div>\n")
		fmt.Fprintf(w, "    </div>\n")
	}

	footer := `  </div>
</div>
</body>
</html>`
	_, err := fmt.Fprintln(w, footer)
	return err
}

func renderHTMLMessageBody(w io.Writer, msg database.Message) {
	b64Img := msg.ThumbnailBase64()

	switch msg.MediaWhatsappType {
	case database.MediaImage, database.MediaGif:
		if b64Img != "" {
			fmt.Fprintf(w, "        <img class=\"media-image\" src=\"%s\" alt=\"thumbnail\">\n", b64Img)
		} else {
			fmt.Fprintf(w, "        <div class=\"media-badge\">📷 Photo</div>\n")
		}
		if msg.MediaCaption != "" {
			fmt.Fprintf(w, "        <div class=\"media-caption\">%s</div>\n", formatHtmlText(msg.MediaCaption))
		}

	case database.MediaAudio:
		fmt.Fprintf(w, "        <div class=\"media-badge\">🎵 Audio (%ds)</div>\n", msg.MediaDuration)

	case database.MediaVideo:
		if b64Img != "" {
			fmt.Fprintf(w, "        <img class=\"media-image\" src=\"%s\" alt=\"video thumbnail\">\n", b64Img)
		}
		fmt.Fprintf(w, "        <div class=\"media-badge\">🎬 Video %s</div>\n", html.EscapeString(msg.MediaName))
		if msg.MediaCaption != "" {
			fmt.Fprintf(w, "        <div class=\"media-caption\">%s</div>\n", formatHtmlText(msg.MediaCaption))
		}

	case database.MediaContact:
		fmt.Fprintf(w, "        <div class=\"media-badge\">👤 Contact Card</div>\n")
		if msg.Data != "" {
			fmt.Fprintf(w, "        <div>%s</div>\n", formatHtmlText(msg.Data))
		}

	case database.MediaLocation, database.MediaLiveLocation:
		mapUrl := fmt.Sprintf("https://www.google.com/maps?q=%f,%f", msg.Latitude, msg.Longitude)
		fmt.Fprintf(w, "        <div class=\"media-badge\">📍 <a class=\"link\" href=\"%s\" target=\"_blank\">Location (%.5f, %.5f)</a></div>\n",
			mapUrl, msg.Latitude, msg.Longitude)

	case database.MediaFile:
		name := msg.MediaName
		if name == "" {
			name = "Document"
		}
		fmt.Fprintf(w, "        <div class=\"media-badge\">📄 %s</div>\n", html.EscapeString(name))

	default:
		if msg.IsLink {
			if b64Img != "" {
				fmt.Fprintf(w, "        <img class=\"media-image\" src=\"%s\" alt=\"link preview\">\n", b64Img)
			}
			if msg.MediaCaption != "" {
				fmt.Fprintf(w, "        <div><strong>%s</strong></div>\n", html.EscapeString(msg.MediaCaption))
			}
			fmt.Fprintf(w, "        <div><a class=\"link\" href=\"%s\" target=\"_blank\">%s</a></div>\n",
				html.EscapeString(msg.Data), html.EscapeString(msg.Data))
		} else {
			fmt.Fprintf(w, "        <div>%s</div>\n", formatHtmlText(msg.Data))
		}
	}
}

func formatHtmlText(s string) string {
	escaped := html.EscapeString(s)
	return strings.ReplaceAll(escaped, "\n", "<br>")
}

func getInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "W"
	}
	if len(parts) == 1 {
		runes := []rune(parts[0])
		if len(runes) > 0 {
			return strings.ToUpper(string(runes[:1]))
		}
		return "W"
	}
	r1 := []rune(parts[0])
	r2 := []rune(parts[len(parts)-1])
	return strings.ToUpper(string(r1[:1]) + string(r2[:1]))
}

// getMessageSenderName returns a friendly sender string
func getMessageSenderName(m database.Message) string {
	if m.FromMe {
		return "You"
	}
	if m.RemoteResourceDisplayName != "" {
		return m.RemoteResourceDisplayName
	}
	if m.RemoteResource != "" {
		return m.RemoteResource
	}
	return "Sender"
}
