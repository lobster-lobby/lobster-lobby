import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import type { Representative, VotingSummary, VotingRecord } from '../types/representative'
import { Pagination } from '../components/ui'
import styles from './RepresentativeDetail.module.css'

export default function RepresentativeDetail() {
  const { id } = useParams<{ id: string }>()
  const [rep, setRep] = useState<Representative | null>(null)
  const [summary, setSummary] = useState<VotingSummary | null>(null)
  const [votes, setVotes] = useState<VotingRecord[]>([])
  const [votesTotal, setVotesTotal] = useState(0)
  const [votesPage, setVotesPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [photoError, setPhotoError] = useState(false)

  const perPage = 20

  useEffect(() => {
    fetchProfile()
  }, [id])

  useEffect(() => {
    if (id) fetchVotes()
  }, [id, votesPage])

  const fetchProfile = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch(`/api/representatives/${id}`)
      if (!res.ok) {
        setError(res.status === 404 ? 'Representative not found' : 'Failed to load profile')
        return
      }
      const data = await res.json()
      setRep(data.representative)
      setSummary(data.votingSummary)
    } catch {
      setError('Failed to load profile')
    } finally {
      setLoading(false)
    }
  }

  const fetchVotes = async () => {
    try {
      const params = new URLSearchParams({
        page: String(votesPage),
        perPage: String(perPage),
      })
      const res = await fetch(`/api/representatives/${id}/votes?${params}`)
      if (!res.ok) return
      const data = await res.json()
      setVotes(data.votes || [])
      setVotesTotal(data.total || 0)
    } catch {
      // fail silently
    }
  }

  const getPartyClass = (p: string) => {
    const lower = p?.toLowerCase() || ''
    if (lower.includes('democrat')) return styles.partyDemocratic
    if (lower.includes('republican')) return styles.partyRepublican
    return styles.partyOther
  }

  const getChamberLabel = (c: string) => {
    switch (c) {
      case 'senate': return 'Senate'
      case 'house': return 'House'
      case 'governor': return 'Governor'
      case 'local': return 'Local'
      default: return c
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

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  if (loading) {
    return <div className={styles.loading}><div className={styles.spinner} /></div>
  }

  if (error || !rep) {
    return (
      <div className={styles.errorState}>
        <h2>{error || 'Representative not found'}</h2>
        <Link to="/representatives" className={styles.backLink}>Back to Representatives</Link>
      </div>
    )
  }

  const totalVotesPages = Math.ceil(votesTotal / perPage)

  return (
    <div className={styles.page}>
      <Link to="/representatives" className={styles.backLink}>← Back to Representatives</Link>

      {/* Profile Header */}
      <header className={styles.profileHeader}>
        <div className={styles.profilePhoto}>
          {rep.photoUrl && !photoError ? (
            <img
              src={rep.photoUrl}
              alt={rep.name}
              className={styles.photo}
              onError={() => setPhotoError(true)}
            />
          ) : (
            <div className={styles.avatar}>{rep.name.charAt(0)}</div>
          )}
        </div>
        <div className={styles.profileInfo}>
          <h1 className={styles.profileName}>{rep.name}</h1>
          <p className={styles.profileTitle}>{rep.title || getChamberLabel(rep.chamber)}</p>
          <div className={styles.profileMeta}>
            <span className={`${styles.profileParty} ${getPartyClass(rep.party)}`}>
              {rep.party}
            </span>
            <span className={styles.profileState}>
              {rep.state}{rep.district ? ` - District ${rep.district}` : ''}
            </span>
            <span className={styles.profileChamber}>{getChamberLabel(rep.chamber)}</span>
          </div>
        </div>
      </header>

      {/* Bio */}
      {rep.bio && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>About</h2>
          <p className={styles.bio}>{rep.bio}</p>
        </section>
      )}

      {/* Contact Info */}
      {(rep.contactInfo?.phone || rep.contactInfo?.email || rep.contactInfo?.website || rep.contactInfo?.office) && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Contact</h2>
          <div className={styles.contactGrid}>
            {rep.contactInfo.phone && (
              <div className={styles.contactItem}>
                <span className={styles.contactLabel}>Phone</span>
                <a href={`tel:${rep.contactInfo.phone}`}>{rep.contactInfo.phone}</a>
              </div>
            )}
            {rep.contactInfo.email && (
              <div className={styles.contactItem}>
                <span className={styles.contactLabel}>Email</span>
                <a href={`mailto:${rep.contactInfo.email}`}>{rep.contactInfo.email}</a>
              </div>
            )}
            {rep.contactInfo.website && (
              <div className={styles.contactItem}>
                <span className={styles.contactLabel}>Website</span>
                <a href={rep.contactInfo.website} target="_blank" rel="noopener noreferrer">
                  {rep.contactInfo.website}
                </a>
              </div>
            )}
            {rep.contactInfo.office && (
              <div className={styles.contactItem}>
                <span className={styles.contactLabel}>Office</span>
                <span>{rep.contactInfo.office}</span>
              </div>
            )}
          </div>
        </section>
      )}

      {/* Voting Summary */}
      {summary && summary.totalVotes > 0 && (
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Voting Summary</h2>
          <div className={styles.statsGrid}>
            <div className={styles.statCard}>
              <div className={styles.statValue}>{summary.totalVotes}</div>
              <div className={styles.statLabel}>Total Votes</div>
            </div>
            <div className={`${styles.statCard} ${styles.statYea}`}>
              <div className={styles.statValue}>{summary.yeaPercent.toFixed(1)}%</div>
              <div className={styles.statLabel}>Yea ({summary.yeaCount})</div>
            </div>
            <div className={`${styles.statCard} ${styles.statNay}`}>
              <div className={styles.statValue}>{summary.nayPercent.toFixed(1)}%</div>
              <div className={styles.statLabel}>Nay ({summary.nayCount})</div>
            </div>
            <div className={`${styles.statCard} ${styles.statAbstain}`}>
              <div className={styles.statValue}>{summary.abstainPercent.toFixed(1)}%</div>
              <div className={styles.statLabel}>Abstain ({summary.abstainCount})</div>
            </div>
          </div>
        </section>
      )}

      {/* Voting Records Table */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Voting Record</h2>
        {votes.length === 0 ? (
          <p className={styles.emptyText}>No voting records available.</p>
        ) : (
          <>
            <div className={styles.tableWrapper}>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Session</th>
                    <th>Vote</th>
                    <th>Notes</th>
                  </tr>
                </thead>
                <tbody>
                  {votes.map((vote) => (
                    <tr key={vote.id}>
                      <td>{formatDate(vote.date)}</td>
                      <td>{vote.session}</td>
                      <td>
                        <span className={`${styles.voteBadge} ${getVoteClass(vote.vote)}`}>
                          {vote.vote.toUpperCase()}
                        </span>
                      </td>
                      <td>{vote.notes || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {totalVotesPages > 1 && (
              <Pagination
                currentPage={votesPage}
                totalPages={totalVotesPages}
                onPageChange={setVotesPage}
              />
            )}
          </>
        )}
      </section>
    </div>
  )
}
