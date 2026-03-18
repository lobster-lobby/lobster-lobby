import { useState, useEffect, useCallback } from 'react'
import { Button, Spinner, EmptyState, Toast } from '../../ui'
import type { CampaignComment } from '../../../types/campaign'
import { useAuth, getAccessToken } from '../../../hooks/useAuth'
import { relativeTime } from '../../../utils/time'
import styles from './DiscussionTab.module.css'

interface DiscussionTabProps {
  campaignId: string
  campaignCreatedBy: string
}

type SortOption = 'newest' | 'votes'

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'newest', label: 'Newest' },
  { value: 'votes', label: 'Most Voted' },
]

interface CommentItemProps {
  comment: CampaignComment
  userVote: number
  onVote: (commentId: string, value: number) => void
  onReply: (parentId: string) => void
  onTogglePin: (commentId: string) => void
  replies: CampaignComment[]
  userVotes: Record<string, number>
  isAuthenticated: boolean
  canPin: boolean
  depth?: number
}

function CommentItem({
  comment,
  userVote,
  onVote,
  onReply,
  onTogglePin,
  replies,
  userVotes,
  isAuthenticated,
  canPin,
  depth = 0,
}: CommentItemProps) {
  const maxDepth = 1

  return (
    <div
      className={`${styles.comment}${comment.pinned ? ` ${styles.pinned}` : ''}`}
      style={{ marginLeft: depth > 0 ? 'var(--ll-space-lg)' : 0 }}
    >
      <div className={styles.commentHeader}>
        <span className={styles.avatar}>{(comment.authorName || '?').slice(0, 2).toUpperCase()}</span>
        <span className={styles.author}>{comment.authorName}</span>
        <span className={styles.time}>{relativeTime(comment.createdAt)}</span>
        {comment.pinned && <span className={styles.pinBadge}>Pinned</span>}
      </div>
      <div className={styles.commentBody}>{comment.body}</div>
      <div className={styles.commentActions}>
        <div className={styles.voteButtons}>
          <button
            className={[styles.voteBtn, userVote === 1 && styles.active].filter(Boolean).join(' ')}
            onClick={() => onVote(comment.id, 1)}
            disabled={!isAuthenticated}
            title={isAuthenticated ? 'Upvote' : 'Log in to vote'}
          >
            +
          </button>
          <span className={styles.voteCount}>{comment.votes}</span>
          <button
            className={[styles.voteBtn, userVote === -1 && styles.active].filter(Boolean).join(' ')}
            onClick={() => onVote(comment.id, -1)}
            disabled={!isAuthenticated}
            title={isAuthenticated ? 'Downvote' : 'Log in to vote'}
          >
            -
          </button>
        </div>
        {depth < maxDepth && isAuthenticated && (
          <button className={styles.replyBtn} onClick={() => onReply(comment.id)}>
            Reply
          </button>
        )}
        {canPin && depth === 0 && (
          <button
            className={styles.pinBtn}
            onClick={() => onTogglePin(comment.id)}
            title={comment.pinned ? 'Unpin comment' : 'Pin comment'}
          >
            {comment.pinned ? 'Unpin' : 'Pin'}
          </button>
        )}
      </div>
      {replies.length > 0 && (
        <div className={styles.replies}>
          {replies.map((reply) => (
            <CommentItem
              key={reply.id}
              comment={reply}
              userVote={userVotes[reply.id] || 0}
              onVote={onVote}
              onReply={onReply}
              onTogglePin={onTogglePin}
              replies={[]}
              userVotes={userVotes}
              isAuthenticated={isAuthenticated}
              canPin={false}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export default function DiscussionTab({ campaignId, campaignCreatedBy }: DiscussionTabProps) {
  const { user, isAuthenticated } = useAuth()
  const [comments, setComments] = useState<CampaignComment[]>([])
  const [userVotes, setUserVotes] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sort, setSort] = useState<SortOption>('newest')
  const [newComment, setNewComment] = useState('')
  const [replyTo, setReplyTo] = useState<string | null>(null)
  const [replyText, setReplyText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [toast, setToast] = useState<{ message: string; variant: 'success' | 'error' | 'info' } | null>(null)

  const canPin = isAuthenticated && user != null && (user.id === campaignCreatedBy || user.role === 'admin')

  const fetchComments = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const params = new URLSearchParams({ sort })
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }

      const res = await fetch(`/api/campaigns/${campaignId}/comments?${params}`, { headers })
      if (!res.ok) throw new Error('Failed to fetch comments')

      const data = await res.json()
      setComments(data.comments || [])
      setUserVotes(data.userVotes || {})
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }, [campaignId, sort])

  useEffect(() => {
    fetchComments()
  }, [fetchComments])

  const handleVote = async (commentId: string, value: number) => {
    if (!isAuthenticated) {
      setToast({ message: 'Please log in to vote', variant: 'info' })
      return
    }

    const token = getAccessToken()
    try {
      const res = await fetch(`/api/campaigns/${campaignId}/comments/${commentId}/vote`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ value }),
      })

      if (!res.ok) throw new Error('Failed to vote')

      const data = await res.json()
      setComments((prev) =>
        prev.map((c) => (c.id === commentId ? data.comment : c))
      )
      setUserVotes((prev) => ({ ...prev, [commentId]: data.userVote }))
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : 'Failed to vote',
        variant: 'error',
      })
    }
  }

  const handleTogglePin = async (commentId: string) => {
    const token = getAccessToken()
    try {
      const res = await fetch(`/api/campaigns/${campaignId}/comments/${commentId}/pin`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.error || 'Failed to toggle pin')
      }

      const data = await res.json()
      setComments((prev) =>
        prev.map((c) => (c.id === commentId ? data.comment : c))
      )
      setToast({
        message: data.comment.pinned ? 'Comment pinned' : 'Comment unpinned',
        variant: 'success',
      })
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : 'Failed to toggle pin',
        variant: 'error',
      })
    }
  }

  const handleSubmitComment = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newComment.trim()) return

    setSubmitting(true)
    const token = getAccessToken()

    try {
      const res = await fetch(`/api/campaigns/${campaignId}/comments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ body: newComment.trim() }),
      })

      if (!res.ok) throw new Error('Failed to post comment')

      setNewComment('')
      setToast({ message: 'Comment posted!', variant: 'success' })
      fetchComments()
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : 'Failed to post comment',
        variant: 'error',
      })
    } finally {
      setSubmitting(false)
    }
  }

  const handleSubmitReply = async (parentId: string) => {
    if (!replyText.trim()) return

    setSubmitting(true)
    const token = getAccessToken()

    try {
      const res = await fetch(`/api/campaigns/${campaignId}/comments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ body: replyText.trim(), parentId }),
      })

      if (!res.ok) throw new Error('Failed to post reply')

      setReplyTo(null)
      setReplyText('')
      setToast({ message: 'Reply posted!', variant: 'success' })
      fetchComments()
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : 'Failed to post reply',
        variant: 'error',
      })
    } finally {
      setSubmitting(false)
    }
  }

  // Nest comments by parentId, pinned float to top
  const topLevelComments = comments.filter((c) => !c.parentId)
  const pinnedComments = topLevelComments.filter((c) => c.pinned)
  const unpinnedComments = topLevelComments.filter((c) => !c.pinned)
  const sortedTopLevel = [...pinnedComments, ...unpinnedComments]

  const repliesByParent = comments.reduce<Record<string, CampaignComment[]>>((acc, c) => {
    if (c.parentId) {
      if (!acc[c.parentId]) acc[c.parentId] = []
      acc[c.parentId].push(c)
    }
    return acc
  }, {})

  if (loading && comments.length === 0) {
    return (
      <div className={styles.loading}>
        <Spinner size="lg" />
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2 className={styles.title}>Discussion ({comments.length})</h2>
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value as SortOption)}
          className={styles.select}
        >
          {SORT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {isAuthenticated ? (
        <form onSubmit={handleSubmitComment} className={styles.commentForm}>
          <textarea
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            placeholder="Share your thoughts..."
            className={styles.textarea}
            rows={3}
            maxLength={2000}
          />
          <div className={styles.formActions}>
            <span className={styles.charCount}>{newComment.length}/2000</span>
            <Button type="submit" disabled={submitting || !newComment.trim()}>
              {submitting ? 'Posting...' : 'Post Comment'}
            </Button>
          </div>
        </form>
      ) : (
        <div className={styles.loginPrompt}>
          <p>Log in to join the discussion</p>
        </div>
      )}

      {error && <div className={styles.error}>{error}</div>}

      {topLevelComments.length === 0 && !loading ? (
        <EmptyState
          heading="No comments yet"
          description={
            isAuthenticated
              ? 'Be the first to start the discussion!'
              : 'No one has commented on this campaign yet.'
          }
        />
      ) : (
        <div className={styles.commentList}>
          {sortedTopLevel.map((comment) => (
            <div key={comment.id}>
              <CommentItem
                comment={comment}
                userVote={userVotes[comment.id] || 0}
                onVote={handleVote}
                onReply={(parentId) => {
                  setReplyTo(parentId)
                  setReplyText('')
                }}
                onTogglePin={handleTogglePin}
                replies={repliesByParent[comment.id] || []}
                userVotes={userVotes}
                isAuthenticated={isAuthenticated}
                canPin={canPin}
              />
              {replyTo === comment.id && (
                <div className={styles.replyForm}>
                  <textarea
                    value={replyText}
                    onChange={(e) => setReplyText(e.target.value)}
                    placeholder="Write a reply..."
                    className={styles.textarea}
                    rows={2}
                    maxLength={2000}
                    autoFocus
                  />
                  <div className={styles.formActions}>
                    <Button variant="ghost" size="sm" onClick={() => setReplyTo(null)}>
                      Cancel
                    </Button>
                    <Button
                      size="sm"
                      onClick={() => handleSubmitReply(comment.id)}
                      disabled={submitting || !replyText.trim()}
                    >
                      {submitting ? 'Posting...' : 'Reply'}
                    </Button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {toast && (
        <div className={styles.toastContainer}>
          <Toast variant={toast.variant} onClose={() => setToast(null)} autoDismiss={3000}>
            {toast.message}
          </Toast>
        </div>
      )}
    </div>
  )
}
