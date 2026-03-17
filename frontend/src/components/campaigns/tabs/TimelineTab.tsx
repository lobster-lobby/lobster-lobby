import { useState, useEffect, useCallback } from 'react'
import { Spinner, EmptyState } from '../../ui'
import type { CampaignEvent } from '../../../types/campaign'
import { EVENT_TYPE_CONFIG } from '../../../types/campaign'
import { relativeTime } from '../../../utils/time'
import styles from './TimelineTab.module.css'

interface TimelineTabProps {
  campaignId: string
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

interface TimelineItemProps {
  event: CampaignEvent
  isLast: boolean
}

function TimelineItem({ event, isLast }: TimelineItemProps) {
  const config = EVENT_TYPE_CONFIG[event.type] || EVENT_TYPE_CONFIG.created

  return (
    <div className={styles.timelineItem}>
      <div className={styles.timelineLeft}>
        <div className={styles.timelineIcon} style={{ background: config.color }}>
          {config.icon}
        </div>
        {!isLast && <div className={styles.timelineLine} />}
      </div>
      <div className={styles.timelineContent}>
        <div className={styles.timelineHeader}>
          <span className={styles.timelineTitle}>{event.title}</span>
          <span className={styles.timelineType} style={{ color: config.color }}>
            {config.label}
          </span>
        </div>
        <p className={styles.timelineDesc}>{event.description}</p>
        <div className={styles.timelineTime}>
          <span title={formatDate(event.createdAt)}>{relativeTime(event.createdAt)}</span>
        </div>
      </div>
    </div>
  )
}

export default function TimelineTab({ campaignId }: TimelineTabProps) {
  const [events, setEvents] = useState<CampaignEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchEvents = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const res = await fetch(`/api/campaigns/${campaignId}/events`)
      if (!res.ok) throw new Error('Failed to fetch timeline')

      const data = await res.json()
      setEvents(data.events || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }, [campaignId])

  useEffect(() => {
    fetchEvents()
  }, [fetchEvents])

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

  return (
    <div className={styles.container}>
      <h2 className={styles.title}>Campaign Timeline</h2>

      {events.length === 0 ? (
        <EmptyState
          heading="No events yet"
          description="Campaign activity will appear here as it happens."
        />
      ) : (
        <div className={styles.timeline}>
          {events.map((event, index) => (
            <TimelineItem
              key={event.id}
              event={event}
              isLast={index === events.length - 1}
            />
          ))}
        </div>
      )}
    </div>
  )
}
