import { useState, useEffect, useCallback } from 'react'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Card } from '../ui'
import type { SummaryPoint, SummaryPosition, SummaryResponse } from '../../types/summary'
import styles from './CommunitySummary.module.css'

interface CommunitySummaryProps {
  policyId: string
  userPosition?: SummaryPosition
}

export default function CommunitySummary({ policyId, userPosition }: CommunitySummaryProps) {
  const { isAuthenticated } = useAuth()
  const [support, setSupport] = useState<SummaryPoint[]>([])
  const [oppose, setOppose] = useState<SummaryPoint[]>([])
  const [consensus, setConsensus] = useState<SummaryPoint[]>([])
  const [showAll, setShowAll] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [formContent, setFormContent] = useState('')
  const [formPosition, setFormPosition] = useState<SummaryPosition>('support')
  const [submitting, setSubmitting] = useState(false)

  const totalPoints = support.length + oppose.length + consensus.length

  const fetchSummary = useCallback(async () => {
    try {
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`

      const params = new URLSearchParams()
      if (showAll) params.set('all', 'true')

      const res = await fetch(`/api/policies/${policyId}/debate/summary?${params.toString()}`, { headers })
      if (!res.ok) return
      const data: SummaryResponse = await res.json()
      setSupport(data.support)
      setOppose(data.oppose)
      setConsensus(data.consensus)
    } catch {
      // Silently fail
    }
  }, [policyId, showAll])

  useEffect(() => {
    fetchSummary()
  }, [fetchSummary])

  async function handleEndorse(pointId: string) {
    if (!isAuthenticated) return
    // Send the USER's position (their stance on the policy), not the point's position.
    // This is what enables cross-position bridging: a support user endorsing an oppose point.
    const endorserPosition: SummaryPosition = userPosition ?? 'consensus'
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/policies/${policyId}/debate/summary/${pointId}/endorse`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ position: endorserPosition }),
      })
      if (res.ok) fetchSummary()
    } catch {
      // Silently fail
    }
  }

  async function handleRemoveEndorsement(pointId: string) {
    if (!isAuthenticated) return
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/policies/${policyId}/debate/summary/${pointId}/endorse`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` },
      })
      if (res.ok) fetchSummary()
    } catch {
      // Silently fail
    }
  }

  async function handleSubmit() {
    if (!formContent.trim() || submitting) return
    setSubmitting(true)
    try {
      const token = getAccessToken()
      const res = await fetch(`/api/policies/${policyId}/debate/summary`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ content: formContent.trim(), position: formPosition }),
      })
      if (res.ok) {
        setFormContent('')
        setShowForm(false)
        fetchSummary()
      }
    } catch {
      // Silently fail
    } finally {
      setSubmitting(false)
    }
  }

  function renderPoint(point: SummaryPoint) {
    return (
      <div key={point.id} className={`${styles.point} ${!point.visible ? styles.pointHidden : ''}`}>
        <p className={styles.pointContent}>{point.content}</p>
        <div className={styles.pointMeta}>
          <span className={styles.author}>{point.authorUsername}</span>
          <span className={styles.endorseCount}>{point.endorseCount} endorsements</span>
          {point.crossCount > 0 && (
            <span className={styles.crossBadge}>{point.crossCount} cross-position</span>
          )}
          {isAuthenticated && (
            point.userEndorsed ? (
              <button
                className={`${styles.endorseBtn} ${styles.endorseBtnActive}`}
                onClick={() => handleRemoveEndorsement(point.id)}
                type="button"
              >
                Endorsed
              </button>
            ) : (
              <button
                className={styles.endorseBtn}
                onClick={() => handleEndorse(point.id)}
                type="button"
              >
                Endorse
              </button>
            )
          )}
        </div>
      </div>
    )
  }

  // Only return null when we're showing everything and there are genuinely no points.
  // When showAll=false and totalPoints=0, there may be hidden points — show the toggle.
  if (totalPoints === 0 && showAll) return null

  return (
    <Card>
      <div className={styles.wrapper}>
        <div className={styles.headerRow}>
          <h4 className={styles.title}>Community Summary</h4>
          <div className={styles.actions}>
            <button
              className={styles.toggleBtn}
              onClick={() => setShowAll(!showAll)}
              type="button"
            >
              {showAll ? 'Show visible only' : 'Show all points'}
            </button>
            {isAuthenticated && (
              <button
                className={styles.nominateBtn}
                onClick={() => setShowForm(!showForm)}
                type="button"
              >
                {showForm ? 'Cancel' : 'Nominate a point'}
              </button>
            )}
          </div>
        </div>

        {showForm && (
          <div className={styles.form}>
            <textarea
              className={styles.textarea}
              value={formContent}
              onChange={(e) => setFormContent(e.target.value)}
              placeholder="Summarize a key point from the debate (10-500 chars)..."
              maxLength={500}
            />
            <div className={styles.formRow}>
              <select
                className={styles.select}
                value={formPosition}
                onChange={(e) => setFormPosition(e.target.value as SummaryPosition)}
              >
                <option value="support">Support</option>
                <option value="oppose">Oppose</option>
                <option value="consensus">Consensus</option>
              </select>
              <button
                className={styles.submitBtn}
                onClick={handleSubmit}
                disabled={submitting || formContent.trim().length < 10}
                type="button"
              >
                {submitting ? 'Submitting...' : 'Submit Point'}
              </button>
            </div>
          </div>
        )}

        <div className={styles.sections}>
          <div className={styles.section}>
            <h5 className={styles.sectionTitle} style={{ color: 'var(--ll-support)' }}>
              Key Support ({support.length})
            </h5>
            {support.length === 0 ? (
              <p className={styles.emptySection}>No support points yet</p>
            ) : support.map(renderPoint)}
          </div>

          <div className={styles.section}>
            <h5 className={styles.sectionTitle} style={{ color: 'var(--ll-oppose)' }}>
              Key Opposition ({oppose.length})
            </h5>
            {oppose.length === 0 ? (
              <p className={styles.emptySection}>No opposition points yet</p>
            ) : oppose.map(renderPoint)}
          </div>

          {consensus.length > 0 && (
            <div className={`${styles.section} ${styles.consensusSection}`}>
              <h5 className={styles.sectionTitle} style={{ color: 'var(--ll-neutral)' }}>
                Consensus Points ({consensus.length})
              </h5>
              {consensus.map(renderPoint)}
            </div>
          )}
        </div>
      </div>
    </Card>
  )
}
