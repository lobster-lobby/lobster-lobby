import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { Card, Button } from '../ui'
import { CampaignCard, type CampaignCardData } from '../campaigns/CampaignCard'
import { useAuth } from '../../hooks/useAuth'
import styles from './CampaignsTab.module.css'

interface CampaignsTabProps {
  policyId: string
  policyTitle?: string
  totalVotes?: number
}

const PER_PAGE = 12

export default function CampaignsTab({ policyId, policyTitle, totalVotes = 0 }: CampaignsTabProps) {
  const { isAuthenticated } = useAuth()
  const [campaigns, setCampaigns] = useState<CampaignCardData[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchCampaigns = useCallback(async () => {
    if (!policyId) return
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams({
        policyId,
        page: String(page),
        perPage: String(PER_PAGE),
        sort: 'trending',
      })
      const res = await fetch(`/api/campaigns?${params}`)
      if (!res.ok) throw new Error('Failed to load campaigns')
      const data = await res.json()
      setCampaigns(data.campaigns || [])
      setTotal(data.total || 0)
    } catch {
      setError('Failed to load campaigns')
    } finally {
      setLoading(false)
    }
  }, [policyId, page])

  useEffect(() => {
    fetchCampaigns()
  }, [fetchCampaigns])

  const totalPages = Math.ceil(total / PER_PAGE)
  const showReadyBadge = totalVotes > 10

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
        <Button variant="secondary" size="sm" onClick={fetchCampaigns}>Try again</Button>
      </Card>
    )
  }

  return (
    <div className={styles.wrapper}>
      {/* Header with badge and action */}
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <h3 className={styles.title}>
            Campaigns {total > 0 && <span className={styles.count}>({total})</span>}
          </h3>
          {showReadyBadge && (
            <span className={styles.readyBadge}>Ready for Campaign</span>
          )}
        </div>
        {isAuthenticated && (
          <Link
            to={`/campaigns/new?policyId=${policyId}`}
            className={styles.startButton}
          >
            Start a Campaign
          </Link>
        )}
      </div>

      {policyTitle && showReadyBadge && (
        <p className={styles.readyMessage}>
          This policy has strong community support with {totalVotes} votes — a great candidate for a campaign.
        </p>
      )}

      {campaigns.length === 0 ? (
        <Card>
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>📢</div>
            <p className={styles.emptyTitle}>No campaigns yet</p>
            <p className={styles.emptyText}>
              Be the first to start a campaign for this policy.
            </p>
            {isAuthenticated && (
              <Link
                to={`/campaigns/new?policyId=${policyId}`}
                className={styles.startButton}
              >
                Start a Campaign
              </Link>
            )}
          </div>
        </Card>
      ) : (
        <>
          <div className={styles.grid}>
            {campaigns.map((campaign) => (
              <CampaignCard key={campaign.id} campaign={campaign} />
            ))}
          </div>

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <button
                className={styles.pageBtn}
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </button>
              <span className={styles.pageInfo}>
                Page {page} of {totalPages}
              </span>
              <button
                className={styles.pageBtn}
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
