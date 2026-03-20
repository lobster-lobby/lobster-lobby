import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Card, Badge } from '../ui'
import DraftComments from './DraftComments'
import DraftForm from './DraftForm'
import type { Draft, DraftCategory } from '../../types/draft'
import styles from './DraftCard.module.css'

interface DraftCardProps {
  draft: Draft
  onUpdated: (draft: Draft) => void
  onDeleted: (draftId: string) => void
}

const categoryLabel: Record<DraftCategory, string> = {
  'amendment': 'Amendment',
  'talking-point': 'Talking Point',
  'position-statement': 'Position Statement',
  'full-text': 'Full Text',
}

const categoryVariant: Record<DraftCategory, 'default' | 'success' | 'neutral' | 'human'> = {
  'amendment': 'human',
  'talking-point': 'default',
  'position-statement': 'success',
  'full-text': 'neutral',
}

export default function DraftCard({ draft, onUpdated, onDeleted }: DraftCardProps) {
  const { isAuthenticated, user } = useAuth()
  const [expanded, setExpanded] = useState(false)
  const [editing, setEditing] = useState(false)
  const [endorsing, setEndorsing] = useState(false)
  const [endorsed, setEndorsed] = useState(draft.userEndorsed ?? false)
  const [endorsementCount, setEndorsementCount] = useState(draft.endorsements)
  const [endorseError, setEndorseError] = useState('')

  const isAuthor = user?.id === draft.authorId

  async function handleEndorse() {
    if (!isAuthenticated || endorsing) return
    setEndorsing(true)
    setEndorseError('')
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/drafts/${draft.id}/endorse`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
      })
      if (!res.ok) throw new Error('Failed to endorse')
      const data = await res.json()
      const newEndorsed = !endorsed
      setEndorsed(newEndorsed)
      setEndorsementCount(data.endorsements ?? endorsementCount + (newEndorsed ? 1 : -1))
    } catch (err) {
      console.error('Endorse error:', err)
      setEndorseError('Failed to update endorsement. Please try again.')
    } finally {
      setEndorsing(false)
    }
  }

  async function handleDelete() {
    if (!confirm('Archive this draft?')) return
    const token = getAccessToken()
    const res = await fetch(`/api/drafts/${draft.id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (!res.ok) {
      console.error('Failed to delete draft')
      return
    }
    onDeleted(draft.id)
  }

  const updatedAt = new Date(draft.updatedAt)
  const timeAgo = formatTimeAgo(updatedAt)

  return (
    <Card className={styles.card}>
      <div
        className={styles.cardHeader}
        onClick={() => !editing && setExpanded(!expanded)}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if ((e.key === 'Enter' || e.key === ' ') && !editing) {
            e.preventDefault()
            setExpanded(!expanded)
          }
        }}
      >
        <div className={styles.titleRow}>
          <h4 className={styles.title}>{draft.title}</h4>
          <Badge variant={categoryVariant[draft.category]}>{categoryLabel[draft.category]}</Badge>
        </div>
        <div className={styles.meta}>
          <span>{draft.authorName}</span>
          <span>v{draft.version} · {timeAgo}</span>
          {draft.status === 'draft' && <Badge variant="neutral">Draft</Badge>}
        </div>
        {!expanded && (
          <p className={styles.snippet}>{draft.content.slice(0, 200)}{draft.content.length > 200 ? '…' : ''}</p>
        )}
        <div className={styles.cardFooter} onClick={(e) => e.stopPropagation()}>
          <button
            className={[styles.endorseBtn, endorsed && styles.endorsed].filter(Boolean).join(' ')}
            onClick={handleEndorse}
            disabled={!isAuthenticated || endorsing}
            type="button"
          >
            ★ {endorsementCount} Endorse{endorsementCount !== 1 ? 's' : ''}
          </button>
          {endorseError && <span className={styles.errorText}>{endorseError}</span>}
          {isAuthor && (
            <div className={styles.authorActions}>
              <button className={styles.actionBtn} onClick={() => { setEditing(true); setExpanded(true) }} type="button">Edit</button>
              <button className={styles.actionBtn} onClick={handleDelete} type="button">Archive</button>
            </div>
          )}
          <button className={styles.expandBtn} onClick={() => setExpanded(!expanded)} type="button">
            {expanded ? '▲ Collapse' : '▼ Expand'}
          </button>
        </div>
      </div>

      {expanded && (
        <div className={styles.expanded}>
          {editing ? (
            <DraftForm
              draft={draft}
              onSaved={(updated) => { onUpdated(updated); setEditing(false) }}
              onCancel={() => setEditing(false)}
            />
          ) : (
            <>
              <div className={styles.content}>
                <ReactMarkdown>{draft.content}</ReactMarkdown>
              </div>
              <DraftComments draftId={draft.id} />
            </>
          )}
        </div>
      )}
    </Card>
  )
}

function formatTimeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  return date.toLocaleDateString()
}
