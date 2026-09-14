package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// LoadContacts loads WhatsApp contacts mapping JID to DisplayName from wa.db
func LoadContacts(waDbPath string) (map[string]string, error) {
	contacts := make(map[string]string)
	if waDbPath == "" {
		return contacts, nil
	}

	if _, err := os.Stat(waDbPath); err != nil {
		return nil, fmt.Errorf("wa.db file not found: %w", err)
	}

	db, err := sql.Open("sqlite", waDbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open wa.db: %w", err)
	}
	defer db.Close()

	// Query wa_contacts table
	rows, err := db.Query("SELECT jid, display_name, wa_name, number FROM wa_contacts")
	if err != nil {
		// Try simpler query if columns differ
		rows, err = db.Query("SELECT jid, display_name FROM wa_contacts")
		if err != nil {
			return contacts, nil // ignore if table not present
		}
		defer rows.Close()

		for rows.Next() {
			var jid, displayName sql.NullString
			if err := rows.Scan(&jid, &displayName); err == nil && jid.Valid && jid.String != "" {
				if displayName.Valid && displayName.String != "" {
					contacts[jid.String] = displayName.String
				}
			}
		}
		return contacts, nil
	}
	defer rows.Close()

	for rows.Next() {
		var jid, displayName, waName, number sql.NullString
		if err := rows.Scan(&jid, &displayName, &waName, &number); err == nil && jid.Valid && jid.String != "" {
			name := ""
			if displayName.Valid && displayName.String != "" {
				name = displayName.String
			} else if waName.Valid && waName.String != "" {
				name = waName.String
			} else if number.Valid && number.String != "" {
				name = number.String
			}
			if name != "" {
				contacts[jid.String] = name
			}
		}
	}

	return contacts, nil
}
