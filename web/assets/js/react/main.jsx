import React from 'react'
import { createRoot } from 'react-dom/client'
import BlogEditor from './components/BlogEditor'

const container = document.getElementById('react-root')
if (container) {
    const root = createRoot(container)

    // Parse initial data injected by Templ
    // Parse initial data injected by Templ
    const dataElement = document.getElementById('editor-data')
    const dataProps = dataElement ? JSON.parse(dataElement.textContent) : {}
    const actionUrl = container.dataset.action || '/a/blogs/create'

    root.render(<BlogEditor initialData={dataProps} actionUrl={actionUrl} />)
}
