package api

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"time-guard-bot/internal/helpers"
	"time-guard-bot/internal/storage"
)

// Custom type for context keys
type ContextKey string

// Context key constants
const (
	ChatIDKey ContextKey = "chatID"
)

// Extracts the chat ID from the context // FIXME?
func GetChatIDFromContext(ctx context.Context) (int64, bool) {
	chatID, ok := ctx.Value(ChatIDKey).(int64)
	return chatID, ok
}

// Represents the API server
type Server struct {
	storage storage.Storage
	addr    string
	server  *http.Server
}

// Represents API server configuration
type Config struct {
	Addr string
}

// Creates a new API server
func NewServer(config *Config, s storage.Storage) *Server {
	return &Server{
		storage: s,
		addr:    config.Addr,
	}
}

// Starts the API server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/task/status", s.authMiddleware(s.handleTaskStatus))
	mux.HandleFunc("/api/task/list", s.authMiddleware(s.handleTaskList))

	// Register Swagger routes
	RegisterSwaggerRoutes(mux)

	// Create HTTP server
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		log.Printf("API server starting on %s", s.addr)
		log.Printf("Swagger UI available at http://%s/swagger/index.html", s.addr)

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	return nil
}

// Gracefully stops the API server
func (s *Server) Stop() error {
	log.Println("Stopping API server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// Checks API key validity
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			sendJSONError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			sendJSONError(w, "Invalid Authorization format. Expected: Bearer API_KEY", http.StatusUnauthorized)
			return
		}

		apiKey := parts[1]

		chatID, err := helpers.ExtractChatID(apiKey)
		if err != nil {
			log.Printf("Error verifying API key: %v", err)
			sendJSONError(w, "Internal server error", http.StatusInternalServerError)

			return
		}

		// Check if the chat exists in the database
		exists, err := s.storage.ChatExists(r.Context(), chatID)
		if err != nil {
			log.Printf("Error checking if chat exists: %v", err)
			sendJSONError(w, "Internal server error", http.StatusInternalServerError)

			return
		}

		if !exists {
			sendJSONError(w, "Chat not found or has no tasks", http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), ChatIDKey, chatID)

		// Call the next handler
		next(w, r.WithContext(ctx))
	}
}
