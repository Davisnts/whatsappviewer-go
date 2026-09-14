package exporter

import (
	"bytes"
	"strings"
	"testing"

	"whatsapp-viewer/pkg/database"
)

func TestExporters(t *testing.T) {
	chat := database.Chat{
		JID:         "551199999999@s.whatsapp.net",
		DisplayName: "Test Contact",
		Subject:     "Test Subject",
	}

	msgs := []database.Message{
		{
			ID:        "M1",
			FromMe:    false,
			Data:      "Hello!",
			Timestamp: 1700000000000,
		},
		{
			ID:                "M2",
			FromMe:            true,
			Data:              "How are you?",
			Timestamp:         1700000060000,
			QuotedMessageID:   "M1",
			MediaWhatsappType: database.MediaText,
		},
	}
	msgs[1].QuotedMessage = &msgs[0]

	// Test TXT
	var txtBuf bytes.Buffer
	if err := WriteTxt(chat, msgs, &txtBuf); err != nil {
		t.Fatalf("WriteTxt failed: %v", err)
	}
	txtStr := txtBuf.String()
	if !strings.Contains(txtStr, "551199999999@s.whatsapp.net") || !strings.Contains(txtStr, "Hello!") || !strings.Contains(txtStr, "Quote from") {
		t.Errorf("TXT output missing expected elements: %s", txtStr)
	}

	// Test JSON
	var jsonBuf bytes.Buffer
	if err := WriteJSON(chat, msgs, &jsonBuf); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}
	jsonStr := jsonBuf.String()
	if !strings.Contains(jsonStr, "\"key\": \"551199999999@s.whatsapp.net\"") || !strings.Contains(jsonStr, "\"contactName\": \"Test Contact\"") {
		t.Errorf("JSON output missing expected elements: %s", jsonStr)
	}

	// Test HTML
	var htmlBuf bytes.Buffer
	if err := WriteHTML(chat, msgs, &htmlBuf); err != nil {
		t.Fatalf("WriteHTML failed: %v", err)
	}
	htmlStr := htmlBuf.String()
	if !strings.Contains(htmlStr, "Test Contact") || !strings.Contains(htmlStr, "Hello!") || !strings.Contains(htmlStr, "How are you?") {
		t.Errorf("HTML output missing expected elements: %s", htmlStr)
	}
}
