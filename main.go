package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	uploadPath = os.Getenv("UPLOAD_PATH")
	publicURL  = os.Getenv("PUBLIC_URL")
	apiKey     = os.Getenv("API_KEY")
	port       = os.Getenv("PORT")
)

type uploadResponse struct {
	URL string `json:"url"`
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming %s request from %s", r.Method, r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		log.Printf("Rejected: method not POST")
		return
	}

	if r.Header.Get("X-API-Key") != apiKey {
		http.Error(w, "Authorised access only", http.StatusUnauthorized)
		log.Printf("Rejected: invalid API key from %s", r.RemoteAddr)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		log.Printf("Error reading file from %s: %v", r.RemoteAddr, err)
		return
	}
	defer file.Close()

	ext := filepath.Ext(handler.Filename)
	if ext == "" || len(ext) > 10 {
		http.Error(w, "File must have a valid extension", http.StatusUnsupportedMediaType)
		log.Printf("Invalid file extension '%s' from %s", ext, r.RemoteAddr)
		return
	}

	uniqueName := uuid.New().String() + ext
	dstPath := filepath.Join(uploadPath, uniqueName)

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		log.Printf("Error creating file %s: %v", dstPath, err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		log.Printf("Error writing file %s: %v", dstPath, err)
		return
	}

	log.Printf("Successfully uploaded file %s from %s (original filename: %s)", uniqueName, r.RemoteAddr, handler.Filename)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(uploadResponse{URL: publicURL + uniqueName}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Error encoding JSON response: %v", err)
		return
	}
}

func main() {
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	http.HandleFunc("/upload", uploadHandler)
	log.Printf("CDN started on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
