import React, { useState, useEffect } from 'react'
import { X, Upload, Check } from 'lucide-react'

const ImagePicker = ({ blogId, onSelect, onClose }) => {
    const [images, setImages] = useState([])
    const [isLoading, setIsLoading] = useState(true)
    const [selectedImage, setSelectedImage] = useState(null)

    useEffect(() => {
        if (blogId) {
            fetchImages()
        }
    }, [blogId])

    const fetchImages = async () => {
        try {
            setIsLoading(true)
            const response = await fetch(`/a/blogs/${blogId}/gallery`)
            if (!response.ok) throw new Error('Failed to fetch images')
            const data = await response.json()
            setImages(data)
        } catch (err) {
            console.error('Error fetching images:', err)
        } finally {
            setIsLoading(false)
        }
    }

    const handleSelect = (image) => {
        setSelectedImage(image)
    }

    const handleConfirm = () => {
        if (selectedImage) {
            onSelect(`/gallery/${selectedImage.id}`)
            onClose()
        }
    }

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={(e) => {
            if (e.target === e.currentTarget) onClose()
        }}>
            <div className="bg-base-100 rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-base-300">
                    <h3 className="text-lg font-semibold">Select Image from Gallery</h3>
                    <button
                        type="button"
                        onClick={onClose}
                        className="btn btn-sm btn-circle btn-ghost"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-4">
                    {isLoading ? (
                        <div className="flex items-center justify-center py-12">
                            <span className="loading loading-spinner loading-lg"></span>
                        </div>
                    ) : images.length === 0 ? (
                        <div className="text-center py-12">
                            <Upload className="w-12 h-12 mx-auto text-base-content/30 mb-3" />
                            <p className="text-base-content/60">No images in gallery yet.</p>
                            <p className="text-sm text-base-content/40 mt-1">Upload images using the Gallery Management section below.</p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                            {images.map((img) => (
                                <div
                                    key={img.id}
                                    className={`relative aspect-square bg-base-200 rounded-lg overflow-hidden border-2 cursor-pointer transition-all hover:shadow-lg ${selectedImage?.id === img.id
                                            ? 'border-primary ring-2 ring-primary/50'
                                            : 'border-base-300 hover:border-primary/50'
                                        }`}
                                    onClick={() => handleSelect(img)}
                                >
                                    <img
                                        src={`/gallery/${img.id}`}
                                        alt={img.alt_text || 'Gallery image'}
                                        className="w-full h-full object-cover"
                                    />
                                    {selectedImage?.id === img.id && (
                                        <div className="absolute inset-0 bg-primary/20 flex items-center justify-center">
                                            <div className="bg-primary text-primary-content rounded-full p-2">
                                                <Check className="w-6 h-6" />
                                            </div>
                                        </div>
                                    )}
                                    {img.alt_text && (
                                        <div className="absolute bottom-0 left-0 right-0 bg-black/60 text-white text-xs p-2 truncate">
                                            {img.alt_text}
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="flex items-center justify-end gap-2 p-4 border-t border-base-300">
                    <button
                        type="button"
                        onClick={onClose}
                        className="btn btn-ghost"
                    >
                        Cancel
                    </button>
                    <button
                        type="button"
                        onClick={handleConfirm}
                        disabled={!selectedImage}
                        className="btn btn-primary"
                    >
                        Insert Image
                    </button>
                </div>
            </div>
        </div>
    )
}

export default ImagePicker
