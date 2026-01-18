import React, { useState, useCallback } from 'react'
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
    Highlighter, Minus
} from 'lucide-react'

// Lowlight setup for syntax highlighting
const lowlight = createLowlight(common)

const MenuBar = ({ editor }) => {
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

    const addImage = useCallback(() => {
        const url = window.prompt('URL')
        if (url) {
            editor.chain().focus().setImage({ src: url }).run()
        }
    }, [editor])

    const addYoutube = useCallback(() => {
        const url = window.prompt('Enter YouTube URL')
        if (url) {
            editor.commands.setYoutubeVideo({ src: url })
        }
    }, [editor])

    return (
        <div className="bg-base-200/50 border-b border-base-300 p-2 flex flex-wrap gap-1">
            {/* Formatting */}
            <button
                type="button"
                onClick={() => editor.chain().focus().toggleBold().run()}
                disabled={!editor.can().chain().focus().toggleBold().run()}
                className={`btn btn-ghost btn-sm btn-square ${editor.isActive('bold') ? 'bg-base-300 text-primary' : ''}`}
                title="Bold"
            >
                <Bold className="w-4 h-4" />
            </button>
            <button
                type="button"
                onClick={() => editor.chain().focus().toggleItalic().run()}
                disabled={!editor.can().chain().focus().toggleItalic().run()}
                className={`btn btn-ghost btn-sm btn-square ${editor.isActive('italic') ? 'bg-base-300 text-primary' : ''}`}
                title="Italic"
            >
                <Italic className="w-4 h-4" />
            </button>
            <button
                type="button"
                onClick={() => editor.chain().focus().toggleUnderline().run()}
                className={`btn btn-ghost btn-sm btn-square ${editor.isActive('underline') ? 'bg-base-300 text-primary' : ''}`}
                title="Underline"
            >
                <UnderlineIcon className="w-4 h-4" />
            </button>
            <button
                type="button"
                onClick={() => editor.chain().focus().toggleStrike().run()}
                disabled={!editor.can().chain().focus().toggleStrike().run()}
                className={`btn btn-ghost btn-sm btn-square ${editor.isActive('strike') ? 'bg-base-300 text-primary' : ''}`}
                title="Strike"
            >
                <Strikethrough className="w-4 h-4" />
            </button>
            <button
                type="button"
                onClick={() => editor.chain().focus().toggleHighlight().run()}
                className={`btn btn-ghost btn-sm btn-square ${editor.isActive('highlight') ? 'bg-base-300 text-primary' : ''}`}
                title="Highlight"
            >
                <Highlighter className="w-4 h-4" />
            </button>
            <button
                type="button"
                onClick={() => editor.chain().focus().toggleSuperscript().run()}
                className={`btn btn-ghost btn-sm btn-square ${editor.isActive('superscript') ? 'bg-base-300 text-primary' : ''}`}
                title="Superscript"
            >
                <SuperscriptIcon className="w-4 h-4" />
            </button>
            <button
                type="button"
                onClick={() => editor.chain().focus().toggleSubscript().run()}
                className={`btn btn-ghost btn-sm btn-square ${editor.isActive('subscript') ? 'bg-base-300 text-primary' : ''}`}
                title="Subscript"
            >
                <SubscriptIcon className="w-4 h-4" />
            </button>


            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Links & Media */}
            <button type="button" onClick={setLink} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('link') ? 'bg-base-300 text-primary' : ''}`} title="Link">
                <LinkIcon className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().unsetLink().run()} disabled={!editor.isActive('link')} className="btn btn-ghost btn-sm btn-square" title="Unlink">
                <Unlink className="w-4 h-4" />
            </button>
            <button type="button" onClick={addImage} className="btn btn-ghost btn-sm btn-square" title="Image">
                <ImageIcon className="w-4 h-4" />
            </button>
            <button type="button" onClick={addYoutube} className="btn btn-ghost btn-sm btn-square" title="Youtube">
                <YoutubeIcon className="w-4 h-4" />
            </button>

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Table */}
            <button
                type="button"
                onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()}
                className="btn btn-ghost btn-sm btn-square"
                title="Insert Table"
            >
                <TableIcon className="w-4 h-4" />
            </button>
            {editor.isActive('table') && (
                <>
                    <button type="button" onClick={() => editor.chain().focus().addColumnAfter().run()} className="btn btn-ghost btn-sm btn-square" title="Add Column">
                        <Columns className="w-4 h-4" />
                    </button>
                    <button type="button" onClick={() => editor.chain().focus().addRowAfter().run()} className="btn btn-ghost btn-sm btn-square" title="Add Row">
                        <Rows className="w-4 h-4" />
                    </button>
                    <button type="button" onClick={() => editor.chain().focus().deleteTable().run()} className="btn btn-ghost btn-sm btn-square text-error" title="Delete Table">
                        <Trash2 className="w-4 h-4" />
                    </button>
                </>
            )}

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Alignment */}
            <button type="button" onClick={() => editor.chain().focus().setTextAlign('left').run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive({ textAlign: 'left' }) ? 'bg-base-300 text-primary' : ''}`} title="Left">
                <AlignLeft className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().setTextAlign('center').run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive({ textAlign: 'center' }) ? 'bg-base-300 text-primary' : ''}`} title="Center">
                <AlignCenter className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().setTextAlign('right').run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive({ textAlign: 'right' }) ? 'bg-base-300 text-primary' : ''}`} title="Right">
                <AlignRight className="w-4 h-4" />
            </button>

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Headings */}
            <button type="button" onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('heading', { level: 1 }) ? 'bg-base-300 text-primary' : ''}`} title="H1">
                <Heading1 className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('heading', { level: 2 }) ? 'bg-base-300 text-primary' : ''}`} title="H2">
                <Heading2 className="w-4 h-4" />
            </button>

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            {/* Lists & Blocks */}
            <button type="button" onClick={() => editor.chain().focus().toggleBulletList().run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('bulletList') ? 'bg-base-300 text-primary' : ''}`} title="Bullet List">
                <List className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().toggleOrderedList().run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('orderedList') ? 'bg-base-300 text-primary' : ''}`} title="Ordered List">
                <ListOrdered className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().toggleTaskList().run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('taskList') ? 'bg-base-300 text-primary' : ''}`} title="Task List">
                <CheckSquare className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().toggleBlockquote().run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('blockquote') ? 'bg-base-300 text-primary' : ''}`} title="Quote">
                <Quote className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().toggleCodeBlock().run()} className={`btn btn-ghost btn-sm btn-square ${editor.isActive('codeBlock') ? 'bg-base-300 text-primary' : ''}`} title="Code">
                <Code className="w-4 h-4" />
            </button>

            {editor.isActive('codeBlock') && (
                <select
                    className="select select-bordered select-xs max-w-xs"
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
            <button type="button" onClick={() => editor.chain().focus().setHorizontalRule().run()} className="btn btn-ghost btn-sm btn-square" title="Horizontal Rule">
                <Minus className="w-4 h-4" />
            </button>

            <div className="divider divider-horizontal mx-0 w-px h-6 my-auto bg-base-300"></div>

            <button type="button" onClick={() => editor.chain().focus().undo().run()} disabled={!editor.can().chain().focus().undo().run()} className="btn btn-ghost btn-sm btn-square" title="Undo">
                <Undo className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => editor.chain().focus().redo().run()} disabled={!editor.can().chain().focus().redo().run()} className="btn btn-ghost btn-sm btn-square" title="Redo">
                <Redo className="w-4 h-4" />
            </button>
        </div>
    )
}

