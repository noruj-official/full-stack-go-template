import React, { useState, useEffect, useRef } from 'react'
import { Trash2, Copy, Upload, Image as ImageIcon, X } from 'lucide-react'

const GalleryManager = ({ blogId }) => {
    const [images, setImages] = useState([])
    const [isUploading, setIsUploading] = useState(false)
    const [error, setError] = useState(null)
    const fileInputRef = useRef(null)

    useEffect(() => {
        if (blogId) {
            loadAllImages()
        }
    }, [blogId])

    const fetchImages = async () => {
        try {
            const response = await fetch(`/a/blogs/${blogId}/gallery`)
            if (!response.ok) throw new Error('Failed to fetch images')
            const data = await response.json()
            setImages(data)
        } catch (err) {
            console.error('Error fetching images:', err)
            setError('Failed to load gallery images')
        }
    }

    const fetchBlogCover = async () => {
        try {
            // Fetch blog details to get cover image
            const response = await fetch(`/a/blogs/${blogId}`)
            if (response.ok) {
                const blogData = await response.json()
                if (blogData.cover_media_id) {
                    // Add cover image to the beginning of images array with special flag
                    return {
                        id: 'cover-' + blogData.cover_media_id,
                        mediaId: blogData.cover_media_id,
                        alt_text: 'Cover Image',
                        isCover: true
                    }
                }
            }
        } catch (err) {
            console.error('Error fetching cover:', err)
        }
        return null
    }

    const loadAllImages = async () => {
        await fetchImages()
        const coverImg = await fetchBlogCover()
        if (coverImg) {
            setImages(prev => [coverImg, ...prev.filter(img => !img.isCover)])
        }
    }

    const handleUpload = async (e) => {
        const file = e.target.files[0]
        if (!file) return

        setIsUploading(true)
        setError(null)
        const formData = new FormData()
        formData.append('image', file)
        formData.append('position', images.length)

        try {
            const response = await fetch(`/a/blogs/${blogId}/gallery`, {
                method: 'POST',
                body: formData,
            })

            if (!response.ok) {
                const errorData = await response.text()
                throw new Error(errorData || 'Failed to upload image')
            }

            // Refresh all images to show the new upload
            await loadAllImages()
            if (fileInputRef.current) {
                fileInputRef.current.value = ''
            }
        } catch (err) {
            console.error('Error uploading image:', err)
            setError(err.message)
        } finally {
            setIsUploading(false)
        }
    }

    const handleDelete = async (imageId) => {
        if (!confirm('Are you sure you want to delete this image?')) return

        try {
            const response = await fetch(`/a/blogs/${blogId}/gallery/${imageId}`, {
                method: 'DELETE',
            })

            if (!response.ok) throw new Error('Failed to delete image')

            setImages(images.filter(img => img.id !== imageId))
        } catch (err) {
            console.error('Error deleting image:', err)
            setError('Failed to delete image')
        }
    }

    const copyToClipboard = (url) => {
        navigator.clipboard.writeText(url)
        // Could add a toast or tooltip here for feedback
    }

    if (!blogId) {
        return (
            <div className="alert alert-info shadow-sm">
                <ImageIcon className="w-5 h-5" />
                <span>Save the blog post first to manage the gallery.</span>
            </div>
        )
    }

    return (
        <div className="space-y-4">
            <h3 className="text-lg font-semibold flex items-center gap-2">
                <ImageIcon className="w-5 h-5" />
                Gallery Management
            </h3>

            {error && (
                <div className="alert alert-error text-sm py-2 shadow-sm rounded-md">
                    <X className="w-4 h-4 cursor-pointer" onClick={() => setError(null)} />
                    <span>{error}</span>
                </div>
            )}

            {/* Upload Area */}
            <div className="border-2 border-dashed border-base-300 rounded-lg p-6 text-center hover:border-primary/50 transition-colors bg-base-100">
                <input
                    type="file"
                    accept="image/*"
                    onChange={handleUpload}
                    className="hidden"
                    ref={fileInputRef}
                    disabled={isUploading}
                />
                <div
                    className="flex flex-col items-center gap-2 cursor-pointer"
                    onClick={() => fileInputRef.current?.click()}
                >
                    <Upload className={`w-8 h-8 text-base-content/50 ${isUploading ? 'animate-bounce' : ''}`} />
                    <span className="text-sm font-medium">
                        {isUploading ? 'Uploading...' : 'Click to upload gallery image'}
                    </span>
                    <span className="text-xs text-base-content/40">Max 5MB (JPEG, PNG, WebP)</span>
                </div>
            </div>

            {/* Image Grid */}
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {images.map((img) => (
                    <div key={img.id} className="group relative aspect-square bg-base-200 rounded-lg overflow-hidden border border-base-200">
                        <img
                            src={img.isCover ? `/media/${img.mediaId}` : `/gallery/${img.id}`}
                            alt={img.alt_text || 'Gallery image'}
                            className="w-full h-full object-cover"
                        />

                        {/* Cover Badge */}
                        {img.isCover && (
                            <div className="absolute top-2 left-2 bg-primary text-primary-content px-2 py-1 rounded text-xs font-semibold">
                                Cover
                            </div>
                        )}

                        {/* Overlay Actions */}
                        <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                            <button
                                type="button"
                                onClick={() => copyToClipboard(img.isCover ? `/media/${img.mediaId}` : `/gallery/${img.id}`)}
                                className="btn btn-sm btn-circle btn-ghost text-white hover:bg-white/20"
                                title="Copy URL"
                            >
                                <Copy className="w-4 h-4" />
                            </button>
                            {!img.isCover && (
                                <button
                                    type="button"
                                    onClick={() => handleDelete(img.id)}
                                    className="btn btn-sm btn-circle btn-ghost text-error hover:bg-red-500/20"
                                    title="Delete Image"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            )}
                        </div>
                    </div>
                ))}
            </div>

            {images.length === 0 && !isUploading && (
                <div className="text-center text-sm text-base-content/50 py-4 italic">
                    No images in gallery yet.
                </div>
            )}
        </div>
    )
}

export default GalleryManager
