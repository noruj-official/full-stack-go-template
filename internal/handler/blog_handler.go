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
	blogService      *service.BlogService
	blogImageService *service.BlogImageService
}

func NewBlogHandler(base *Handler, blogService *service.BlogService, blogImageService *service.BlogImageService) *BlogHandler {
	return &BlogHandler{
		Handler:          base,
		blogService:      blogService,
		blogImageService: blogImageService,
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

	imageData, imageType, err := h.blogImageService.GetCoverImageBySlug(r.Context(), slug)
	if err != nil || imageData == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", imageType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	w.Write(imageData)
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

	// Check if user selected an existing gallery image as cover
	galleryImageID := r.FormValue("cover_image_gallery_id")
	if galleryImageID != "" {
		imageID, err := uuid.Parse(galleryImageID)
		if err == nil {
			fmt.Printf("Setting cover from gallery image: %s\n", galleryImageID)
			// We need to get the media_id from the blog_image
			// First, get the blog_image to find its media_id
			blogImage, err := h.blogImageService.GetByID(r.Context(), imageID)
			if err == nil && blogImage != nil {
				fmt.Printf("Found blog image, media_id: %s\n", blogImage.MediaID)
				input.CoverMediaID = &blogImage.MediaID
			} else {
				fmt.Printf("Error getting blog image: %v\n", err)
			}
		}
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

// Gallery Image Endpoints

func (h *BlogHandler) UploadGalleryImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	blogID, err := uuid.Parse(idStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid blog ID")
		return
	}

	// Verify blog exists
	blog, err := h.blogService.GetByID(r.Context(), blogID)
	if err != nil || blog == nil {
		h.Error(w, r, http.StatusNotFound, "Blog not found")
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid form data")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "No image file provided")
		return
	}
	defer file.Close()

	imageData, imageType, _, err := service.ProcessImageUpload(file, header)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid image: %v", err))
		return
	}

	altText := r.FormValue("alt_text")
	caption := r.FormValue("caption")
	position, _ := strconv.Atoi(r.FormValue("position"))

	input := domain.CreateBlogImageInput{
		BlogID:      blogID,
		ImageData:   imageData,
		ContentType: imageType,
		// ImageSize: imageSize, // calculated in service/repo? Input doesn't have ImageSize field anymore?
		// Wait, CreateBlogImageInput in step 241 has: ImageData, ContentType, AltText... No ImageSize?
		// Let me check step 241 output carefully.
		// "ContentType string", "// Size calculated from data"
		// "AltText", "Caption", "Position".
		// So ImageSize is REMOVED from input.
		AltText:  altText,
		Caption:  caption,
		Position: position,
	}

	img, err := h.blogImageService.Upload(r.Context(), input)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to upload image: %v", err))
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":"%s","image_type":"%s","image_size":%d,"alt_text":"%s","caption":"%s","position":%d}`,
		img.ID, imageType, len(imageData), img.AltText, img.Caption, img.Position)
}

func (h *BlogHandler) ListGalleryImages(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	blogID, err := uuid.Parse(idStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid blog ID")
		return
	}

	images, err := h.blogImageService.List(r.Context(), blogID)
	if err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to load images")
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	fmt.Printf("[DEBUG] ListGalleryImages: Returning %d images for blog %s\n", len(images), blogID)
	fmt.Fprint(w, "[")
	for i, img := range images {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		// For list, we might not have media type/size if not joined?
		// GetByIDWithoutData was used.
		// MediaID is available.
		// We can return media_id or constructed URL.
		// The json response expected: id, image_type, image_size...
		// If we don't have them, we might send empty or fetch?
		// For listing, performance matters.
		// Maybe just send ID and let frontend fetch image?
		// "image_type" and "image_size" might be null or we need to fetch media metadata?
		// For now, let's just return ID and basic info.
		// Frontend uses src={`/gallery/${img.id}`} so it works.
		// image_type/size are informative.
		fmt.Fprintf(w, `{"id":"%s","alt_text":"%s","caption":"%s","position":%d}`,
			img.ID, img.AltText, img.Caption, img.Position)
	}
	fmt.Fprint(w, "]")
}

func (h *BlogHandler) GetGalleryImage(w http.ResponseWriter, r *http.Request) {
	imageIDStr := r.PathValue("imageId")
	fmt.Printf("[DEBUG] GetGalleryImage: Request for %s\n", imageIDStr)
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		fmt.Printf("[DEBUG] GetGalleryImage: Invalid UUID %v\n", err)
		http.NotFound(w, r)
		return
	}

	imageData, contentType, err := h.blogImageService.GetImageData(r.Context(), imageID)
	if err != nil {
		fmt.Printf("[DEBUG] GetGalleryImage: Failed to get data %v\n", err)
		http.NotFound(w, r)
		return
	}

	fmt.Printf("[DEBUG] GetGalleryImage: Serving %d bytes, type %s\n", len(imageData), contentType)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	w.Write(imageData)
}

func (h *BlogHandler) DeleteGalleryImage(w http.ResponseWriter, r *http.Request) {
	imageIDStr := r.PathValue("imageId")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		h.Error(w, r, http.StatusBadRequest, "Invalid image ID")
		return
	}

	if err := h.blogImageService.Delete(r.Context(), imageID); err != nil {
		h.Error(w, r, http.StatusInternalServerError, "Failed to delete image")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"success":true}`)
}

// GetMediaImage serves an image directly by media ID
func (h *BlogHandler) GetMediaImage(w http.ResponseWriter, r *http.Request) {
	mediaIDStr := r.PathValue("mediaId")
	mediaID, err := uuid.Parse(mediaIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Use the blog image service's media service to get the media
	// We need to access the media service, but it's not directly available
	// Let's use the blog service instead
	// Actually, we need to fetch from the media table directly
	// For now, let me add a simple method

	// Get media through blog image service
	media, err := h.blogImageService.GetMediaByID(r.Context(), mediaID)
	if err != nil {
		fmt.Printf("[DEBUG] GetMediaImage: Failed to get media %v\n", err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", media.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(media.Data)
}
