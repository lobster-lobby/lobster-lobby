import { useState, useRef, useEffect } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { UserBadge, VoteButtons } from '../ui'
import type { Argument, FlagReason } from '../../types/debates'
import { relativeTime } from '../../utils/time'
import styles from './DebateArgument.module.css'

const FLAG_REASONS: { value: FlagReason; label: string }[] = [
  { value: 'spam', label: 'Spam' },
  { value: 'harassment', label: 'Harassment' },
  { value: 'misinformation', label: 'Misinformation' },
  { value: 'off-topic', label: 'Off-topic' },
]

interface DebateArgumentProps {
  argument: Argument
  debateSlug: string
}

export default function DebateArgument({ argument, debateSlug }: DebateArgumentProps) {
  const { isAuthenticated } = useAuth()
  const [currentVote, setCurrentVote] = useState(argument.userVote)
  const [upvotes, setUpvotes] = useState(argument.upvotes)
  const [downvotes, setDownvotes] = useState(argument.downvotes)
  const [showFlagMenu, setShowFlagMenu] = useState(false)
  const [flagStatus, setFlagStatus] = useState<'idle' | 'flagged' | 'error'>(
    argument.flagged ? 'flagged' : 'idle'
  )
  const flagRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (flagRef.current && !flagRef.current.contains(e.target as Node)) {
        setShowFlagMenu(false)
      }
    }
    if (showFlagMenu) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showFlagMenu])

  async function handleVote(value: number) {
    if (!isAuthenticated) return
    const token = getAccessToken()

    // Compute optimistic new state (toggle logic)
    const newValue = currentVote === value ? 0 : value

    // Optimistic update
    if (currentVote === 1) setUpvotes((v) => v - 1)
    else if (currentVote === -1) setDownvotes((v) => v - 1)

    if (newValue === 1) setUpvotes((v) => v + 1)
    else if (newValue === -1) setDownvotes((v) => v + 1)

    const prevVote = currentVote
    const prevUp = upvotes
    const prevDown = downvotes
    setCurrentVote(newValue)

    try {
      const res = await fetch(`/api/debates/${debateSlug}/arguments/${argument.id}/vote`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ value }),
      })
      if (!res.ok) {
        // Revert on failure
        setCurrentVote(prevVote)
        setUpvotes(prevUp)
        setDownvotes(prevDown)
      }
    } catch {
      // Revert on failure
      setCurrentVote(prevVote)
      setUpvotes(prevUp)
      setDownvotes(prevDown)
    }
  }

  async function handleFlag(reason: FlagReason) {
    setShowFlagMenu(false)
    const token = getAccessToken()

    try {
      const res = await fetch(`/api/debates/${debateSlug}/arguments/${argument.id}/flag`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ reason }),
      })
      if (res.ok) {
        setFlagStatus('flagged')
      } else if (res.status === 409) {
        setFlagStatus('flagged')
      } else {
        setFlagStatus('error')
      }
    } catch {
      setFlagStatus('error')
    }
  }

  const sideClass = argument.side === 'pro' ? styles.pro : styles.con
  const flaggedClass = flagStatus === 'flagged' || argument.flagged ? styles.flagged : ''
  const userVote = currentVote === 1 ? 'up' as const : currentVote === -1 ? 'down' as const : null
  const netScore = upvotes - downvotes

  return (
    <div className={[styles.wrapper, sideClass, flaggedClass].filter(Boolean).join(' ')}>
      <div className={styles.header}>
        <UserBadge username={argument.authorUsername} type={argument.authorType} />
        <span className={styles.meta}>
          {argument.authorRepTier && <span className={styles.tier}>{argument.authorRepTier}</span>}
          <span className={styles.time}>{relativeTime(argument.createdAt)}</span>
        </span>
        {argument.flagged && <span className={styles.flagBadge}>Flagged</span>}
        <span className={styles.sideLabel}>{argument.side === 'pro' ? 'Pro' : 'Con'}</span>
      </div>

      <div className={styles.content}>
        <ReactMarkdown remarkPlugins={[remarkGfm]}>{argument.content}</ReactMarkdown>
      </div>

      <div className={styles.actions}>
        <VoteButtons
          upvotes={upvotes}
          downvotes={downvotes}
          userVote={userVote}
          onUpvote={() => handleVote(1)}
          onDownvote={() => handleVote(-1)}
        />
        <span className={styles.netScore}>{netScore > 0 ? '+' : ''}{netScore}</span>

        {isAuthenticated && (
          <div className={styles.flagWrapper} ref={flagRef}>
            <button
              className={[styles.flagBtn, flagStatus === 'flagged' ? styles.flagBtnActive : ''].filter(Boolean).join(' ')}
              onClick={() => setShowFlagMenu(!showFlagMenu)}
              title="Flag this argument"
            >
              {'\u{1F6A9}'}
            </button>
            {showFlagMenu && (
              <div className={styles.flagMenu}>
                {FLAG_REASONS.map((r) => (
                  <button
                    key={r.value}
                    className={styles.flagMenuItem}
                    onClick={() => handleFlag(r.value)}
                  >
                    {r.label}
                  </button>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
