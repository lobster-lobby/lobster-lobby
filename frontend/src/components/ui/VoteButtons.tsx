import type { HTMLAttributes } from 'react'
import styles from './VoteButtons.module.css'

export interface VoteButtonsProps extends HTMLAttributes<HTMLDivElement> {
  upvotes: number
  downvotes: number
  userVote?: 'up' | 'down' | null
  onUpvote?: () => void
  onDownvote?: () => void
}

export function VoteButtons({
  upvotes,
  downvotes,
  userVote,
  onUpvote,
  onDownvote,
  className = '',
  ...props
}: VoteButtonsProps) {
  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} {...props}>
      <button
        className={[styles.button, styles.upvote, userVote === 'up' && styles.active]
          .filter(Boolean)
          .join(' ')}
        onClick={onUpvote}
        aria-label="Upvote"
        aria-pressed={userVote === 'up'}
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="m18 15-6-6-6 6" />
        </svg>
        <span className={styles.count}>{upvotes}</span>
      </button>
      <button
        className={[styles.button, styles.downvote, userVote === 'down' && styles.active]
          .filter(Boolean)
          .join(' ')}
        onClick={onDownvote}
        aria-label="Downvote"
        aria-pressed={userVote === 'down'}
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
        <span className={styles.count}>{downvotes}</span>
      </button>
    </div>
  )
}
