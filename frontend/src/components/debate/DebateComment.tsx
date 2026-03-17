import { useState, useCallback } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { UserBadge, VoteButtons } from '../ui'
import CommentComposer from './CommentComposer'
import type { Comment } from '../../types/debate'
import styles from './DebateComment.module.css'

interface DebateCommentProps {
  comment: Comment
  policyId: string
  depth?: number
}

function relativeTime(dateStr: string): string {
  const seconds = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  return `${months}mo ago`
}

export default function DebateComment({ comment, policyId, depth = 0 }: DebateCommentProps) {
  const { isAuthenticated } = useAuth()
  const [showReplyForm, setShowReplyForm] = useState(false)
  const [replies, setReplies] = useState<Comment[]>([])
  const [loadedReplies, setLoadedReplies] = useState(false)
  const [loadingReplies, setLoadingReplies] = useState(false)
  const [currentReaction, setCurrentReaction] = useState(comment.userReaction)
  const [upvotes, setUpvotes] = useState(comment.upvotes)
  const [downvotes, setDownvotes] = useState(comment.downvotes)

  const loadReplies = useCallback(async () => {
    if (loadingReplies) return
    setLoadingReplies(true)
    try {
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`

      const res = await fetch(`/api/policies/${policyId}/debate/${comment.id}/replies`, { headers })
      if (!res.ok) throw new Error('Failed to load replies')
      const data = await res.json()
      setReplies(data.replies)
      setLoadedReplies(true)
    } catch {
      // Silently fail
    } finally {
      setLoadingReplies(false)
    }
  }, [policyId, comment.id, loadingReplies])

  async function handleReact(value: number) {
    if (!isAuthenticated) return
    const newValue = currentReaction === value ? 0 : value
    const token = getAccessToken()

    try {
      const res = await fetch(`/api/policies/${policyId}/debate/${comment.id}/react`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ value: newValue }),
      })
      if (!res.ok) return

      // Optimistic update
      if (currentReaction === 1) setUpvotes((v) => v - 1)
      else if (currentReaction === -1) setDownvotes((v) => v - 1)

      if (newValue === 1) setUpvotes((v) => v + 1)
      else if (newValue === -1) setDownvotes((v) => v + 1)

      setCurrentReaction(newValue)
    } catch {
      // Silently fail
    }
  }

  function handleReplyCreated(newComment: Comment) {
    setReplies((prev) => [...prev, newComment])
    setShowReplyForm(false)
    setLoadedReplies(true)
  }

  const positionClass = styles[comment.position] || ''
  const userVote = currentReaction === 1 ? 'up' as const : currentReaction === -1 ? 'down' as const : null

  return (
    <div className={[styles.wrapper, positionClass].join(' ')} style={{ marginLeft: depth > 0 ? 'var(--ll-space-md)' : undefined }}>
      <div className={styles.header}>
        <UserBadge username={comment.authorUsername} type={comment.authorType} />
        <span className={styles.meta}>
          {comment.authorRepTier && <span className={styles.tier}>{comment.authorRepTier}</span>}
          <span className={styles.time}>{relativeTime(comment.createdAt)}</span>
          {comment.editedAt && <span className={styles.edited}>(edited)</span>}
        </span>
      </div>

      <div className={styles.content}>
        <ReactMarkdown remarkPlugins={[remarkGfm]}>{comment.content}</ReactMarkdown>
      </div>

      <div className={styles.actions}>
        <VoteButtons
          upvotes={upvotes}
          downvotes={downvotes}
          userVote={userVote}
          onUpvote={() => handleReact(1)}
          onDownvote={() => handleReact(-1)}
        />

        {depth < 3 && (
          <button
            className={styles.replyBtn}
            onClick={() => setShowReplyForm(!showReplyForm)}
            type="button"
          >
            Reply{comment.replyCount > 0 ? ` (${comment.replyCount})` : ''}
          </button>
        )}

        {comment.endorsed && <span className={styles.endorsed}>Endorsed</span>}
      </div>

      {showReplyForm && (
        <CommentComposer
          policyId={policyId}
          parentId={comment.id}
          onCommentCreated={handleReplyCreated}
        />
      )}

      {comment.replyCount > 0 && !loadedReplies && (
        <button className={styles.loadReplies} onClick={loadReplies} type="button">
          {loadingReplies ? 'Loading...' : `Show ${comment.replyCount} ${comment.replyCount === 1 ? 'reply' : 'replies'}`}
        </button>
      )}

      {loadedReplies && replies.map((reply) => (
        <DebateComment
          key={reply.id}
          comment={reply}
          policyId={policyId}
          depth={depth + 1}
        />
      ))}
    </div>
  )
}
