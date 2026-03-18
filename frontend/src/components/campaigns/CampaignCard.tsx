import { Link } from 'react-router-dom'
import { Badge } from '../ui'
import { relativeTime } from '../../utils/time'
import styles from './CampaignCard.module.css'

export interface CampaignCardData {
  id: string
  title: string
  slug: string
  objective: string
  status: string
  metrics: {
    totalDownloads: number
    totalShares: number
    uniqueParticipants: number
    assetCount: number
    commentCount: number
  }
  trendingScore: number
  createdAt: string
  creatorName?: string
  policyTitle?: string
  policySlug?: string
}

interface CampaignCardProps {
  campaign: CampaignCardData
}

const STATUS_CONFIG: Record<string, { label: string; variant: 'success' | 'neutral' | 'default' }> = {
  active: { label: 'Active', variant: 'success' },
  paused: { label: 'Paused', variant: 'neutral' },
  completed: { label: 'Completed', variant: 'default' },
  archived: { label: 'Archived', variant: 'default' },
}

function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return text.slice(0, maxLength).trimEnd() + '...'
}

export function CampaignCard({ campaign }: CampaignCardProps) {
  const statusInfo = STATUS_CONFIG[campaign.status] || STATUS_CONFIG.active
  const supporters = campaign.metrics.uniqueParticipants

  return (
    <Link to={`/campaigns/${campaign.slug}`} className={styles.card}>
      <div className={styles.header}>
        <Badge variant={statusInfo.variant}>{statusInfo.label}</Badge>
        {campaign.trendingScore > 0 && (
          <span className={styles.trending} title="Trending score">
            {campaign.trendingScore.toFixed(1)}
          </span>
        )}
      </div>

      <h3 className={styles.title}>{campaign.title}</h3>

      <p className={styles.objective}>{truncate(campaign.objective, 120)}</p>

      {campaign.policyTitle && (
        <div className={styles.policyBadge}>
          <span className={styles.policyIcon}>📜</span>
          <span className={styles.policyName}>{truncate(campaign.policyTitle, 40)}</span>
        </div>
      )}

      <div className={styles.stats}>
        <div className={styles.stat}>
          <span className={styles.statValue}>{supporters}</span>
          <span className={styles.statLabel}>supporters</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statValue}>{campaign.metrics.assetCount}</span>
          <span className={styles.statLabel}>assets</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statValue}>{campaign.metrics.totalShares}</span>
          <span className={styles.statLabel}>shares</span>
        </div>
      </div>

      <div className={styles.footer}>
        {campaign.creatorName && (
          <span className={styles.creator}>
            <span className={styles.creatorAvatar}>{campaign.creatorName.charAt(0).toUpperCase()}</span>
            {campaign.creatorName}
          </span>
        )}
        <span className={styles.time}>{relativeTime(campaign.createdAt)}</span>
      </div>
    </Link>
  )
}
