import { useState, useEffect, useCallback } from 'react'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Spinner, Button } from '../ui'
import type { DraftComment } from '../../types/draft'
import styles from './DraftComments.module.css'

interface DraftCommentsProps {
  draftId: string
}

export default function DraftComments({ draftId }: DraftCommentsProps) {
  const { isAuthenticated } = useAuth()
  const [comments, setComments] = useState<DraftComment[]>([])
  const [loading, setLoading] = useState(true)
  const [text, setText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const fetchComments = useCallback(async () => {
    setLoading(true)
    try {
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`
      const res = await fetch(`/api/drafts/${draftId}/comments`, { headers })
      if (!res.ok) throw new Error('Failed to fetch comments')
      const data = await res.json()
      setComments(Array.isArray(data) ? data : data.comments ?? [])
    } catch (err) {
      console.error('Failed to fetch comments:', err)
      setError('Failed to load comments.')
    } finally {
      setLoading(false)
    }
  }, [draftId])

  useEffect(() => {
    fetchComments()
  }, [fetchComments])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!text.trim() || submitting) return
    setSubmitting(true)
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/drafts/${draftId}/comments`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: text.trim() }),
      })
      if (!res.ok) throw new Error('Failed to post comment')
      const comment: DraftComment = await res.json()
      setComments((prev) => [...prev, comment])
      setText('')
      setError('')
    } catch (err) {
      console.error('Failed to post comment:', err)
      setError('Failed to post comment. Please try again.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className={styles.wrapper}>
      <h4 className={styles.heading}>Comments</h4>

      {error && <p className={styles.error}>{error}</p>}

      {loading ? (
        <div className={styles.loading}><Spinner size="sm" /></div>
      ) : comments.length === 0 ? (
        <p className={styles.empty}>No comments yet.</p>
      ) : (
        <div className={styles.list}>
          {comments.map((c) => (
            <div key={c.id} className={styles.comment}>
              <div className={styles.commentMeta}>
                <strong>{c.authorName}</strong>
                <span className={styles.commentTime}>{new Date(c.createdAt).toLocaleDateString()}</span>
              </div>
              <p className={styles.commentText}>{c.content}</p>
            </div>
          ))}
        </div>
      )}

      {isAuthenticated && (
        <form onSubmit={handleSubmit} className={styles.form}>
          <textarea
            className={styles.textarea}
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Add a comment…"
            rows={2}
          />
          <Button type="submit" size="sm" disabled={!text.trim() || submitting}>
            {submitting ? 'Posting…' : 'Post'}
          </Button>
        </form>
      )}
    </div>
  )
}
