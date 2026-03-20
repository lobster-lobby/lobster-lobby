import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { getAccessToken } from '../../hooks/useAuth'
import { Button, Input, Textarea } from '../ui'
import type { Draft, DraftCategory, DraftStatus, CreateDraftPayload } from '../../types/draft'
import styles from './DraftForm.module.css'

interface DraftFormProps {
  policyId?: string
  draft?: Draft
  onSaved: (draft: Draft) => void
  onCancel: () => void
}

const CATEGORIES: { value: DraftCategory; label: string }[] = [
  { value: 'amendment', label: 'Amendment' },
  { value: 'talking-point', label: 'Talking Point' },
  { value: 'position-statement', label: 'Position Statement' },
  { value: 'full-text', label: 'Full Text' },
]

export default function DraftForm({ policyId, draft, onSaved, onCancel }: DraftFormProps) {
  const [title, setTitle] = useState(draft?.title ?? '')
  const [content, setContent] = useState(draft?.content ?? '')
  const [category, setCategory] = useState<DraftCategory>(draft?.category ?? 'amendment')
  const [status, setStatus] = useState<DraftStatus>(draft?.status ?? 'draft')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [preview, setPreview] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!title.trim() || !content.trim()) {
      setError('Title and content are required.')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      const token = getAccessToken()
      const payload: CreateDraftPayload = { title: title.trim(), content: content.trim(), category, status }

      const url = draft ? `/api/drafts/${draft.id}` : `/api/policies/${policyId}/drafts`
      const method = draft ? 'PUT' : 'POST'

      const res = await fetch(url, {
        method,
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw new Error('Failed to save draft')
      const saved: Draft = await res.json()
      onSaved(saved)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save draft')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <div className={styles.field}>
        <label className={styles.label}>Title</label>
        <Input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Draft title…"
          maxLength={200}
          required
        />
        <span className={styles.charCount}>{title.length}/200</span>
      </div>

      <div className={styles.field}>
        <label className={styles.label}>Category</label>
        <select
          value={category}
          onChange={(e) => setCategory(e.target.value as DraftCategory)}
          className={styles.select}
        >
          {CATEGORIES.map((c) => (
            <option key={c.value} value={c.value}>{c.label}</option>
          ))}
        </select>
      </div>

      <div className={styles.field}>
        <div className={styles.contentHeader}>
          <label className={styles.label}>Content (Markdown)</label>
          <button type="button" className={styles.previewToggle} onClick={() => setPreview(!preview)}>
            {preview ? 'Edit' : 'Preview'}
          </button>
        </div>
        {preview ? (
          <div className={styles.preview}>
            <ReactMarkdown>{content}</ReactMarkdown>
          </div>
        ) : (
          <Textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Write your draft content in Markdown…"
            rows={10}
            required
          />
        )}
      </div>

      <div className={styles.statusRow}>
        <label className={styles.label}>Status:</label>
        <label className={styles.radioLabel}>
          <input type="radio" name="status" value="draft" checked={status === 'draft'} onChange={() => setStatus('draft')} />
          Save as Draft
        </label>
        <label className={styles.radioLabel}>
          <input type="radio" name="status" value="published" checked={status === 'published'} onChange={() => setStatus('published')} />
          Publish
        </label>
      </div>

      {error && <p className={styles.error}>{error}</p>}

      <div className={styles.actions}>
        <Button type="button" variant="secondary" onClick={onCancel}>Cancel</Button>
        <Button type="submit" disabled={submitting}>
          {submitting ? 'Saving…' : draft ? 'Update Draft' : 'Create Draft'}
        </Button>
      </div>
    </form>
  )
}
