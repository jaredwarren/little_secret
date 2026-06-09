package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jaredwarren/LittleSecrets/internal/server"
)

func main() {
	// Setup port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Paths
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	packsDir := filepath.Join(cwd, "data", "packs")
	staticDir := filepath.Join(cwd, "static")

	// Ensure static directory exists (or alert)
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		log.Fatalf("failed to create static directory: %v", err)
	}

	// Initialize pack system with default Classic Pack
	_, err = server.LoadPacks(packsDir)
	if err != nil {
		log.Printf("Warning: failed to load/seed packs: %v", err)
	}

	hub := server.NewHub(packsDir)

	// HTTP Routing
	mux := http.NewServeMux()

	// Serve static files, falling back to index.html for SPA/custom routes (like /home)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(staticDir, filepath.Clean(r.URL.Path))

		// Check if file exists and is not a directory
		info, err := os.Stat(filePath)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, filePath)
			return
		}

		// Fallback to index.html
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	// WebSocket handler
	mux.HandleFunc("/ws", hub.HandleConnection)

	// API endpoints for custom card packs
	mux.HandleFunc("/api/packs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			packs, err := server.LoadPacks(packsDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Return list of pack names
			var packNames []string
			for name := range packs {
				packNames = append(packNames, name)
			}
			if err := json.NewEncoder(w).Encode(packNames); err != nil {
				log.Printf("Error encoding pack names: %v", err)
			}

		case http.MethodPost:
			var pack server.MissionPack
			if err := json.NewDecoder(r.Body).Decode(&pack); err != nil {
				http.Error(w, "Invalid JSON structure", http.StatusBadRequest)
				return
			}

			if pack.Name == "" {
				http.Error(w, "Pack name is required", http.StatusBadRequest)
				return
			}

			if len(pack.Words) != 21 {
				http.Error(w, "Pack must contain exactly 21 word pairs", http.StatusBadRequest)
				return
			}

			// Save pack
			if err := server.SavePack(packsDir, pack); err != nil {
				http.Error(w, "Failed to save pack: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(map[string]string{"status": "created", "name": pack.Name}); err != nil {
				log.Printf("Error encoding created response: %v", err)
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Little Secret server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
