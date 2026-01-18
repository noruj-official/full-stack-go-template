import React, { useState, useCallback, useRef } from 'react'
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import Youtube from '@tiptap/extension-youtube'
import TextAlign from '@tiptap/extension-text-align'
import Underline from '@tiptap/extension-underline'
import { Table, TableCell, TableHeader, TableRow } from '@tiptap/extension-table'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Subscript from '@tiptap/extension-subscript'
import Superscript from '@tiptap/extension-superscript'
import Highlight from '@tiptap/extension-highlight'

import {
    Bold, Italic, Strikethrough, Underline as UnderlineIcon,
    Heading1, Heading2, List, ListOrdered, Quote, Code,
    Undo, Redo, ArrowLeft, Link as LinkIcon, Unlink,
    Image as ImageIcon, Youtube as YoutubeIcon,
    AlignLeft, AlignCenter, AlignRight,
    Table as TableIcon, Columns, Rows, Trash2,
    CheckSquare, Superscript as SuperscriptIcon, Subscript as SubscriptIcon,
    Highlighter, Minus, Upload, Eye, Settings, Type, Globe, Save, FileText, Hash, Layout
} from 'lucide-react'

// Lowlight setup for syntax highlighting
const lowlight = createLowlight(common)

// --- Reusable Components ---

const FormField = ({ label, subLabel, children, required, error, icon: Icon }) => (
    <div className="form-control w-full">
        <label className="label">
            <span className="label-text font-bold text-base flex items-center gap-2">
                {Icon && <Icon className="w-4 h-4 text-primary" />}
                {label}
                {required && <span className="text-error">*</span>}
            </span>
            {subLabel && <span className="label-text-alt text-base-content/60">{subLabel}</span>}
        </label>
        {children}
        {error && <label className="label"><span className="label-text-alt text-error">{error}</span></label>}
    </div>
)

