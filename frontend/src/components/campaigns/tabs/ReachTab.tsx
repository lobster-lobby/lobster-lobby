import { useState, useEffect, useCallback } from 'react'
import { Spinner, Button } from '../../ui'
import type {
  ReachMetrics,
  CampaignActivity,
  CampaignActivityType,
} from '../../../types/campaign'
import { ACTIVITY_TYPE_CONFIG } from '../../../types/campaign'
import { relativeTime } from '../../../utils/time'
import styles from './ReachTab.module.css'

interface ReachTabProps {
  campaignId: string
}

interface MetricCardProps {
  label: string
  value: number | string
}

function MetricCard({ label, value }: MetricCardProps) {
  return (
    <div className={styles.metricCard}>
      <span className={styles.metricValue}>{value}</span>
      <span className={styles.metricLabel}>{label}</span>
    </div>
  )
}

interface TrendingIndicatorProps {
  score: number
}

function TrendingIndicator({ score }: TrendingIndicatorProps) {
  let direction: 'up' | 'flat' | 'down'
  let className: string

  if (score > 1) {
    direction = 'up'
    className = styles.trendUp
  } else if (score > 0) {
    direction = 'flat'
    className = styles.trendFlat
  } else {
    direction = 'down'
    className = styles.trendDown
  }

  const arrow = direction === 'up' ? '\u2191' : direction === 'down' ? '\u2193' : '\u2192'

  return (
    <div className={`${styles.trendingCard} ${className}`}>
      <span className={styles.trendingArrow}>{arrow}</span>
      <div className={styles.trendingInfo}>
        <span className={styles.trendingScore}>{score.toFixed(1)}</span>
        <span className={styles.trendingLabel}>Trending Score</span>
      </div>
    </div>
  )
}

function getActivityConfig(type_: CampaignActivityType) {
  return ACTIVITY_TYPE_CONFIG[type_] || ACTIVITY_TYPE_CONFIG.join
}

export default function ReachTab({ campaignId }: ReachTabProps) {
  const [metrics, setMetrics] = useState<ReachMetrics | null>(null)
  const [activities, setActivities] = useState<CampaignActivity[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchMetrics = useCallback(async () => {
    try {
      const res = await fetch(`/api/campaigns/${campaignId}/reach`)
      if (!res.ok) throw new Error('Failed to fetch reach metrics')
      const data: ReachMetrics = await res.json()
      setMetrics(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    }
  }, [campaignId])

  const fetchActivities = useCallback(
    async (pageNum: number, append: boolean) => {
      try {
        const res = await fetch(
          `/api/campaigns/${campaignId}/activity?page=${pageNum}&limit=20`
        )
        if (!res.ok) throw new Error('Failed to fetch activities')
        const data = await res.json()
        setActivities((prev) => (append ? [...prev, ...data.activities] : data.activities))
        setTotal(data.total)
        setPage(pageNum)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Something went wrong')
      }
    },
    [campaignId]
  )

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      setError(null)
      await Promise.all([fetchMetrics(), fetchActivities(1, false)])
      setLoading(false)
    }
    load()
  }, [fetchMetrics, fetchActivities])

  const handleLoadMore = async () => {
    setLoadingMore(true)
    await fetchActivities(page + 1, true)
    setLoadingMore(false)
  }

  if (loading) {
    return (
      <div className={styles.loading}>
        <Spinner size="lg" />
      </div>
    )
  }

  if (error) {
    return <div className={styles.error}>{error}</div>
  }

  const hasMore = activities.length < total

  return (
    <div className={styles.container}>
      <h2 className={styles.title}>Campaign Reach</h2>

      {metrics && (
        <section className={styles.section}>
          <h3 className={styles.sectionTitle}>Overview</h3>
          <div className={styles.metricsGrid}>
            <MetricCard label="Supporters" value={metrics.totalSupporters} />
            <MetricCard label="Total Shares" value={metrics.totalShares} />
            <MetricCard label="Downloads" value={metrics.totalDownloads} />
            <MetricCard label="Participants" value={metrics.uniqueParticipants} />
            <TrendingIndicator score={metrics.trendingScore} />
          </div>
        </section>
      )}

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Activity Feed</h3>
        {activities.length === 0 ? (
          <div className={styles.emptyFeed}>No activity yet</div>
        ) : (
          <div className={styles.activityFeed}>
            {activities.map((activity) => {
              const config = getActivityConfig(activity.type)
              return (
                <div key={activity.id} className={styles.activityItem}>
                  <div className={styles.activityIcon} style={{ background: config.color }}>
                    {config.icon}
                  </div>
                  <div className={styles.activityContent}>
                    <span className={styles.activityDesc}>{activity.description}</span>
                    <span className={styles.activityTime}>
                      {relativeTime(activity.createdAt)}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        )}

        {hasMore && (
          <div className={styles.loadMore}>
            <Button variant="secondary" size="sm" onClick={handleLoadMore} disabled={loadingMore}>
              {loadingMore ? 'Loading...' : 'Load More'}
            </Button>
          </div>
        )}
      </section>
    </div>
  )
}
