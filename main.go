package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"whatsapp-viewer/pkg/crypto"
	"whatsapp-viewer/pkg/database"
	"whatsapp-viewer/pkg/exporter"
	"whatsapp-viewer/pkg/web"
)

const version = "2.0.0-go"

func displayUsage() {
	fmt.Printf(`WhatsApp Viewer (Go Version %s)

Supported commands:

Decryption:
  -decrypt5  <input database> <account name> <output filename>
             example: -decrypt5 msgstore.db.crypt5 your@email.com msgstore.decrypted.db

  -decrypt7  <input database> <key filename> <output filename>
             example: -decrypt7 msgstore.db.crypt7 key msgstore.decrypted.db

  -decrypt8  <input database> <key filename> <output filename>
             example: -decrypt8 msgstore.db.crypt8 key msgstore.decrypted.db

  -decrypt12 <input database> <key filename> <output filename>
             example: -decrypt12 msgstore.db.crypt12 key msgstore.decrypted.db

  -decrypt14 <input database> <key filename> <output filename>
             example: -decrypt14 msgstore.db.crypt14 key msgstore.decrypted.db

Export:
  export -db <msgstore.db> [-wa <wa.db>] [-jid <chat_jid>] -format <html|json|txt> -out <file_or_dir>
             example: export -db msgstore.db -format html -out exported_chats

Web UI:
  serve -db <msgstore.db> [-wa <wa.db>] [-port 8080]
             example: serve -db msgstore.db
             (opens modern WhatsApp Web interface in your default browser)
`, version)
}

func main() {
	if len(os.Args) < 2 {
		// If test-database.db or msgstore.db exists, start server automatically
		defaultDb := ""
		candidates := []string{"msgstore.db", "data/test-database.db", "../data/test-database.db"}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				defaultDb = c
				break
			}
		}

		if defaultDb != "" {
			fmt.Printf("Default database found: %s. Launching Web UI...\n", defaultDb)
			startWebUI(defaultDb, "", 8080, true)
			return
		}

		displayUsage()
		return
	}

	arg1 := strings.ToLower(os.Args[1])

	switch arg1 {
	case "-decrypt5":
		if len(os.Args) < 5 {
			fmt.Println("Error: Invalid arguments for -decrypt5")
			displayUsage()
			os.Exit(1)
		}
		err := crypto.DecryptCrypt5(os.Args[2], os.Args[4], os.Args[3])
		handleResult(err, "Decrypted crypt5 database successfully saved to: "+os.Args[4])

	case "-decrypt7":
		if len(os.Args) < 5 {
			fmt.Println("Error: Invalid arguments for -decrypt7")
			displayUsage()
			os.Exit(1)
		}
		err := crypto.DecryptCrypt7WithKeyFile(os.Args[2], os.Args[4], os.Args[3])
		handleResult(err, "Decrypted crypt7 database successfully saved to: "+os.Args[4])

	case "-decrypt8":
		if len(os.Args) < 5 {
			fmt.Println("Error: Invalid arguments for -decrypt8")
			displayUsage()
			os.Exit(1)
		}
		err := crypto.DecryptCrypt8WithKeyFile(os.Args[2], os.Args[4], os.Args[3])
		handleResult(err, "Decrypted crypt8 database successfully saved to: "+os.Args[4])

	case "-decrypt12":
		if len(os.Args) < 5 {
			fmt.Println("Error: Invalid arguments for -decrypt12")
			displayUsage()
			os.Exit(1)
		}
		err := crypto.DecryptCrypt12WithKeyFile(os.Args[2], os.Args[4], os.Args[3])
		handleResult(err, "Decrypted crypt12 database successfully saved to: "+os.Args[4])

	case "-decrypt14":
		if len(os.Args) < 5 {
			fmt.Println("Error: Invalid arguments for -decrypt14")
			displayUsage()
			os.Exit(1)
		}
		err := crypto.DecryptCrypt14WithKeyFile(os.Args[2], os.Args[4], os.Args[3])
		handleResult(err, "Decrypted crypt14 database successfully saved to: "+os.Args[4])

	case "export":
		runExport(os.Args[2:])

	case "serve":
		runServe(os.Args[2:])

	case "-h", "--help", "help":
		displayUsage()

	default:
		// Check if user passed a .db file directly as first argument
		if strings.HasSuffix(strings.ToLower(os.Args[1]), ".db") {
			startWebUI(os.Args[1], "", 8080, true)
			return
		}
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		displayUsage()
		os.Exit(1)
	}
}

func handleResult(err error, successMsg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ %s\n", successMsg)
}

