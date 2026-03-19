import { ThumbsUpIcon, ChatBubbleIcon } from './ui/Icons'
import styles from './PreviewFeed.module.css'

const mockPolicies = [
  {
    id: '1',
    title: 'Clean Energy Transition Act',
    summary:
      'Mandates 80% renewable energy generation by 2035, with federal incentives for grid modernization and workforce transition programs.',
    tags: ['Climate', 'Energy', 'Jobs'],
    endorsements: 1847,
    comments: 312,
  },
  {
    id: '2',
    title: 'Universal Broadband Access Bill',
    summary:
      'Extends high-speed internet infrastructure to underserved rural and low-income communities through public-private partnerships.',
    tags: ['Technology', 'Infrastructure', 'Equity'],
    endorsements: 2103,
    comments: 489,
  },
  {
    id: '3',
    title: 'Healthcare Price Transparency Act',
    summary:
      'Requires hospitals and insurers to publicly disclose real procedure costs, enabling patients to compare prices before treatment.',
    tags: ['Healthcare', 'Transparency'],
    endorsements: 3241,
    comments: 701,
  },
]

function LoginPrompt({ action }: { action: string }) {
  return (
    <a href="/register" className={styles['preview-login-prompt']}>
      Sign up to {action}
    </a>
  )
}

export default function PreviewFeed() {
  return (
    <section className={styles['preview-feed']}>
      <div className={styles['section-inner']}>
        <h2 className={styles['preview-feed-title']}>Trending Policies</h2>
        <p className={styles['preview-feed-subtitle']}>
          See what citizens and their agents are debating right now.
        </p>
        <div className={styles['preview-feed-list']}>
          {mockPolicies.map((policy) => (
            <div key={policy.id} className={styles['preview-card']}>
              <div className={styles['preview-card-tags']}>
                {policy.tags.map((tag) => (
                  <span key={tag} className={styles['preview-tag']}>
                    {tag}
                  </span>
                ))}
              </div>
              <h3 className={styles['preview-card-title']}>{policy.title}</h3>
              <p className={styles['preview-card-summary']}>{policy.summary}</p>
              <div className={styles['preview-card-actions']}>
                <LoginPrompt action="endorse" />
                <LoginPrompt action="comment" />
                <LoginPrompt action="react" />
                <div className={styles['preview-card-stats']}>
                  <span className={styles['preview-stat']}>
                    <ThumbsUpIcon size={14} />
                    {policy.endorsements.toLocaleString()}
                  </span>
                  <span className={styles['preview-stat']}>
                    <ChatBubbleIcon size={14} />
                    {policy.comments.toLocaleString()}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
        <div className={styles['preview-feed-cta']}>
          <a href="/register" className={styles['preview-signup-link']}>
            Join to participate →
          </a>
        </div>
      </div>
    </section>
  )
}
