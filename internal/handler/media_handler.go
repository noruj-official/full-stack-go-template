package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/service"
)

type MediaHandler struct {
	*Handler
	mediaService *service.MediaService
}

func NewMediaHandler(base *Handler, mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{
		Handler:      base,
		mediaService: mediaService,
	}
}

// Upload handles generic media uploads (used by editor, gallery, etc.)
// Returns JSON: { id, url, filename }
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		h.Error(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 10MB max upload
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid form data")
		return
	}

	file, header, err := r.FormFile("file") // Tiptap usually sends "file" or "image"
	if err != nil {
		// Try "image" field as fallback
		file, header, err = r.FormFile("image")
		if err != nil {
			h.Error(w, r, http.StatusBadRequest, "No file provided")
			return
		}
	}
	defer file.Close()

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Detect Content-Type (sniffer or header)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	// Basic validation (allow images only for now?)
	if !strings.HasPrefix(contentType, "image/") {
		h.Error(w, r, http.StatusBadRequest, "Only image files are allowed")
		return
	}

	// Use original filename or generate one
	filename := header.Filename
	if filename == "" {
		ext := "jpg"
		if strings.Contains(contentType, "png") {
			ext = "png"
		}
		if strings.Contains(contentType, "gif") {
			ext = "gif"
		}
		if strings.Contains(contentType, "webp") {
			ext = "webp"
		}
		filename = fmt.Sprintf("upload.%s", ext)
	}

	input := domain.CreateMediaInput{
		UserID:          &user.ID,
		Filename:        filename,
		Data:            data,
		ContentType:     contentType,
		SizeBytes:       len(data),
		AltText:         filename, // Default alt text
		StorageProvider: domain.StorageProviderDatabase,
	}

	media, err := h.mediaService.Upload(r.Context(), input)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to upload media")
		return
	}

	// Construct URL: /media/{uuid}.{ext}
	ext := filepath.Ext(media.Filename)
	if ext == "" {
		ext = ".jpg" // Default fallback
	}
	// Ensure extension in URL matches file type if possible, or just use what we have.
	// Users asked to support upload .png -> store .png -> serve .png.
	// We stored the filename. We can use the media ID + detected extension.
	// Or just append the extension from the original filename.

	publicURL := fmt.Sprintf("/media/%s%s", media.ID, ext)

	// Return JSON for editor
	response := map[string]string{
		"id":       media.ID.String(),
		"url":      publicURL,
		"filename": media.Filename,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Serve handles GET /media/{filename}
// Filename is expect to be UUID.ext or just UUID
func (h *MediaHandler) Serve(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Extract UUID (remove extension)
	idStr := filename
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		idStr = filename[:idx]
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	media, err := h.mediaService.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// If in future we have S3, we might redirect here
	if media.StorageProvider == domain.StorageProviderS3 && media.PublicURL != "" {
		http.Redirect(w, r, media.PublicURL, http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", media.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
	w.Write(media.Data)
}
