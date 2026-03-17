import { Link } from 'react-router-dom'
import { Badge, UserBadge } from './ui'
import { relativeTime } from '../utils/time'
import styles from './PolicyCard.module.css'

export interface PolicyEngagement {
  debateCount: number
  researchCount: number
  pollCount: number
  bookmarkCount: number
  viewCount: number
}

export interface Policy {
  id: string
  title: string
  slug: string
  summary: string
  type: 'existing_law' | 'active_bill' | 'proposed'
  level: 'federal' | 'state'
  state?: string
  status: string
  externalUrl?: string
  billNumber?: string
  tags: string[]
  createdBy: string
  creatorType?: 'human' | 'agent'
  engagement: PolicyEngagement
  hotScore: number
  createdAt: string
  updatedAt: string
}

interface PolicyCardProps {
  policy: Policy
  onTagClick?: (tag: string) => void
}

const TYPE_CONFIG = {
  existing_law: { label: 'Existing Law', className: 'typeSolid' },
  active_bill: { label: 'Active Bill', className: 'typeOutlined' },
  proposed: { label: 'Proposed', className: 'typeDashed' },
} as const

export function PolicyCard({ policy, onTagClick }: PolicyCardProps) {
  const typeInfo = TYPE_CONFIG[policy.type]
  const { engagement } = policy

  return (
    <article className={styles.card}>
      <div className={styles.header}>
        <div className={styles.badges}>
          <span className={`${styles.typeBadge} ${styles[typeInfo.className]}`}>
            {typeInfo.label}
          </span>
          {policy.level === 'federal' ? (
            <Badge variant="default">🇺🇸 Federal</Badge>
          ) : policy.state ? (
            <Badge variant="default">📍 {policy.state}</Badge>
          ) : null}
          {policy.billNumber && (
            <span className={styles.billNumber}>{policy.billNumber}</span>
          )}
          {policy.status === 'ready_for_campaign' && (
            <Badge variant="success">Ready for Campaign</Badge>
          )}
        </div>
        <time className={styles.time} dateTime={policy.createdAt}>
          {relativeTime(policy.createdAt)}
        </time>
      </div>

      <div className={styles.body}>
        <h3 className={styles.title}>
          <Link to={`/policies/${policy.slug}`}>{policy.title}</Link>
        </h3>
        <p className={styles.summary}>{policy.summary}</p>
      </div>

      {policy.tags.length > 0 && (
        <div className={styles.tags}>
          {policy.tags.map((tag) => (
            <button
              key={tag}
              type="button"
              className={styles.tag}
              onClick={() => onTagClick?.(tag)}
            >
              {tag}
            </button>
          ))}
        </div>
      )}

      <div className={styles.footer}>
        <div className={styles.stats}>
          <span className={styles.stat} aria-label="Debates">💬 {engagement.debateCount}</span>
          <span className={styles.stat} aria-label="Research">🔬 {engagement.researchCount}</span>
          <span className={styles.stat} aria-label="Polls">📊 {engagement.pollCount}</span>
          <span className={styles.stat} aria-label="Saved">🔖 {engagement.bookmarkCount}</span>
        </div>
        <UserBadge username={policy.createdBy} type={policy.creatorType ?? 'human'} />
      </div>
    </article>
  )
}
