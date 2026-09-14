package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"whatsapp-viewer/pkg/database"
	"whatsapp-viewer/pkg/exporter"
)

//go:embed static/*
var staticFS embed.FS

// Server handles HTTP API requests and serves the web frontend
type Server struct {
	mu       sync.RWMutex
	db       *database.Database
	contacts map[string]string
	port     int
}

// NewServer creates a new web server instance
func NewServer(db *database.Database, contacts map[string]string, port int) *Server {
	if port <= 0 {
		port = 8080
	}
	return &Server{
		db:       db,
		contacts: contacts,
		port:     port,
	}
}

// SetDatabase updates the active database in the server
func (s *Server) SetDatabase(db *database.Database) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db = db
}

// Start launches the HTTP listener and optionally opens the browser
func (s *Server) Start(openBrowser bool) error {
	mux := http.NewServeMux()

	// Static files
	subFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		return fmt.Errorf("failed to load embedded static filesystem: %w", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFS))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Failed to load index.html", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// API endpoints
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/chats", s.handleChats)
	mux.HandleFunc("/api/messages", s.handleMessages)
	mux.HandleFunc("/api/export", s.handleExport)

	addr := fmt.Sprintf(":%d", s.port)
	url := fmt.Sprintf("http://localhost:%d", s.port)
	fmt.Printf("\n🚀 WhatsApp Viewer Web UI running at: %s\n", url)

	if openBrowser {
		go openURL(url)
	}

	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	loaded := s.db != nil
	isModern := false
	if loaded {
		isModern = s.db.IsModern()
	}

	resp := map[string]interface{}{
		"loaded":   loaded,
		"isModern": isModern,
		"contacts": len(s.contacts),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleChats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		http.Error(w, `{"error":"No database loaded"}`, http.StatusBadRequest)
		return
	}

	chats, err := s.db.GetChats(s.contacts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		http.Error(w, `{"error":"No database loaded"}`, http.StatusBadRequest)
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		http.Error(w, `{"error":"jid parameter required"}`, http.StatusBadRequest)
		return
	}

	msgs, err := s.db.GetMessages(jid, s.contacts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	type MessageResponse struct {
		database.Message
		TimeFormatted string `json:"timeFormatted"`
		DateFormatted string `json:"dateFormatted"`
		ThumbnailURL  string `json:"thumbnailUrl,omitempty"`
	}

	res := make([]MessageResponse, len(msgs))
	for i, m := range msgs {
		res[i] = MessageResponse{
			Message:       m,
			TimeFormatted: m.FormattedTime(),
			DateFormatted: m.FormattedDate(),
			ThumbnailURL:  m.ThumbnailBase64(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		http.Error(w, "No database loaded", http.StatusBadRequest)
		return
	}

	jid := r.URL.Query().Get("jid")
	format := strings.ToLower(r.URL.Query().Get("format"))
	if jid == "" {
		http.Error(w, "jid parameter required", http.StatusBadRequest)
		return
	}

	chats, err := s.db.GetChats(s.contacts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetChat database.Chat
	found := false
	for _, c := range chats {
		if c.JID == jid {
			targetChat = c
			found = true
			break
		}
	}
	if !found {
		targetChat = database.Chat{JID: jid}
	}

	msgs, err := s.db.GetMessages(jid, s.contacts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	safeName := strings.ReplaceAll(targetChat.Name(), " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")

	switch format {
	case "txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"chat_%s.txt\"", safeName))
		exporter.WriteTxt(targetChat, msgs, w)
	case "json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"chat_%s.json\"", safeName))
		exporter.WriteJSON(targetChat, msgs, w)
	case "html":
		fallthrough
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"chat_%s.html\"", safeName))
		exporter.WriteHTML(targetChat, msgs, w)
	}
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// Unused import guard
var _ = io.EOF
