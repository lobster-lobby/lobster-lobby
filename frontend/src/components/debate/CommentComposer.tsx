import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAuth } from '../../hooks/useAuth'
import { Button } from '../ui'
import type { Comment, Position } from '../../types/debate'
import styles from './CommentComposer.module.css'

interface CommentComposerProps {
  policyId: string
  onCommentCreated: (comment: Comment) => void
  parentId?: string
}

const positions: { value: Position; label: string }[] = [
  { value: 'support', label: 'Support' },
  { value: 'oppose', label: 'Oppose' },
  { value: 'neutral', label: 'Neutral' },
]

export default function CommentComposer({ policyId, onCommentCreated, parentId }: CommentComposerProps) {
  const { isAuthenticated } = useAuth()
  const [position, setPosition] = useState<Position | null>(null)
  const [content, setContent] = useState('')
  const [preview, setPreview] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isAuthenticated) {
    return (
      <div className={styles.wrapper}>
        <p className={styles.loginPrompt}>Login to join the debate</p>
      </div>
    )
  }

  const canSubmit = position !== null && content.trim().length > 0 && !submitting

  async function handleSubmit() {
    if (!canSubmit) return
    setSubmitting(true)
    setError(null)

    try {
      const token = localStorage.getItem('ll_token')
      const body: Record<string, string> = { content: content.trim(), position: position! }
      if (parentId) body.parentId = parentId

      const res = await fetch(`/api/policies/${policyId}/debate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.error || 'Failed to post comment')
      }

      const data = await res.json()
      onCommentCreated(data.comment)
      setContent('')
      setPosition(null)
      setPreview(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to post comment')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.positionSelector}>
        {positions.map((p) => (
          <button
            key={p.value}
            className={[
              styles.positionBtn,
              styles[p.value],
              position === p.value && styles.active,
            ].filter(Boolean).join(' ')}
            onClick={() => setPosition(p.value)}
            type="button"
          >
            {p.label}
          </button>
        ))}
      </div>

      <div className={styles.editorArea}>
        <div className={styles.editorTabs}>
          <button
            className={[styles.tab, !preview && styles.activeTab].filter(Boolean).join(' ')}
            onClick={() => setPreview(false)}
            type="button"
          >
            Write
          </button>
          <button
            className={[styles.tab, preview && styles.activeTab].filter(Boolean).join(' ')}
            onClick={() => setPreview(true)}
            type="button"
          >
            Preview
          </button>
        </div>

        {preview ? (
          <div className={styles.preview}>
            {content
              ? <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
              : 'Nothing to preview'}
          </div>
        ) : (
          <textarea
            className={styles.textarea}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Share your perspective..."
            rows={4}
          />
        )}
      </div>

      {error && <p className={styles.error}>{error}</p>}

      <div className={styles.actions}>
        <Button
          variant="primary"
          size="sm"
          disabled={!canSubmit}
          onClick={handleSubmit}
        >
          {submitting ? 'Posting...' : 'Post Comment'}
        </Button>
      </div>
    </div>
  )
}
