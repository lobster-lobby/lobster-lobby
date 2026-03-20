import { useState } from 'react'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Card, Badge, Button } from '../ui'
import type { Poll } from '../../types/poll'
import styles from './PollCard.module.css'

interface PollCardProps {
  poll: Poll
  onVoted: (updatedPoll: Poll) => void
  onDelete?: (pollId: string) => void
}

export default function PollCard({ poll, onVoted, onDelete }: PollCardProps) {
  const { isAuthenticated, user } = useAuth()
  const [selected, setSelected] = useState<string[]>(poll.userVoteOptionIds || [])
  const [voting, setVoting] = useState(false)
  const [hasVoted, setHasVoted] = useState((poll.userVoteOptionIds?.length ?? 0) > 0)
  const [voteError, setVoteError] = useState('')

  const isAuthor = user?.id === poll.authorId
  const isClosed = poll.status === 'closed'
  const showResults = hasVoted || isClosed

  function toggleOption(optionId: string) {
    if (!poll.multiSelect) {
      setSelected([optionId])
      return
    }
    setSelected((prev) =>
      prev.includes(optionId) ? prev.filter((id) => id !== optionId) : [...prev, optionId]
    )
  }

  async function handleVote() {
    if (!isAuthenticated || selected.length === 0 || voting) return
    setVoting(true)
    setVoteError('')
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/polls/${poll.id}/vote`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({ optionIds: selected }),
      })
      if (!res.ok) throw new Error('Vote failed')
      const updated: Poll = await res.json()
      updated.userVoteOptionIds = selected
      setHasVoted(true)
      onVoted(updated)
    } catch (err) {
      console.error('Vote error:', err)
      setVoteError('Failed to submit vote. Please try again.')
    } finally {
      setVoting(false)
    }
  }

  async function handleDelete() {
    if (!confirm('Close this poll?')) return
    const token = getAccessToken()
    const res = await fetch(`/api/polls/${poll.id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (!res.ok) {
      console.error('Failed to delete poll')
      return
    }
    onDelete?.(poll.id)
  }

  const endsAt = poll.endsAt ? new Date(poll.endsAt) : null
  const timeLabel = endsAt
    ? isClosed
      ? `Ended ${endsAt.toLocaleDateString()}`
      : `Ends ${endsAt.toLocaleDateString()}`
    : null

  return (
    <Card className={styles.card}>
      <div className={styles.header}>
        <div className={styles.meta}>
          <span className={styles.author}>{poll.authorName}</span>
          {timeLabel && <span className={styles.time}>{timeLabel}</span>}
        </div>
        <div className={styles.badges}>
          {isClosed && <Badge variant="neutral">Closed</Badge>}
          {poll.multiSelect && <Badge variant="default">Multi-select</Badge>}
          {isAuthor && !isClosed && (
            <button className={styles.closeBtn} onClick={handleDelete} type="button">
              Close Poll
            </button>
          )}
        </div>
      </div>

      <h3 className={styles.question}>{poll.question}</h3>

      <div className={styles.options}>
        {poll.options.map((opt) => {
          const pct = poll.totalVotes > 0 ? Math.round((opt.votes / poll.totalVotes) * 100) : 0
          const isUserChoice = selected.includes(opt.id)

          if (showResults) {
            return (
              <div key={opt.id} className={[styles.resultBar, isUserChoice && styles.userChoice].filter(Boolean).join(' ')}>
                <div className={styles.barFill} style={{ width: `${pct}%` }} />
                <span className={styles.barLabel}>{opt.text}</span>
                <span className={styles.barPct}>{pct}%</span>
              </div>
            )
          }

          return (
            <label key={opt.id} className={[styles.optionLabel, isUserChoice && styles.optionSelected].filter(Boolean).join(' ')}>
              <input
                type={poll.multiSelect ? 'checkbox' : 'radio'}
                name={`poll-${poll.id}`}
                checked={isUserChoice}
                onChange={() => toggleOption(opt.id)}
                disabled={!isAuthenticated || isClosed}
              />
              <span>{opt.text}</span>
            </label>
          )
        })}
      </div>

      {voteError && <p className={styles.voteError}>{voteError}</p>}

      <div className={styles.footer}>
        <span className={styles.totalVotes}>{poll.totalVotes} vote{poll.totalVotes !== 1 ? 's' : ''}</span>
        {!showResults && isAuthenticated && !isClosed && (
          <Button
            size="sm"
            onClick={handleVote}
            disabled={selected.length === 0 || voting}
          >
            {voting ? 'Voting…' : 'Vote'}
          </Button>
        )}
      </div>
    </Card>
  )
}
