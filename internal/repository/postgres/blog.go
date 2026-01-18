package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noruj-official/full-stack-go-template/internal/domain"
)

type BlogRepository struct {
	db *DB
}

func NewBlogRepository(db *DB) *BlogRepository {
	return &BlogRepository{db: db}
}

func (r *BlogRepository) Create(ctx context.Context, blog *domain.Blog) error {
	query := `
		INSERT INTO blogs (id, title, slug, content, excerpt, author_id, is_published, published_at, 
			created_at, updated_at, cover_media_id,
			meta_title, meta_description, meta_keywords, og_image, og_image_type, og_image_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		blog.ID, blog.Title, blog.Slug, blog.Content, blog.Excerpt, blog.AuthorID,
		blog.IsPublished, blog.PublishedAt, blog.CreatedAt, blog.UpdatedAt,
		blog.CoverMediaID,
		blog.MetaTitle, blog.MetaDescription, blog.MetaKeywords,
		blog.OGImage, blog.OGImageType, blog.OGImageSize,
	)
	return err
}

func (r *BlogRepository) Update(ctx context.Context, blog *domain.Blog) error {
	query := `
		UPDATE blogs
		SET title = $1, slug = $2, content = $3, excerpt = $4, is_published = $5, published_at = $6, updated_at = $7,
			cover_media_id = $8,
			meta_title = $9, meta_description = $10, meta_keywords = $11,
			og_image = $12, og_image_type = $13, og_image_size = $14
		WHERE id = $15
	`
	_, err := r.db.Pool.Exec(ctx, query,
		blog.Title, blog.Slug, blog.Content, blog.Excerpt, blog.IsPublished, blog.PublishedAt, blog.UpdatedAt,
		blog.CoverMediaID,
		blog.MetaTitle, blog.MetaDescription, blog.MetaKeywords,
		blog.OGImage, blog.OGImageType, blog.OGImageSize,
		blog.ID,
	)
	return err
}

func (r *BlogRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM blogs WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *BlogRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Blog, error) {
	query := `
		SELECT b.id, b.title, b.slug, b.content, b.excerpt, b.author_id, b.is_published, b.published_at, b.created_at, b.updated_at,
		       b.cover_media_id, b.meta_title, b.meta_description, b.meta_keywords,
		       b.og_image_type, b.og_image_size,
		       u.id, u.name, u.email, u.profile_media_id IS NOT NULL as has_image
		FROM blogs b
		JOIN users u ON b.author_id = u.id
		WHERE b.id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)
	return scanBlog(row)
}

func (r *BlogRepository) GetBySlug(ctx context.Context, slug string) (*domain.Blog, error) {
	query := `
		SELECT b.id, b.title, b.slug, b.content, b.excerpt, b.author_id, b.is_published, b.published_at, b.created_at, b.updated_at,
		       b.cover_media_id, b.meta_title, b.meta_description, b.meta_keywords,
		       b.og_image_type, b.og_image_size,
		       u.id, u.name, u.email, u.profile_media_id IS NOT NULL as has_image
		FROM blogs b
		JOIN users u ON b.author_id = u.id
		WHERE b.slug = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, slug)
	return scanBlog(row)
}

func (r *BlogRepository) List(ctx context.Context, filter domain.BlogFilter) ([]*domain.Blog, int, error) {
	var where []string
	var args []interface{}
	argIdx := 1

	if filter.IsPublished != nil {
		where = append(where, fmt.Sprintf("b.is_published = $%d", argIdx))
		args = append(args, *filter.IsPublished)
		argIdx++
	}

	if filter.AuthorID != nil {
		where = append(where, fmt.Sprintf("b.author_id = $%d", argIdx))
		args = append(args, *filter.AuthorID)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM blogs b %s", whereClause)
	var total int
	if err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT b.id, b.title, b.slug, b.content, b.excerpt, b.author_id, b.is_published, b.published_at, b.created_at, b.updated_at,
		       b.cover_media_id, b.meta_title, b.meta_description, b.meta_keywords,
		       b.og_image_type, b.og_image_size,
		       u.id, u.name, u.email, u.profile_media_id IS NOT NULL as has_image
		FROM blogs b
		JOIN users u ON b.author_id = u.id
		%s
		ORDER BY b.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var blogs []*domain.Blog
	for rows.Next() {
		blog, err := scanBlogRow(rows)
		if err != nil {
			return nil, 0, err
		}
		blogs = append(blogs, blog)
	}

	return blogs, total, nil
}

func scanBlog(row pgx.Row) (*domain.Blog, error) {
	var b domain.Blog
	var u domain.User
	var hasImage bool
	var metaTitle, metaDescription, metaKeywords, ogImageType sql.NullString
	var ogImageSize sql.NullInt32

	err := row.Scan(
		&b.ID, &b.Title, &b.Slug, &b.Content, &b.Excerpt, &b.AuthorID, &b.IsPublished, &b.PublishedAt, &b.CreatedAt, &b.UpdatedAt,
		&b.CoverMediaID, &metaTitle, &metaDescription, &metaKeywords,
		&ogImageType, &ogImageSize,
		&u.ID, &u.Name, &u.Email, &hasImage,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Or return a specific ErrNotFound
		}
		return nil, err
	}

	// Convert nullable types
	if metaTitle.Valid {
		b.MetaTitle = metaTitle.String
	}
	if metaDescription.Valid {
		b.MetaDescription = metaDescription.String
	}
	if metaKeywords.Valid {
		b.MetaKeywords = metaKeywords.String
	}
	if ogImageType.Valid {
		b.OGImageType = ogImageType.String
	}
	if ogImageSize.Valid {
		b.OGImageSize = int(ogImageSize.Int32)
	}

	b.Author = &u
	return &b, nil
}

func scanBlogRow(rows pgx.Rows) (*domain.Blog, error) {
	var b domain.Blog
	var u domain.User
	var hasImage bool
	var metaTitle, metaDescription, metaKeywords, ogImageType sql.NullString
	var ogImageSize sql.NullInt32

	err := rows.Scan(
		&b.ID, &b.Title, &b.Slug, &b.Content, &b.Excerpt, &b.AuthorID, &b.IsPublished, &b.PublishedAt, &b.CreatedAt, &b.UpdatedAt,
		&b.CoverMediaID, &metaTitle, &metaDescription, &metaKeywords,
		&ogImageType, &ogImageSize,
		&u.ID, &u.Name, &u.Email, &hasImage,
	)
	if err != nil {
		return nil, err
	}

	// Convert nullable types
	if metaTitle.Valid {
		b.MetaTitle = metaTitle.String
	}
	if metaDescription.Valid {
		b.MetaDescription = metaDescription.String
	}
	if metaKeywords.Valid {
		b.MetaKeywords = metaKeywords.String
	}
	if ogImageType.Valid {
		b.OGImageType = ogImageType.String
	}
	if ogImageSize.Valid {
		b.OGImageSize = int(ogImageSize.Int32)
	}

	b.Author = &u
	return &b, nil
}
