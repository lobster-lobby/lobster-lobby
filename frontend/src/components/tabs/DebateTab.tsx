import { useState, useEffect, useCallback } from 'react'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Card, TabNav, Spinner, EmptyState, Badge } from '../ui'
import PositionBar from '../debate/PositionBar'
import CommunitySummary from '../debate/CommunitySummary'
import CommentComposer from '../debate/CommentComposer'
import DebateComment from '../debate/DebateComment'
import type { Comment, Position, DebateResponse } from '../../types/debate'
import styles from './DebateTab.module.css'

interface DebateTabProps {
  policyId: string
}

const sortTabs = [
  { id: 'best', label: 'Best' },
  { id: 'newest', label: 'Newest' },
  { id: 'top', label: 'Top' },
  { id: 'discussed', label: 'Most Discussed' },
]

const positionFilters: { id: string; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'support', label: 'Support' },
  { id: 'oppose', label: 'Oppose' },
  { id: 'neutral', label: 'Neutral' },
]

export default function DebateTab({ policyId }: DebateTabProps) {
  const { isAuthenticated } = useAuth()
  const [comments, setComments] = useState<Comment[]>([])
  const [positions, setPositions] = useState({ support: 0, oppose: 0, neutral: 0 })
  const [total, setTotal] = useState(0)
  const [sort, setSort] = useState('best')
  const [positionFilter, setPositionFilter] = useState('all')
  const [loading, setLoading] = useState(true)
  const [stance, setStance] = useState<string | null>(null)

  const fetchComments = useCallback(async () => {
    setLoading(true)
    try {
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`

      const params = new URLSearchParams()
      params.set('sort', sort)
      params.set('position', positionFilter)
      params.set('page', '1')
      params.set('perPage', '50')

      const res = await fetch(`/api/policies/${policyId}/debate?${params.toString()}`, { headers })
      if (!res.ok) throw new Error('Failed to fetch debate')
      const data: DebateResponse = await res.json()
      setComments(data.comments)
      setPositions(data.positions)
      setTotal(data.total)
    } catch {
      // Silently fail
    } finally {
      setLoading(false)
    }
  }, [policyId, sort, positionFilter])

  const fetchStance = useCallback(async () => {
    if (!isAuthenticated) return
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/policies/${policyId}/stance`, {
        headers: { 'Authorization': `Bearer ${token}` },
      })
      if (!res.ok) return
      const data = await res.json()
      if (data.stance) setStance(data.stance.position)
    } catch {
      // Silently fail
    }
  }, [policyId, isAuthenticated])

  useEffect(() => {
    fetchComments()
  }, [fetchComments])

  useEffect(() => {
    fetchStance()
  }, [fetchStance])

  function handleCommentCreated(comment: Comment) {
    setComments((prev) => [comment, ...prev])
    setTotal((t) => t + 1)
    const pos = comment.position as Position
    setPositions((prev) => ({ ...prev, [pos]: prev[pos] + 1 }))
    setStance(comment.position)
  }

  const supportComments = comments.filter((c) => c.position === 'support')
  const opposeComments = comments.filter((c) => c.position === 'oppose')
  const neutralComments = comments.filter((c) => c.position === 'neutral')

  return (
    <div className={styles.wrapper}>
      <Card>
        <div className={styles.header}>
          <div className={styles.titleRow}>
            <h3 className={styles.title}>Debate ({total})</h3>
            {stance && (
              <Badge variant={stance as 'support' | 'oppose' | 'neutral'}>
                Your stance: {stance}
              </Badge>
            )}
          </div>
          <PositionBar support={positions.support} oppose={positions.oppose} neutral={positions.neutral} />
        </div>
      </Card>

      <CommunitySummary policyId={policyId} />

      <CommentComposer policyId={policyId} onCommentCreated={handleCommentCreated} />

      <div className={styles.controls}>
        <TabNav tabs={sortTabs} activeTab={sort} onTabChange={setSort} />
        <div className={styles.filters}>
          {positionFilters.map((f) => (
            <button
              key={f.id}
              className={[styles.filterBtn, positionFilter === f.id && styles.activeFilter].filter(Boolean).join(' ')}
              onClick={() => setPositionFilter(f.id)}
              type="button"
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className={styles.loading}><Spinner size="lg" /></div>
      ) : comments.length === 0 ? (
        <EmptyState heading="No comments yet" description="Be the first to share your perspective on this policy." />
      ) : (
        <>
          {/* Desktop: side-by-side when showing all */}
          {positionFilter === 'all' ? (
            <div className={styles.columns}>
              <div className={styles.column}>
                <h4 className={styles.columnTitle} style={{ color: 'var(--ll-support)' }}>Support</h4>
                {supportComments.length === 0 ? (
                  <p className={styles.emptyColumn}>No support comments</p>
                ) : supportComments.map((c) => (
                  <DebateComment key={c.id} comment={c} policyId={policyId} />
                ))}
              </div>
              <div className={styles.column}>
                <h4 className={styles.columnTitle} style={{ color: 'var(--ll-oppose)' }}>Oppose</h4>
                {opposeComments.length === 0 ? (
                  <p className={styles.emptyColumn}>No opposition comments</p>
                ) : opposeComments.map((c) => (
                  <DebateComment key={c.id} comment={c} policyId={policyId} />
                ))}
              </div>
              {neutralComments.length > 0 && (
                <div className={styles.neutralSection}>
                  <h4 className={styles.columnTitle} style={{ color: 'var(--ll-neutral)' }}>Neutral</h4>
                  {neutralComments.map((c) => (
                    <DebateComment key={c.id} comment={c} policyId={policyId} />
                  ))}
                </div>
              )}
            </div>
          ) : (
            <div className={styles.feed}>
              {comments.map((c) => (
                <DebateComment key={c.id} comment={c} policyId={policyId} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
