package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
	"github.com/noruj-official/full-stack-go-template/internal/middleware"
	"github.com/noruj-official/full-stack-go-template/internal/service"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages/admin"
	"github.com/noruj-official/full-stack-go-template/web/templ/pages/blog"
)

type BlogHandler struct {
	*Handler
	blogService  *service.BlogService
	mediaService *service.MediaService
}

func NewBlogHandler(base *Handler, blogService *service.BlogService, mediaService *service.MediaService) *BlogHandler {
	return &BlogHandler{
		Handler:      base,
		blogService:  blogService,
		mediaService: mediaService,
	}
}

func (h *BlogHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)
	h.RenderTempl(w, r, pages.NotFound("Page Not Found", "The page you requested was not found.", user, theme, themeEnabled, oauthEnabled))
}

// Public Routes

func (h *BlogHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	isPublished := true
	filter := domain.BlogFilter{
		IsPublished: &isPublished,
		Limit:       limit,
		Offset:      offset,
	}

	blogs, total, err := h.blogService.List(r.Context(), filter)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load blogs")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	h.RenderTempl(w, r, blog.List("Blog", blogs, total, page, limit, user, theme, themeEnabled, oauthEnabled))
}

func (h *BlogHandler) View(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug") // Using Go 1.22+ routing matched path value
	if slug == "" {
		slug = chi.URLParam(r, "slug") // Fallback if using chi in some places or different mux
	}

	b, err := h.blogService.GetBySlug(r.Context(), slug)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	if !b.IsPublished {
		// Only admin/author can see unpublished? For now just 404
		h.NotFound(w, r)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	h.RenderTempl(w, r, blog.View(b.Title, b, user, theme, themeEnabled, oauthEnabled))
}

// GetCoverImage serves the cover image for a blog post (public)
func (h *BlogHandler) GetCoverImage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		slug = chi.URLParam(r, "slug")
	}

	b, err := h.blogService.GetBySlug(r.Context(), slug)
	if err != nil || b.CoverMediaID == nil {
		http.NotFound(w, r)
		return
	}

	media, err := h.mediaService.GetByID(r.Context(), *b.CoverMediaID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", media.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	w.Write(media.Data)
}

// Admin Routes

func (h *BlogHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	filter := domain.BlogFilter{
		Limit:  limit,
		Offset: offset,
	}

	blogs, total, err := h.blogService.List(r.Context(), filter)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load blogs")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	h.RenderTempl(w, r, admin.BlogList("Manage Blogs", blogs, total, page, limit, user, theme, themeEnabled, oauthEnabled))
}

func (h *BlogHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	h.RenderTempl(w, r, admin.ReactBlogEditor("Create Blog", nil, user, theme, themeEnabled, oauthEnabled))
}

func (h *BlogHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid form data")
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	excerpt := r.FormValue("excerpt")
	isPublished := r.FormValue("is_published") == "on"

	// SEO metadata
	metaTitle := r.FormValue("meta_title")
	metaDescription := r.FormValue("meta_description")
	metaKeywords := r.FormValue("meta_keywords")

	input := domain.CreateBlogInput{
		Title:           title,
		Content:         content,
		Excerpt:         excerpt,
		IsPublished:     isPublished,
		MetaTitle:       metaTitle,
		MetaDescription: metaDescription,
		MetaKeywords:    metaKeywords,
	}

	blog, err := h.blogService.Create(r.Context(), input, user.ID)
	if err != nil {
		// In a real app we'd re-render the form with errors
		h.Error(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to create blog: %v", err))
		return
	}

	// Handle cover image upload if provided
	file, header, err := r.FormFile("cover_image")
	if err == nil {
		defer file.Close()

		imageData, _, _, err := service.ProcessImageUpload(file, header)
		if err != nil {
			// Log error but don't fail the blog creation
			fmt.Printf("Cover image upload failed: %v\n", err)
		} else {
			// Update blog with cover image
			updateInput := domain.UpdateBlogInput{
				CoverImage: imageData,
			}

			if _, err := h.blogService.Update(r.Context(), blog.ID, updateInput); err != nil {
				fmt.Printf("Failed to save cover image: %v\n", err)
			}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/a/blogs/%s/edit", blog.ID), http.StatusSeeOther)
}

func (h *BlogHandler) EditPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	blog, err := h.blogService.GetByID(r.Context(), id)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	theme, themeEnabled := h.GetTheme(r)
	oauthEnabled := h.GetOAuthEnabled(r)

	h.RenderTempl(w, r, admin.ReactBlogEditor("Edit Blog", blog, user, theme, themeEnabled, oauthEnabled))
}

func (h *BlogHandler) Edit(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid form data")
		return
	}

	title := r.FormValue("title")
	slug := r.FormValue("slug")
	content := r.FormValue("content")
	excerpt := r.FormValue("excerpt")
	isPublished := r.FormValue("is_published") == "on"

	// SEO metadata
	metaTitle := r.FormValue("meta_title")
	metaDescription := r.FormValue("meta_description")
	metaKeywords := r.FormValue("meta_keywords")

	input := domain.UpdateBlogInput{
		Title:           &title,
		Slug:            &slug,
		Content:         &content,
		Excerpt:         &excerpt,
		IsPublished:     &isPublished,
		MetaTitle:       &metaTitle,
		MetaDescription: &metaDescription,
		MetaKeywords:    &metaKeywords,
	}

	// Check if user wants to remove cover image
	removeCover := r.FormValue("remove_cover_image") == "true"
	if removeCover {
		fmt.Println("Cover image removal requested")
		input.RemoveCoverImage = true
	}

	// Handle cover image upload if provided
	file, header, err := r.FormFile("cover_image")
	if err == nil {
		fmt.Printf("File upload detected: %s, Size: %d\n", header.Filename, header.Size)
		defer file.Close()

		imageData, imageType, imageSize, err := service.ProcessImageUpload(file, header)
		if err != nil {
			fmt.Printf("Image processing error: %v\n", err)
			h.Error(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid cover image: %v", err))
			return
		}
		fmt.Printf("Image processed successfully. Type: %s, Size: %d\n", imageType, imageSize)

		input.CoverImage = imageData
	} else {
		fmt.Printf("No cover_image file found in request: %v\n", err)
	}

	_, err = h.blogService.Update(r.Context(), id, input)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to update blog: %v", err))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/a/blogs/%s/edit", id), http.StatusSeeOther)
}

// GetBlogJSON returns blog details as JSON (for API calls)
func (h *BlogHandler) GetBlogJSON(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}

	blog, err := h.blogService.GetByID(r.Context(), id)
	if err != nil || blog == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Simple JSON response with just what we need
	fmt.Fprintf(w, `{"id":"%s","cover_media_id":"%s"}`,
		blog.ID,
		func() string {
			if blog.CoverMediaID != nil {
				return blog.CoverMediaID.String()
			}
			return ""
		}())
}

func (h *BlogHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.blogService.Delete(r.Context(), id); err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to delete blog")
		return
	}

	// If HTMX, might just return empty or remove element
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/a/blogs", http.StatusSeeOther)
}
