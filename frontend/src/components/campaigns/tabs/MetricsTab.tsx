import { useState, useEffect, useCallback } from 'react'
import { Spinner } from '../../ui'
import type { CampaignActivityResponse, CampaignEvent, DailyActivity } from '../../../types/campaign'
import { EVENT_TYPE_CONFIG } from '../../../types/campaign'
import styles from './MetricsTab.module.css'

interface MetricsTabProps {
  campaignId: string
}

function relativeTime(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const seconds = Math.floor((now - then) / 1000)

  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo ago`
  return `${Math.floor(months / 12)}y ago`
}

interface BarChartProps {
  data: DailyActivity[]
}

function BarChart({ data }: BarChartProps) {
  if (data.length === 0) {
    return <div className={styles.emptyChart}>No activity data yet</div>
  }

  const maxCount = Math.max(...data.map((d) => d.count), 1)

  return (
    <div className={styles.chart}>
      <div className={styles.chartBars}>
        {data.map((day) => (
          <div key={day.date} className={styles.barContainer}>
            <div
              className={styles.bar}
              style={{ height: `${(day.count / maxCount) * 100}%` }}
              title={`${day.date}: ${day.count} events`}
            />
          </div>
        ))}
      </div>
      <div className={styles.chartLabels}>
        <span>{data[0]?.date.slice(5)}</span>
        <span>{data[data.length - 1]?.date.slice(5)}</span>
      </div>
    </div>
  )
}

interface StatCardProps {
  label: string
  value: number | string
  sublabel?: string
}

function StatCard({ label, value, sublabel }: StatCardProps) {
  return (
    <div className={styles.statCard}>
      <span className={styles.statValue}>{value}</span>
      <span className={styles.statLabel}>{label}</span>
      {sublabel && <span className={styles.statSublabel}>{sublabel}</span>}
    </div>
  )
}

interface ActivityFeedProps {
  events: CampaignEvent[]
}

function ActivityFeed({ events }: ActivityFeedProps) {
  if (events.length === 0) {
    return <div className={styles.emptyFeed}>No recent activity</div>
  }

  return (
    <div className={styles.activityFeed}>
      {events.map((event) => {
        const config = EVENT_TYPE_CONFIG[event.type] || EVENT_TYPE_CONFIG.created
        return (
          <div key={event.id} className={styles.activityItem}>
            <div className={styles.activityIcon} style={{ background: config.color }}>
              {config.icon}
            </div>
            <div className={styles.activityContent}>
              <span className={styles.activityTitle}>{event.title}</span>
              <span className={styles.activityDesc}>{event.description}</span>
              <span className={styles.activityTime}>{relativeTime(event.createdAt)}</span>
            </div>
          </div>
        )
      })}
    </div>
  )
}

export default function MetricsTab({ campaignId }: MetricsTabProps) {
  const [data, setData] = useState<CampaignActivityResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const res = await fetch(`/api/campaigns/${campaignId}/metrics/activity`)
      if (!res.ok) throw new Error('Failed to fetch metrics')

      const result: CampaignActivityResponse = await res.json()
      setData(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }, [campaignId])

  useEffect(() => {
    fetchData()
  }, [fetchData])

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

  if (!data) {
    return null
  }

  const { dailyActivity, recentEvents, metrics } = data

  return (
    <div className={styles.container}>
      <h2 className={styles.title}>Campaign Metrics</h2>

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Overview</h3>
        <div className={styles.statsGrid}>
          <StatCard label="Total Downloads" value={metrics.totalDownloads} />
          <StatCard label="Total Shares" value={metrics.totalShares} />
          <StatCard label="Participants" value={metrics.uniqueParticipants} />
          <StatCard label="Assets" value={metrics.assetCount} />
          <StatCard label="Comments" value={metrics.commentCount} />
        </div>
      </section>

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Activity (Last 30 Days)</h3>
        <div className={styles.chartContainer}>
          <BarChart data={dailyActivity} />
        </div>
      </section>

      {Object.keys(metrics.sharesByPlatform || {}).length > 0 && (
        <section className={styles.section}>
          <h3 className={styles.sectionTitle}>Shares by Platform</h3>
          <div className={styles.platformGrid}>
            {Object.entries(metrics.sharesByPlatform).map(([platform, count]) => (
              <div key={platform} className={styles.platformItem}>
                <span className={styles.platformName}>{platform}</span>
                <span className={styles.platformCount}>{count}</span>
              </div>
            ))}
          </div>
        </section>
      )}

      <section className={styles.section}>
        <h3 className={styles.sectionTitle}>Recent Activity</h3>
        <ActivityFeed events={recentEvents} />
      </section>
    </div>
  )
}