const BlogEditor = ({ initialData, actionUrl }) => {
    const [title, setTitle] = useState(initialData?.title || '')
    const [slug, setSlug] = useState(initialData?.slug || '')
    const [excerpt, setExcerpt] = useState(initialData?.excerpt || '')
    const [content, setContent] = useState(initialData?.content || '')
    const [isPublished, setIsPublished] = useState(initialData?.is_published || false)

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
                class: 'prose prose-sm sm:prose lg:prose-lg xl:prose-xl mx-auto focus:outline-none min-h-[300px] text-base-content p-4',
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

    return (
        <div className="px-4 sm:px-6 lg:px-8 py-8">
            <div className="flex items-center justify-between mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-base-content">{initialData ? 'Edit Blog' : 'Create New Blog'}</h1>
                </div>
                <a href="/a/blogs" className="btn btn-ghost gap-2">
                    <ArrowLeft className="w-4 h-4" />
                    Back to List
                </a>
            </div>

            <div className="bg-base-100 rounded-lg shadow-sm border border-base-200 p-6">
                <form action={actionUrl} method="POST" className="space-y-6">
                    <div className="form-control">
                        <label className="label"><span className="label-text font-medium">Title</span></label>
                        <input type="text" name="title" value={title} onChange={handleTitleChange} placeholder="Enter blog title" className="input input-bordered w-full" required />
                    </div>

                    <div className="form-control">
                        <label className="label"><span className="label-text font-medium">Slug</span></label>
                        <input type="text" name="slug" value={slug} onChange={(e) => setSlug(e.target.value)} className="input input-bordered w-full font-mono text-sm" required />
                    </div>

                    <div className="form-control">
                        <label className="label"><span className="label-text font-medium">Excerpt</span></label>
                        <textarea name="excerpt" value={excerpt} onChange={(e) => setExcerpt(e.target.value)} className="textarea textarea-bordered h-20" placeholder="Brief summary for listings..." />
                    </div>

                    <div className="form-control">
                        <label className="label"><span className="label-text font-medium">Content</span></label>
                        <div className="border border-base-300 rounded-lg bg-base-100 overflow-hidden focus-within:ring-2 focus-within:ring-primary/20">
                            <MenuBar editor={editor} />
                            <EditorContent editor={editor} className="min-h-[300px]" />
                        </div>
                        <input type="hidden" name="content" value={content} />
                    </div>

                    <div className="form-control">
                        <label className="label cursor-pointer justify-start gap-4">
                            <span className="label-text font-medium">Publish immediately</span>
                            <input type="checkbox" name="is_published" checked={isPublished} onChange={(e) => setIsPublished(e.target.checked)} className="toggle toggle-primary" value="on" />
                        </label>
                    </div>

                    <div className="flex justify-end gap-3 pt-4 border-t border-base-200">
                        <a href="/a/blogs" className="btn btn-ghost">Cancel</a>
                        <button type="submit" className="btn btn-primary">
                            {initialData ? 'Save Changes' : 'Create Blog'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}

export default BlogEditor
