import { useState, useEffect, useCallback } from 'react'
import { Card, Spinner, EmptyState } from '../components/ui'
import { getAccessToken } from '../hooks/useAuth'
import type { FlaggedArgument } from '../types/debates'
import { relativeTime } from '../utils/time'
import styles from './AdminModeration.module.css'

type ModerationAction = 'approve' | 'remove' | 'ban'

export default function AdminModeration() {
  const [queue, setQueue] = useState<FlaggedArgument[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionInProgress, setActionInProgress] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const fetchQueue = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const token = getAccessToken()
      const res = await fetch('/api/admin/moderation/queue', {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.status === 403) {
        setError('Admin access required')
        return
      }
      if (!res.ok) {
        setError('Failed to load moderation queue')
        return
      }
      const data = await res.json()
      setQueue(data.queue)
    } catch {
      setError('Something went wrong')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchQueue()
  }, [fetchQueue])

  async function handleAction(argId: string, action: ModerationAction) {
    setActionInProgress(argId)
    setActionError(null)
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/admin/moderation/${argId}/action`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ action }),
      })
      if (res.ok) {
        setQueue((prev) => prev.filter((a) => a.id !== argId))
      } else {
        const data = await res.json().catch(() => ({}))
        setActionError(data.error || `Action failed (${res.status})`)
      }
    } catch {
      setActionError('Network error — please try again')
    } finally {
      setActionInProgress(null)
    }
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.loading}><Spinner size="lg" /></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.container}>
        <Card>
          <div className={styles.errorState}>
            <h2>Error</h2>
            <p>{error}</p>
          </div>
        </Card>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Moderation Queue</h1>

      {actionError && (
        <div className={styles.actionError} role="alert">
          {actionError}
          <button onClick={() => setActionError(null)} aria-label="Dismiss">&times;</button>
        </div>
      )}

      {queue.length === 0 ? (
        <EmptyState heading="Queue is empty" description="No flagged arguments to review." />
      ) : (
        <div className={styles.list}>
          {queue.map((item) => (
            <Card key={item.id}>
              <div className={styles.item}>
                <div className={styles.itemHeader}>
                  <span className={styles.debateLink}>{item.debateTitle}</span>
                  <span className={styles.meta}>
                    by {item.authorUsername} &middot; {relativeTime(item.createdAt)}
                  </span>
                  <span className={styles.flagCount}>{item.flagCount} flags</span>
                </div>
                <div className={styles.itemContent}>{item.content}</div>
                <div className={styles.itemActions}>
                  <button
                    className={styles.approveBtn}
                    disabled={actionInProgress === item.id}
                    onClick={() => handleAction(item.id, 'approve')}
                  >
                    Approve
                  </button>
                  <button
                    className={styles.removeBtn}
                    disabled={actionInProgress === item.id}
                    onClick={() => handleAction(item.id, 'remove')}
                  >
                    Remove
                  </button>
                  <button
                    className={styles.banBtn}
                    disabled={actionInProgress === item.id}
                    onClick={() => handleAction(item.id, 'ban')}
                  >
                    Ban User
                  </button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
