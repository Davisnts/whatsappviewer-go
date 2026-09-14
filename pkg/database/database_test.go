package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestDatabaseLegacySchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_legacy.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}

	schema := `
		CREATE TABLE messages (
			_id INTEGER PRIMARY KEY,
			key_remote_jid TEXT,
			key_from_me INTEGER,
			key_id TEXT,
			status INTEGER,
			data TEXT,
			timestamp INTEGER,
			media_url TEXT,
			media_mime_type TEXT,
			media_wa_type INTEGER,
			media_size INTEGER,
			media_name TEXT,
			media_caption TEXT,
			media_duration INTEGER,
			latitude REAL,
			longitude REAL,
			thumb_image BLOB,
			remote_resource TEXT,
			raw_data BLOB,
			quoted_row_id INTEGER
		);
		CREATE TABLE message_thumbnails (key_id TEXT, thumbnail BLOB);
		CREATE TABLE messages_quotes (_id INTEGER PRIMARY KEY, key_id TEXT);
		CREATE TABLE messages_links (message_row_id INTEGER, _id INTEGER);

		INSERT INTO messages (key_remote_jid, key_from_me, key_id, data, timestamp)
		VALUES ('12345@s.whatsapp.net', 0, 'MSG1', 'Hello world!', 1700000000000);

		INSERT INTO messages (key_remote_jid, key_from_me, key_id, data, timestamp)
		VALUES ('12345@s.whatsapp.net', 1, 'MSG2', 'Hi there!', 1700000010000);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	db.Close()

	wdb, err := OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("OpenDatabase failed: %v", err)
	}
	defer wdb.Close()

	if wdb.IsModern() {
		t.Errorf("expected legacy schema, got modern")
	}

	chats, err := wdb.GetChats(nil)
	if err != nil {
		t.Fatalf("GetChats failed: %v", err)
	}
	if len(chats) != 1 {
		t.Fatalf("expected 1 chat, got %d", len(chats))
	}
	if chats[0].JID != "12345@s.whatsapp.net" {
		t.Errorf("expected chat JID 12345@s.whatsapp.net, got %s", chats[0].JID)
	}
	if chats[0].TotalCount != 2 || chats[0].FromMeCount != 1 {
		t.Errorf("expected counts 1/2, got %d/%d", chats[0].FromMeCount, chats[0].TotalCount)
	}

	msgs, err := wdb.GetMessages("12345@s.whatsapp.net", nil)
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Data != "Hello world!" || msgs[1].Data != "Hi there!" {
		t.Errorf("unexpected message data")
	}
}