const MenuBar = ({ editor }) => {
    const fileInputRef = useRef(null)

    if (!editor) {
        return null
    }

    const setLink = useCallback(() => {
        const previousUrl = editor.getAttributes('link').href
        const url = window.prompt('URL', previousUrl)

        if (url === null) return
        if (url === '') {
            editor.chain().focus().extendMarkRange('link').unsetLink().run()
            return
        }
        editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
    }, [editor])

    const handleImageUpload = async (e) => {
        const file = e.target.files[0]
        if (!file) return

        const formData = new FormData()
        formData.append('file', file)

        try {
            const response = await fetch('/api/media/upload', {
                method: 'POST',
                body: formData,
            })

            if (!response.ok) {
                throw new Error('Upload failed')
            }

            const data = await response.json()
            if (data.url) {
                editor.chain().focus().setImage({ src: data.url, alt: data.filename }).run()
            }
        } catch (error) {
            console.error('Error uploading image:', error)
            alert('Failed to upload image')
        } finally {
            // Reset input
            e.target.value = ''
        }
    }

    const triggerImageUpload = () => {
        fileInputRef.current?.click()
    }

    const addYoutube = useCallback(() => {
        const url = window.prompt('Enter YouTube URL')
        if (url) {
            editor.commands.setYoutubeVideo({ src: url })
        }
    }, [editor])

    return (
        <div className="bg-base-100 border-b border-base-200 p-2 flex flex-wrap gap-1 sticky top-0 z-10 rounded-t-lg items-center">
            <input
                type="file"
                ref={fileInputRef}
                onChange={handleImageUpload}
                className="hidden"
                accept="image/*"
            />

            {/* History */}
            <div className="join mr-2">
                <button type="button" onClick={() => editor.chain().focus().undo().run()} disabled={!editor.can().chain().focus().undo().run()} className="join-item btn btn-ghost btn-sm btn-square" title="Undo">
                    <Undo className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().redo().run()} disabled={!editor.can().chain().focus().redo().run()} className="join-item btn btn-ghost btn-sm btn-square" title="Redo">
                    <Redo className="w-4 h-4" />
                </button>
            </div>

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Formatting */}
            <div className="join mx-2">
                <button type="button" onClick={() => editor.chain().focus().toggleBold().run()} disabled={!editor.can().chain().focus().toggleBold().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('bold') ? 'bg-base-200 text-primary' : ''}`} title="Bold">
                    <Bold className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleItalic().run()} disabled={!editor.can().chain().focus().toggleItalic().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('italic') ? 'bg-base-200 text-primary' : ''}`} title="Italic">
                    <Italic className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleUnderline().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('underline') ? 'bg-base-200 text-primary' : ''}`} title="Underline">
                    <UnderlineIcon className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleStrike().run()} disabled={!editor.can().chain().focus().toggleStrike().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('strike') ? 'bg-base-200 text-primary' : ''}`} title="Strike">
                    <Strikethrough className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleHighlight().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('highlight') ? 'bg-base-200 text-primary' : ''}`} title="Highlight">
                    <Highlighter className="w-4 h-4" />
                </button>
            </div>

            <div className="join mr-2">
                <button type="button" onClick={() => editor.chain().focus().toggleSuperscript().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('superscript') ? 'bg-base-200 text-primary' : ''}`} title="Superscript">
                    <SuperscriptIcon className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleSubscript().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('subscript') ? 'bg-base-200 text-primary' : ''}`} title="Subscript">
                    <SubscriptIcon className="w-4 h-4" />
                </button>
            </div>

            <div className="join mr-2">
                <button type="button" onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('heading', { level: 1 }) ? 'bg-base-200 text-primary' : ''}`} title="H1">
                    <Heading1 className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('heading', { level: 2 }) ? 'bg-base-200 text-primary' : ''}`} title="H2">
                    <Heading2 className="w-4 h-4" />
                </button>
            </div>

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Alignment */}
            <div className="join mx-2">
                <button type="button" onClick={() => editor.chain().focus().setTextAlign('left').run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive({ textAlign: 'left' }) ? 'bg-base-200 text-primary' : ''}`} title="Left">
                    <AlignLeft className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().setTextAlign('center').run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive({ textAlign: 'center' }) ? 'bg-base-200 text-primary' : ''}`} title="Center">
                    <AlignCenter className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().setTextAlign('right').run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive({ textAlign: 'right' }) ? 'bg-base-200 text-primary' : ''}`} title="Right">
                    <AlignRight className="w-4 h-4" />
                </button>
            </div>

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Lists */}
            <div className="join mx-2">
                <button type="button" onClick={() => editor.chain().focus().toggleBulletList().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('bulletList') ? 'bg-base-200 text-primary' : ''}`} title="Bullet List">
                    <List className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleOrderedList().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('orderedList') ? 'bg-base-200 text-primary' : ''}`} title="Ordered List">
                    <ListOrdered className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleTaskList().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('taskList') ? 'bg-base-200 text-primary' : ''}`} title="Task List">
                    <CheckSquare className="w-4 h-4" />
                </button>
            </div>

            {/* Inserts */}
            <div className="join">
                <button type="button" onClick={setLink} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('link') ? 'bg-base-200 text-primary' : ''}`} title="Link">
                    <LinkIcon className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().unsetLink().run()} disabled={!editor.isActive('link')} className="join-item btn btn-ghost btn-sm btn-square" title="Unlink">
                    <Unlink className="w-4 h-4" />
                </button>
                <button type="button" onClick={triggerImageUpload} className="join-item btn btn-ghost btn-sm btn-square" title="Upload Image">
                    <Upload className="w-4 h-4" />
                </button>
                <button type="button" onClick={addYoutube} className="join-item btn btn-ghost btn-sm btn-square" title="Youtube">
                    <YoutubeIcon className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleBlockquote().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('blockquote') ? 'bg-base-200 text-primary' : ''}`} title="Quote">
                    <Quote className="w-4 h-4" />
                </button>
                <button type="button" onClick={() => editor.chain().focus().toggleCodeBlock().run()} className={`join-item btn btn-ghost btn-sm btn-square ${editor.isActive('codeBlock') ? 'bg-base-200 text-primary' : ''}`} title="Code">
                    <Code className="w-4 h-4" />
                </button>
                <button
                    type="button"
                    onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()}
                    className="join-item btn btn-ghost btn-sm btn-square"
                    title="Insert Table"
                >
                    <TableIcon className="w-4 h-4" />
                </button>
            </div>

            {editor.isActive('table') && (
                <div className="join ml-2 bg-base-200 rounded-btn">
                    <button type="button" onClick={() => editor.chain().focus().addColumnAfter().run()} className="join-item btn btn-ghost btn-xs btn-square" title="Add Column">
                        <Columns className="w-3 h-3" />
                    </button>
                    <button type="button" onClick={() => editor.chain().focus().addRowAfter().run()} className="join-item btn btn-ghost btn-xs btn-square" title="Add Row">
                        <Rows className="w-3 h-3" />
                    </button>
                    <button type="button" onClick={() => editor.chain().focus().deleteTable().run()} className="join-item btn btn-ghost btn-xs btn-square text-error" title="Delete Table">
                        <Trash2 className="w-3 h-3" />
                    </button>
                </div>
            )}

            <div className="flex-1"></div>

            {editor.isActive('codeBlock') && (
                <select
                    className="select select-bordered select-xs ml-2 max-w-[100px]"
                    onChange={event => editor.chain().focus().setCodeBlock({ language: event.target.value }).run()}
                    value={editor.getAttributes('codeBlock').language || ''}
                >
                    <option value="null">auto</option>
                    <option disabled>—</option>
                    {lowlight.listLanguages().map((lang, index) => (
                        <option key={index} value={lang}>{lang}</option>
                    ))}
                </select>
            )}
        </div>
    )
}

const BlogEditor = ({ initialData, actionUrl }) => {
    const [activeTab, setActiveTab] = useState('write')
    const [saveStatus, setSaveStatus] = useState('idle') // idle, saving, saved, error

    const [title, setTitle] = useState(initialData?.title || '')
    const [slug, setSlug] = useState(initialData?.slug || '')
    const [excerpt, setExcerpt] = useState(initialData?.excerpt || '')
    const [content, setContent] = useState(initialData?.content || '')
    const [isPublished, setIsPublished] = useState(initialData?.is_published || false)

    // SEO metadata states
    const [metaTitle, setMetaTitle] = useState(initialData?.meta_title || '')
    const [metaDescription, setMetaDescription] = useState(initialData?.meta_description || '')
    const [metaKeywords, setMetaKeywords] = useState(initialData?.meta_keywords || '')

    // Cover image state
    const [coverImageFile, setCoverImageFile] = useState(null)
    const [coverImagePreview, setCoverImagePreview] = useState(
        initialData?.has_cover_image ? `/blogs/${initialData.slug}/cover` : null
    )
    const [removeCoverImageFlag, setRemoveCoverImageFlag] = useState(false)
    const fileInputRef = useRef(null)

    const editor = useEditor({
        extensions: [
            StarterKit.configure({
                codeBlock: false, // We use CodeBlockLowlight instead
            }),
            Placeholder.configure({
                placeholder: 'Write your story here...',
            }),
            Link.configure({
                openOnClick: false,
                HTMLAttributes: { class: 'link link-primary' },
            }),
            Image.configure({
                HTMLAttributes: { class: 'rounded-lg max-w-full h-auto shadow-md my-4' },
            }),
            Youtube.configure({
                controls: false,
                HTMLAttributes: { class: 'w-full aspect-video rounded-lg shadow-md my-4' },
            }),
            TextAlign.configure({
                types: ['heading', 'paragraph'],
            }),
            Underline,
            Table.configure({
                resizable: true,
                HTMLAttributes: { class: 'table table-zebra w-full my-4 border border-base-300' },
            }),
            TableRow,
            TableHeader,
            TableCell,
            CodeBlockLowlight.configure({
                lowlight,
                HTMLAttributes: { class: 'rounded-lg border border-[#3e4451] bg-[#282c34] text-[#abb2bf] p-4 font-mono text-sm overflow-x-auto my-4' },
            }),
            TaskList.configure({
                HTMLAttributes: { class: 'not-prose pl-2' },
            }),
            TaskItem.configure({
                nested: true,
            }),
            Subscript,
            Superscript,
            Highlight.configure({
                multicolor: true,
            }),
        ],
        content: initialData?.content || '',
        onUpdate: ({ editor }) => {
            setContent(editor.getHTML())
        },
        editorProps: {
            attributes: {
                class: 'prose prose-sm sm:prose lg:prose-lg xl:prose-xl mx-auto focus:outline-none min-h-[300px] text-base-content p-4 font-serif', // Added font-serif for writing feel
            },
        },
    })

    const handleTitleChange = (e) => {
        const newTitle = e.target.value
        setTitle(newTitle)
        if (!initialData) {
            setSlug(newTitle.toLowerCase()
                .replace(/[^a-z0-9]+/g, '-')
                .replace(/(^-|-$)+/g, ''))
        }
    }

    const handleCoverImageChange = (e) => {
        const file = e.target.files[0]
        if (file) {
            if (!file.type.startsWith('image/')) {
                alert('Please select an image file')
                return
            }
            if (file.size > 5 * 1024 * 1024) {
                alert('Image size must not exceed 5MB')
                return
            }
            setCoverImageFile(file)
            const reader = new FileReader()
            reader.onloadend = () => {
                setCoverImagePreview(reader.result)
            }
            reader.readAsDataURL(file)
        }
    }

    const removeCoverImage = () => {
        setCoverImageFile(null)
        setCoverImagePreview(null)
        setRemoveCoverImageFlag(true)
        if (fileInputRef.current) {
            fileInputRef.current.value = ''
        }
    }

    return (
        <div className="min-h-screen pb-20">
            <div className="max-w-6xl mx-auto p-4 md:p-8">
                {/* Header */}
                <div className="flex items-center justify-between mb-8 animate-fade-in">
                    <div>
                        <h1 className="text-4xl font-extrabold tracking-tight text-base-content mb-2">
                            {initialData ? 'Edit Blog Post' : 'Create New Post'}
                        </h1>
                        {saveStatus !== 'idle' && (
                            <div className="flex items-center gap-2 text-sm mt-2">
                                {saveStatus === 'saving' && (
                                    <>
                                        <span className="loading loading-spinner loading-xs"></span>
                                        <span className="text-base-content/70">Saving...</span>
                                    </>
                                )}
                                {saveStatus === 'saved' && (
                                    <>
                                        <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                        </svg>
                                        <span className="text-success">Saved successfully</span>
                                    </>
                                )}
                                {saveStatus === 'error' && (
                                    <>
                                        <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                        </svg>
                                        <span className="text-error">Error saving</span>
                                    </>
                                )}
                            </div>
                        )}
                    </div>
                    <div className="flex gap-3">
                        <a href="/a/blogs" className="btn btn-ghost btn-lg gap-2 tooltip tooltip-bottom" data-tip="Return to blog list">
                            <ArrowLeft className="w-5 h-5" />
                            <span className="hidden sm:inline">Back</span>
                        </a>
                        <button
                            type="submit"
                            form="blog-form"
                            className="btn btn-primary btn-lg gap-2 shadow-xl"
                            disabled={saveStatus === 'saving'}
                        >
                            {saveStatus === 'saving' ? (
                                <span className="loading loading-spinner loading-sm"></span>
                            ) : (
                                <Save className="w-5 h-5" />
                            )}
                            {saveStatus === 'saving' ? 'Saving...' : 'Save Post'}
                        </button>
                    </div>
                </div>

                <form id="blog-form" action={actionUrl} method="POST" encType="multipart/form-data">
                    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">

                        {/* Sidebar / Tabs */}
                        <div className="lg:col-span-1 space-y-4 animate-slide-up">
                            <div className="bg-base-100 rounded-xl shadow-lg border border-base-200 overflow-hidden">
                                <ul className="menu menu-lg w-full p-0">
                                    <li>
                                        <a
                                            className={activeTab === 'write' ? 'active font-bold rounded-none border-l-4 border-primary bg-primary/5' : 'rounded-none border-l-4 border-transparent hover:border-primary/30'}
                                            onClick={() => setActiveTab('write')}
                                        >
                                            <Type className="w-5 h-5" /> Write
                                        </a>
                                    </li>
                                    <li>
                                        <a
                                            className={activeTab === 'settings' ? 'active font-bold rounded-none border-l-4 border-primary bg-primary/5' : 'rounded-none border-l-4 border-transparent hover:border-primary/30'}
                                            onClick={() => setActiveTab('settings')}
                                        >
                                            <Settings className="w-5 h-5" /> Settings
                                        </a>
                                    </li>
                                    <li>
                                        <a
                                            className={activeTab === 'seo' ? 'active font-bold rounded-none border-l-4 border-primary bg-primary/5' : 'rounded-none border-l-4 border-transparent hover:border-primary/30'}
                                            onClick={() => setActiveTab('seo')}
                                        >
                                            <Globe className="w-5 h-5" /> SEO
                                        </a>
                                    </li>
                                </ul>
                            </div>

                            {/* Quick Status Card */}
                            <div className="bg-base-100 rounded-xl shadow-lg border border-base-200 p-5">
                                <h3 className="text-sm font-bold text-base-content/50 uppercase mb-4 flex items-center gap-2">
                                    <Eye className="w-4 h-4" />
                                    Visibility
                                </h3>
                                <div className="form-control">
                                    <label className="label cursor-pointer justify-between p-0">
                                        <div className="flex items-center gap-2">
                                            <span className="label-text font-medium text-base">{isPublished ? 'Published' : 'Draft'}</span>
                                            {isPublished ? (
                                                <div className="badge badge-success badge-sm gap-1">
                                                    <span className="w-1.5 h-1.5 rounded-full bg-white"></span>
                                                    Live
                                                </div>
                                            ) : (
                                                <div className="badge badge-warning badge-sm">Private</div>
                                            )}
                                        </div>
                                        <input
                                            type="checkbox"
                                            name="is_published"
                                            checked={isPublished}
                                            onChange={(e) => setIsPublished(e.target.checked)}
                                            className="toggle toggle-success"
                                            value="on"
                                        />
                                    </label>
                                </div>
                                <p className="text-xs text-base-content/60 mt-3">
                                    {isPublished ? 'This post is visible to everyone' : 'Only admins can see this draft'}
                                </p>
                            </div>
                        </div>

                        {/* Main Content Area */}
                        <div className="lg:col-span-3 animate-slide-up" style={{ animationDelay: '0.1s' }}>
                            <div className="bg-base-100 rounded-xl shadow-xl border border-base-200 min-h-[600px] relative overflow-hidden">

                                {/* Write Tab */}
                                <div className={activeTab === 'write' ? 'block p-6 md:p-8' : 'hidden'}>
                                    <div className="max-w-4xl mx-auto space-y-6">
                                        <input
                                            type="text"
                                            name="title"
                                            value={title}
                                            onChange={handleTitleChange}
                                            placeholder="Enter your post title..."
                                            className="input input-ghost text-4xl md:text-5xl font-extrabold w-full h-auto px-0 border-0 focus:outline-none focus:bg-transparent placeholder:text-base-content/20 leading-tight"
                                            required
                                        />
                                        <div className="divider my-2"></div>
                                        <div className="bg-base-100 min-h-[500px]">
                                            <MenuBar editor={editor} />
                                            <EditorContent editor={editor} className="outline-none" />
                                        </div>
                                        <input type="hidden" name="content" value={content} />
                                    </div>
                                </div>

                                {/* Settings Tab */}
                                <div className={activeTab === 'settings' ? 'block p-6 md:p-8' : 'hidden'}>
                                    <div className="max-w-4xl mx-auto">
                                        <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
                                            <Settings className="w-6 h-6 text-primary" /> Post Settings
                                        </h2>

                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                                            <div className="space-y-6">
                                                <FormField label="URL Slug" icon={Hash} required>
                                                    <label className="input input-bordered flex items-center gap-2 bg-base-100">
                                                        <span className="text-base-content/50 font-mono text-sm">/blog/</span>
                                                        <input type="text" name="slug" value={slug} onChange={(e) => setSlug(e.target.value)} className="grow font-mono text-sm" required />
                                                    </label>
                                                </FormField>

                                                <FormField label="Excerpt" icon={FileText}>
                                                    <textarea
                                                        name="excerpt"
                                                        value={excerpt}
                                                        onChange={(e) => setExcerpt(e.target.value)}
                                                        className="textarea textarea-bordered h-48 text-base leading-relaxed resize-none w-full"
                                                        placeholder="Write a brief summary of your post..."
                                                    />
                                                </FormField>
                                            </div>

                                            <div className="space-y-6">
                                                <FormField label="Cover Image" icon={ImageIcon}>
                                                    <div className="w-full">
                                                        {coverImagePreview ? (
                                                            <div className="relative group rounded-xl overflow-hidden border border-base-300 shadow-sm w-full aspect-video">
                                                                <img src={coverImagePreview} alt="Cover preview" className="w-full h-full object-cover" />
                                                                <div className="absolute inset-0 bg-base-300/80 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                                                                    <button
                                                                        type="button"
                                                                        onClick={removeCoverImage}
                                                                        className="btn btn-error btn-sm gap-2"
                                                                    >
                                                                        <Trash2 className="w-4 h-4" />
                                                                        Remove
                                                                    </button>
                                                                </div>
                                                            </div>
                                                        ) : (
                                                            <div
                                                                className="border-2 border-dashed border-base-300 rounded-xl w-full aspect-video flex flex-col items-center justify-center text-center hover:border-primary/50 hover:bg-base-200/50 transition-all cursor-pointer group"
                                                                onClick={() => fileInputRef.current?.click()}
                                                            >
                                                                <div className="p-4 rounded-full bg-base-200 group-hover:bg-base-300 transition-colors mb-3">
                                                                    <Upload className="w-6 h-6 text-base-content/50" />
                                                                </div>
                                                                <p className="font-medium">Upload Cover Image</p>
                                                                <p className="text-sm text-base-content/50 mt-1">Recommended: 1200x630px</p>
                                                            </div>
                                                        )}

                                                        <input
                                                            type="file"
                                                            name="cover_image"
                                                            accept="image/*"
                                                            onChange={handleCoverImageChange}
                                                            className="hidden"
                                                            ref={fileInputRef}
                                                        />
                                                        {removeCoverImageFlag && (
                                                            <input type="hidden" name="remove_cover_image" value="true" />
                                                        )}
                                                    </div>
                                                </FormField>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                {/* SEO Tab */}
                                <div className={activeTab === 'seo' ? 'block p-6 md:p-8' : 'hidden'}>
                                    <div className="max-w-4xl mx-auto">
                                        <div className="flex items-center justify-between mb-6">
                                            <h2 className="text-2xl font-bold flex items-center gap-2">
                                                <Globe className="w-6 h-6 text-primary" /> SEO Metadata
                                            </h2>
                                            <div className="badge badge-info gap-2 p-3">
                                                <Eye className="w-4 h-4" /> Preview Mode
                                            </div>
                                        </div>

                                        <div className="alert alert-info shadow-sm mb-8">
                                            <Globe className="w-5 h-5" />
                                            <div>
                                                <h3 className="font-bold">Optimization Tips</h3>
                                                <div className="text-xs">Keep titles under 60 chars and descriptions under 160 chars for best visibility.</div>
                                            </div>
                                        </div>

                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                                            <div className="space-y-6">
                                                <FormField label="Meta Title" subLabel={`${metaTitle.length}/60 characters`}>
                                                    <input
                                                        type="text"
                                                        name="meta_title"
                                                        value={metaTitle}
                                                        onChange={(e) => setMetaTitle(e.target.value)}
                                                        placeholder="SEO title (defaults to blog title)"
                                                        className={`input input-bordered w-full ${metaTitle.length > 60 ? 'input-warning' : ''}`}
                                                        maxLength={60}
                                                    />
                                                </FormField>

                                                <FormField label="Meta Keywords">
                                                    <input
                                                        type="text"
                                                        name="meta_keywords"
                                                        value={metaKeywords}
                                                        onChange={(e) => setMetaKeywords(e.target.value)}
                                                        placeholder="keyword1, keyword2, keyword3"
                                                        className="input input-bordered w-full"
                                                    />
                                                </FormField>
                                            </div>

                                            <div className="space-y-6">
                                                <FormField label="Meta Description" subLabel={`${metaDescription.length}/160 characters`}>
                                                    <textarea
                                                        name="meta_description"
                                                        value={metaDescription}
                                                        onChange={(e) => setMetaDescription(e.target.value)}
                                                        className={`textarea textarea-bordered h-48 text-base leading-relaxed w-full resize-none ${metaDescription.length > 160 ? 'textarea-warning' : ''}`}
                                                        placeholder="SEO description (defaults to excerpt)"
                                                        maxLength={160}
                                                    />
                                                </FormField>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                            </div>
                        </div>
                    </div>
                </form>
            </div>
        </div>
    )
}

export default BlogEditor
