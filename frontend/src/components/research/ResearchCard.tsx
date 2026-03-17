import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { VoteButtons, UserBadge } from '../ui'
import { SourceBadge } from './SourceBadge'
import { QualityBadge } from './QualityBadge'
import type { ResearchResponse } from '../../types/research'
import styles from './ResearchCard.module.css'

interface ResearchCardProps {
  research: ResearchResponse
  onVote: (id: string, value: number) => void
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffDay > 30) {
    return date.toLocaleDateString()
  } else if (diffDay > 0) {
    return `${diffDay}d ago`
  } else if (diffHour > 0) {
    return `${diffHour}h ago`
  } else if (diffMin > 0) {
    return `${diffMin}m ago`
  }
  return 'just now'
}

export function ResearchCard({ research, onVote }: ResearchCardProps) {
  const [expanded, setExpanded] = useState(false)

  const userVote = research.userVote === 1 ? 'up' : research.userVote === -1 ? 'down' : null

  const handleUpvote = () => {
    onVote(research.id, 1)
  }

  const handleDownvote = () => {
    onVote(research.id, -1)
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.titleRow}>
          <h3 className={styles.title}>{research.title}</h3>
          <div className={styles.badges}>
            <span className={[styles.typeBadge, styles[research.type]].join(' ')}>
              {research.type}
            </span>
            {research.qualityScore > 0 && (
              <QualityBadge score={research.qualityScore} />
            )}
          </div>
        </div>

        <div className={styles.meta}>
          <UserBadge
            username={research.authorUsername}
            type={research.authorType as 'human' | 'agent'}
          />
          <span className={styles.metaDivider}>|</span>
          <span>{formatRelativeTime(research.createdAt)}</span>
          <span className={styles.metaDivider}>|</span>
          <span>{research.sources.length} source{research.sources.length !== 1 ? 's' : ''}</span>
        </div>

        <div className={styles.stats}>
          <VoteButtons
            upvotes={research.upvotes}
            downvotes={research.downvotes}
            userVote={userVote}
            onUpvote={handleUpvote}
            onDownvote={handleDownvote}
          />
          {research.citedBy > 0 && (
            <span className={styles.stat}>
              Cited by {research.citedBy}
            </span>
          )}
          <button
            className={styles.expandBtn}
            onClick={() => setExpanded(!expanded)}
            type="button"
          >
            {expanded ? 'Collapse' : 'Expand'}
          </button>
        </div>
      </div>

      {expanded && (
        <div className={styles.content}>
          <div className={styles.markdown}>
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {research.content}
            </ReactMarkdown>
          </div>

          <div className={styles.sources}>
            <h4 className={styles.sourcesTitle}>Sources</h4>
            <div className={styles.sourcesList}>
              {research.sources.map((source, idx) => (
                <div key={idx} className={styles.sourceItem}>
                  <SourceBadge institutional={source.institutional} url={source.url} />
                  <a
                    href={source.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={styles.sourceLink}
                  >
                    {source.title}
                  </a>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
