-- Create Media Table
CREATE TABLE IF NOT EXISTS media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- Uploader
    filename VARCHAR(255),
    data BYTEA NOT NULL,
    content_type VARCHAR(50) NOT NULL,
    size_bytes INTEGER NOT NULL,
    alt_text TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_media_user_id ON media(user_id);
CREATE INDEX IF NOT EXISTS idx_media_content_type ON media(content_type);

-- Add reference columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_media_id UUID REFERENCES media(id) ON DELETE SET NULL;
ALTER TABLE blogs ADD COLUMN IF NOT EXISTS cover_media_id UUID REFERENCES media(id) ON DELETE SET NULL;
ALTER TABLE blog_images ADD COLUMN IF NOT EXISTS media_id UUID REFERENCES media(id) ON DELETE CASCADE;

-- Migrate Data (This part is tricky in pure SQL without procedural extensions if we want to handle data movement perfectly, 
-- but we will attempt a best-effort migration for the known schema)

-- 1. Migrate User Profile Images
DO $$
DECLARE
    r RECORD;
    mid UUID;
BEGIN
    -- Check if column exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='profile_image') THEN
        FOR r IN EXECUTE 'SELECT id, profile_image, profile_image_type, profile_image_size FROM users WHERE profile_image IS NOT NULL' LOOP
            INSERT INTO media (user_id, data, content_type, size_bytes, alt_text, filename)
            VALUES (r.id, r.profile_image, r.profile_image_type, r.profile_image_size, 'Profile Image', 'profile.jpg')
            RETURNING id INTO mid;
            
            UPDATE users SET profile_media_id = mid WHERE id = r.id;
        END LOOP;
    END IF;
END $$;

-- 2. Migrate Blog Cover Images
DO $$
DECLARE
    r RECORD;
    mid UUID;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='blogs' AND column_name='cover_image') THEN
        FOR r IN EXECUTE 'SELECT id, author_id, cover_image, cover_image_type, cover_image_size, title FROM blogs WHERE cover_image IS NOT NULL' LOOP
            INSERT INTO media (user_id, data, content_type, size_bytes, alt_text, filename)
            VALUES (r.author_id, r.cover_image, r.cover_image_type, r.cover_image_size, 'Cover Image for ' || r.title, 'cover.jpg')
            RETURNING id INTO mid;
            
            UPDATE blogs SET cover_media_id = mid WHERE id = r.id;
        END LOOP;
    END IF;
END $$;

-- 3. Migrate Blog Gallery Images
DO $$
DECLARE
    r RECORD;
    mid UUID;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='blog_images' AND column_name='image_data') THEN
        FOR r IN EXECUTE 'SELECT bi.id, bi.blog_id, bi.image_data, bi.image_type, bi.image_size, bi.alt_text, b.author_id FROM blog_images bi JOIN blogs b ON bi.blog_id = b.id' LOOP
            INSERT INTO media (user_id, data, content_type, size_bytes, alt_text, filename)
            VALUES (r.author_id, r.image_data, r.image_type, r.image_size, r.alt_text, 'gallery.jpg')
            RETURNING id INTO mid;
            
            UPDATE blog_images SET media_id = mid WHERE id = r.id;
        END LOOP;
    END IF;
END $$;

-- Drop old columns (SAFE TO RUN ONLY AFTER VERIFICATION)
ALTER TABLE users DROP COLUMN IF EXISTS profile_image;
ALTER TABLE users DROP COLUMN IF EXISTS profile_image_type;
ALTER TABLE users DROP COLUMN IF EXISTS profile_image_size;

ALTER TABLE blogs DROP COLUMN IF EXISTS cover_image;
ALTER TABLE blogs DROP COLUMN IF EXISTS cover_image_type;
ALTER TABLE blogs DROP COLUMN IF EXISTS cover_image_size;

ALTER TABLE blog_images DROP COLUMN IF EXISTS image_data;
ALTER TABLE blog_images DROP COLUMN IF EXISTS image_type;
ALTER TABLE blog_images DROP COLUMN IF EXISTS image_size;
-- alt_text is specific to the usage in the blog, but we also have it in media. 
-- We'll keep it in blog_images as an override or specific context.
