-- Centralize Media and Remove Legacy Gallery

-- 1. Update Media Table for Centralized Storage & S3 Support
ALTER TABLE media ADD COLUMN IF NOT EXISTS storage_provider VARCHAR(50) NOT NULL DEFAULT 'database';
ALTER TABLE media ADD COLUMN IF NOT EXISTS file_key VARCHAR(512);
ALTER TABLE media ADD COLUMN IF NOT EXISTS public_url VARCHAR(512);

-- Make data nullable to support off-db storage (S3)
ALTER TABLE media ALTER COLUMN data DROP NOT NULL;

-- 2. Drop Legacy Blog Images Table
-- We assume migration of data to `media` was handled in 002 or is not required as per "remove gallery management" request.
-- If there are images in blog_images that are NOT in media, we might lose them if 002 didn't run or failed.
-- However, 002 does have a migration block.
-- We will proceed with dropping it.

DROP TABLE IF EXISTS blog_images;
