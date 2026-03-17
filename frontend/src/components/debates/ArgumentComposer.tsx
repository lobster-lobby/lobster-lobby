import { useState } from 'react'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Card } from '../ui'
import type { Argument, Side } from '../../types/debates'
import styles from './ArgumentComposer.module.css'

interface ArgumentComposerProps {
  debateSlug: string
  onArgumentCreated: (arg: Argument) => void
}

export default function ArgumentComposer({ debateSlug, onArgumentCreated }: ArgumentComposerProps) {
  const { isAuthenticated } = useAuth()
  const [side, setSide] = useState<Side | ''>('')
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isAuthenticated) {
    return (
      <Card>
        <p className={styles.loginPrompt}>Log in to add an argument.</p>
      </Card>
    )
  }

  const canSubmit = side !== '' && content.trim().length > 0 && !submitting

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!canSubmit) return

    setSubmitting(true)
    setError(null)

    try {
      const token = getAccessToken()
      const res = await fetch(`/api/debates/${debateSlug}/arguments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ content: content.trim(), side }),
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.error || 'Failed to post argument')
      }

      const data = await res.json()
      onArgumentCreated(data.argument)
      setContent('')
      setSide('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.sideSelector}>
          <label className={styles.sideLabel}>Position:</label>
          <div className={styles.sideOptions}>
            <button
              type="button"
              className={[styles.sideBtn, styles.proBtn, side === 'pro' && styles.active].filter(Boolean).join(' ')}
              onClick={() => setSide('pro')}
            >
              Pro
            </button>
            <button
              type="button"
              className={[styles.sideBtn, styles.conBtn, side === 'con' && styles.active].filter(Boolean).join(' ')}
              onClick={() => setSide('con')}
            >
              Con
            </button>
          </div>
        </div>

        <textarea
          className={styles.textarea}
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Make your argument... (Markdown supported)"
          rows={4}
          maxLength={10000}
        />

        {error && <p className={styles.error}>{error}</p>}

        <div className={styles.footer}>
          <span className={styles.charCount}>{content.length}/10000</span>
          <button type="submit" className={styles.submitBtn} disabled={!canSubmit}>
            {submitting ? 'Posting...' : 'Post Argument'}
          </button>
        </div>
      </form>
    </Card>
  )
}