func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	dbPath := fs.String("db", "", "Path to msgstore.db database")
	waPath := fs.String("wa", "", "Optional path to wa.db (contacts)")
	jid := fs.String("jid", "", "Optional specific chat JID to export")
	format := fs.String("format", "html", "Export format: html, json, txt")
	outPath := fs.String("out", "exports", "Output file or folder")
	fs.Parse(args)

	if *dbPath == "" {
		fmt.Println("Error: -db parameter is required for export")
		os.Exit(1)
	}

	wdb, err := database.OpenDatabase(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer wdb.Close()

	contacts, err := database.LoadContacts(*waPath)
	if err != nil {
		fmt.Printf("Warning: failed to load contacts from wa.db: %v\n", err)
	}

	chats, err := wdb.GetChats(contacts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying chats: %v\n", err)
		os.Exit(1)
	}

	if *jid != "" {
		// Single chat export
		var targetChat *database.Chat
		for _, c := range chats {
			if c.JID == *jid {
				targetChat = &c
				break
			}
		}
		if targetChat == nil {
			targetChat = &database.Chat{JID: *jid}
		}
		exportSingleChat(wdb, *targetChat, contacts, *format, *outPath)
		return
	}

	// Bulk export all chats
	os.MkdirAll(*outPath, 0755)
	fmt.Printf("Exporting %d chats to '%s' (format: %s)...\n", len(chats), *outPath, *format)

	for i, chat := range chats {
		safeName := sanitizeFilename(chat.Name())
		fileName := fmt.Sprintf("%s_%s.%s", safeName, chat.JID, *format)
		targetFile := filepath.Join(*outPath, fileName)

		msgs, err := wdb.GetMessages(chat.JID, contacts)
		if err != nil {
			fmt.Printf("[%d/%d] ⚠️ Failed chat %s: %v\n", i+1, len(chats), chat.JID, err)
			continue
		}

		switch strings.ToLower(*format) {
		case "txt":
			err = exporter.ExportTxt(chat, msgs, targetFile)
		case "json":
			err = exporter.ExportJSON(chat, msgs, targetFile)
		default:
			err = exporter.ExportHTML(chat, msgs, targetFile)
		}

		if err != nil {
			fmt.Printf("[%d/%d] ⚠️ Failed writing %s: %v\n", i+1, len(chats), fileName, err)
		} else {
			fmt.Printf("[%d/%d] ✅ Exported %s (%d messages)\n", i+1, len(chats), fileName, len(msgs))
		}
	}
	fmt.Println("Export completed successfully!")
}

func exportSingleChat(wdb *database.Database, chat database.Chat, contacts map[string]string, format, outPath string) {
	msgs, err := wdb.GetMessages(chat.JID, contacts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading messages: %v\n", err)
		os.Exit(1)
	}

	outFile := outPath
	if fi, err := os.Stat(outPath); err == nil && fi.IsDir() {
		outFile = filepath.Join(outPath, fmt.Sprintf("chat_%s.%s", sanitizeFilename(chat.Name()), format))
	}

	switch strings.ToLower(format) {
	case "txt":
		err = exporter.ExportTxt(chat, msgs, outFile)
	case "json":
		err = exporter.ExportJSON(chat, msgs, outFile)
	default:
		err = exporter.ExportHTML(chat, msgs, outFile)
	}

	handleResult(err, fmt.Sprintf("Chat exported successfully to: %s", outFile))
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	dbPath := fs.String("db", "", "Path to msgstore.db database")
	waPath := fs.String("wa", "", "Optional path to wa.db (contacts)")
	port := fs.Int("port", 8080, "Port for web UI")
	noBrowser := fs.Bool("no-browser", false, "Do not automatically open browser")
	fs.Parse(args)

	startWebUI(*dbPath, *waPath, *port, !*noBrowser)
}

func startWebUI(dbPath, waPath string, port int, openBrowser bool) {
	var wdb *database.Database
	var contacts map[string]string

	if dbPath != "" {
		var err error
		wdb, err = database.OpenDatabase(dbPath)
		if err != nil {
			fmt.Printf("⚠️ Warning: could not open database %s: %v\n", dbPath, err)
		} else {
			fmt.Printf("Loaded database: %s (Modern: %v)\n", dbPath, wdb.IsModern())
		}
	}

	if waPath != "" {
		var err error
		contacts, err = database.LoadContacts(waPath)
		if err != nil {
			fmt.Printf("⚠️ Warning: could not load contacts: %v\n", err)
		} else {
			fmt.Printf("Loaded %d contacts from wa.db\n", len(contacts))
		}
	}

	srv := web.NewServer(wdb, contacts, port)
	if err := srv.Start(openBrowser); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		os.Exit(1)
	}
}

func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, ch := range invalid {
		result = strings.ReplaceAll(result, ch, "_")
	}
	result = strings.TrimSpace(result)
	if result == "" {
		result = "chat"
	}
	return result
}
