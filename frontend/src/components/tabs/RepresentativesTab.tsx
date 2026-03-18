import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { Card } from '../ui'
import type { Representative, VotingSummary } from '../../types/representative'
import styles from './RepresentativesTab.module.css'

interface VoteWithRep {
  id: string
  representativeId: string
  policyId: string
  vote: 'yea' | 'nay' | 'abstain' | 'absent'
  date: string
  session: string
  notes?: string
  representative?: Representative
}

interface RepresentativesTabProps {
  policyId: string
}

const PER_PAGE = 50

export default function RepresentativesTab({ policyId }: RepresentativesTabProps) {
  const [votes, setVotes] = useState<VoteWithRep[]>([])
  const [summary, setSummary] = useState<VotingSummary | null>(null)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [photoErrors, setPhotoErrors] = useState<Set<string>>(new Set())

  const fetchVotes = useCallback(async () => {
    if (!policyId) return
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams({
        page: String(page),
        perPage: String(PER_PAGE),
      })
      const res = await fetch(`/api/policies/${policyId}/votes?${params}`)
      if (!res.ok) {
        throw new Error('Failed to load voting records')
      }
      const data = await res.json()
      setVotes(data.votes || [])
      setSummary(data.summary || null)
      setTotal(data.total || 0)
    } catch {
      setError('Failed to load voting records')
    } finally {
      setLoading(false)
    }
  }, [policyId, page])

  useEffect(() => {
    fetchVotes()
  }, [fetchVotes])

  const getPartyClass = (party: string) => {
    const lower = party?.toLowerCase() || ''
    if (lower.includes('democrat')) return styles.partyDemocratic
    if (lower.includes('republican')) return styles.partyRepublican
    return styles.partyOther
  }

  const getChamberLabel = (chamber: string) => {
    switch (chamber) {
      case 'senate': return 'Senate'
      case 'house': return 'House'
      case 'governor': return 'Governor'
      case 'local': return 'Local'
      default: return chamber
    }
  }

  const getVoteClass = (vote: string) => {
    switch (vote) {
      case 'yea': return styles.voteYea
      case 'nay': return styles.voteNay
      case 'abstain': return styles.voteAbstain
      case 'absent': return styles.voteAbsent
      default: return ''
    }
  }

  const handlePhotoError = (repId: string) => {
    setPhotoErrors(prev => new Set(prev).add(repId))
  }

  if (loading) {
    return (
      <Card>
        <div className={styles.loading}><div className={styles.spinner} /></div>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <p className={styles.errorText}>{error}</p>
      </Card>
    )
  }

  if (votes.length === 0) {
    return (
      <Card>
        <div className={styles.emptyState}>
          <p className={styles.emptyTitle}>No voting records yet</p>
          <p className={styles.emptyText}>
            No representatives have recorded votes on this policy.
          </p>
        </div>
      </Card>
    )
  }

  const totalPages = Math.ceil(total / PER_PAGE)

  return (
    <div className={styles.wrapper}>
      {/* Voting summary */}
      {summary && summary.totalVotes > 0 && (
        <Card>
          <div className={styles.summaryBar}>
            <span className={`${styles.summaryItem} ${styles.summaryTotal}`}>
              {summary.totalVotes} total vote{summary.totalVotes !== 1 ? 's' : ''}
            </span>
            <span className={`${styles.summaryItem} ${styles.summaryYea}`}>
              {summary.yeaCount} yea
            </span>
            <span className={`${styles.summaryItem} ${styles.summaryNay}`}>
              {summary.nayCount} nay
            </span>
            <span className={`${styles.summaryItem} ${styles.summaryAbstain}`}>
              {summary.abstainCount} abstain
            </span>
          </div>
        </Card>
      )}

      {/* Representative vote list */}
      <Card>
        <div className={styles.list}>
          {votes.map((v) => {
            const rep = v.representative
            const showPhoto = rep?.photoUrl && !photoErrors.has(v.representativeId)

            return (
              <Link
                key={v.id}
                to={`/representatives/${v.representativeId}`}
                className={styles.repRow}
              >
                {showPhoto ? (
                  <img
                    src={rep!.photoUrl}
                    alt={rep!.name}
                    className={styles.avatar}
                    onError={() => handlePhotoError(v.representativeId)}
                  />
                ) : (
                  <div className={styles.avatarFallback}>
                    {rep?.name?.charAt(0) || '?'}
                  </div>
                )}

                <div className={styles.repInfo}>
                  <p className={styles.repName}>{rep?.name || 'Unknown Representative'}</p>
                  <div className={styles.repMeta}>
                    {rep?.party && (
                      <span className={`${styles.repParty} ${getPartyClass(rep.party)}`}>
                        {rep.party}
                      </span>
                    )}
                    {rep?.chamber && (
                      <span className={styles.repChamber}>{getChamberLabel(rep.chamber)}</span>
                    )}
                    {rep?.state && (
                      <span className={styles.repChamber}>
                        {rep.state}{rep.district ? ` - ${rep.district}` : ''}
                      </span>
                    )}
                  </div>
                </div>

                <span className={`${styles.voteBadge} ${getVoteClass(v.vote)}`}>
                  {v.vote.toUpperCase()}
                </span>
              </Link>
            )
          })}
        </div>

        {totalPages > 1 && (
          <div className={styles.pagination}>
            <button
              className={styles.pageBtn}
              disabled={page <= 1}
              onClick={() => setPage(p => p - 1)}
            >
              Previous
            </button>
            <span className={styles.pageInfo}>
              Page {page} of {totalPages}
            </span>
            <button
              className={styles.pageBtn}
              disabled={page >= totalPages}
              onClick={() => setPage(p => p + 1)}
            >
              Next
            </button>
          </div>
        )}
      </Card>
    </div>
  )
}
