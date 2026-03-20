import { useState, useEffect, useCallback } from 'react'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Button, Spinner, EmptyState } from '../ui'
import PollCard from '../polls/PollCard'
import CreatePollModal from '../polls/CreatePollModal'
import type { Poll } from '../../types/poll'
import styles from './PollsTab.module.css'

interface PollsTabProps {
  policyId: string
}

export default function PollsTab({ policyId }: PollsTabProps) {
  const { isAuthenticated } = useAuth()
  const [polls, setPolls] = useState<Poll[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)

  const fetchPolls = useCallback(async () => {
    setLoading(true)
    try {
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`
      const res = await fetch(`/api/policies/${policyId}/polls`, { headers })
      if (!res.ok) throw new Error('Failed to fetch polls')
      const data = await res.json()
      setPolls(Array.isArray(data) ? data : data.polls ?? [])
    } catch {
      // silently fail
    } finally {
      setLoading(false)
    }
  }, [policyId])

  useEffect(() => {
    fetchPolls()
  }, [fetchPolls])

  function handleVoted(updated: Poll) {
    setPolls((prev) => prev.map((p) => (p.id === updated.id ? updated : p)))
  }

  function handleCreated(poll: Poll) {
    setPolls((prev) => [poll, ...prev])
  }

  function handleDelete(pollId: string) {
    setPolls((prev) => prev.filter((p) => p.id !== pollId))
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.topBar}>
        <h3 className={styles.title}>Polls ({polls.length})</h3>
        {isAuthenticated && (
          <Button size="sm" onClick={() => setShowCreate(true)}>+ Create Poll</Button>
        )}
      </div>

      {loading ? (
        <div className={styles.loading}><Spinner size="lg" /></div>
      ) : polls.length === 0 ? (
        <EmptyState
          heading="No polls yet"
          description="Be the first to create a poll for this policy."
        />
      ) : (
        <div className={styles.list}>
          {polls.map((poll) => (
            <PollCard
              key={poll.id}
              poll={poll}
              onVoted={handleVoted}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}

      <CreatePollModal
        policyId={policyId}
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={handleCreated}
      />
    </div>
  )
}
