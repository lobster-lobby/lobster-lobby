import { useState, useEffect, Suspense, lazy } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { TabNav, Button, Badge, UserBadge, Toast, Spinner, Card, Skeleton } from '../components/ui'
import type { Policy } from '../components/PolicyCard'
import { useAuth } from '../hooks/useAuth'
import styles from './PolicyDetail.module.css'

const TYPE_CONFIG = {
  existing_law: { label: 'Existing Law', className: 'typeSolid' },
  active_bill: { label: 'Active Bill', className: 'typeOutlined' },
  proposed: { label: 'Proposed', className: 'typeDashed' },
} as const

const TABS = [
  { id: 'debate', label: 'Debate' },
  { id: 'research', label: 'Research' },
  { id: 'representatives', label: 'Representatives' },
  { id: 'polls', label: 'Polls' },
  { id: 'draft', label: 'Draft' },
]

// Lazy-loaded tab content components
const DebateTab = lazy(() => import('../components/tabs/DebateTab'))
const ResearchTab = lazy(() => import('../components/tabs/ResearchTab'))
const RepresentativesTab = lazy(() => import('../components/tabs/RepresentativesTab'))
const PollsTab = lazy(() => import('../components/tabs/PollsTab'))
const DraftTab = lazy(() => import('../components/tabs/DraftTab'))

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

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  })
}

export default function PolicyDetail() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const { isAuthenticated } = useAuth()

  const [policy, setPolicy] = useState<Policy | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isBookmarked, setIsBookmarked] = useState(false)
  const [bookmarkLoading, setBookmarkLoading] = useState(false)
  const [toast, setToast] = useState<{ message: string; variant: 'success' | 'error' | 'info' } | null>(null)

  const activeTab = searchParams.get('tab') || 'debate'

  useEffect(() => {
    const fetchPolicy = async () => {
      if (!slug) return

      setLoading(true)
      setError(null)

      try {
        const token = localStorage.getItem('ll_token')
        const headers: HeadersInit = {}
        if (token) {
          headers['Authorization'] = `Bearer ${token}`
        }

        const res = await fetch(`/api/policies/${slug}`, { headers })
        if (!res.ok) {
          if (res.status === 404) {
            setError('Policy not found')
          } else {
            setError('Failed to load policy')
          }
          setLoading(false)
          return
        }

        const data = await res.json()
        setPolicy(data.policy)
        document.title = `${data.policy.title} | Lobster Lobby`
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Something went wrong')
      } finally {
        setLoading(false)
      }
    }

    fetchPolicy()
  }, [slug])

  const handleTabChange = (tabId: string) => {
    if (tabId === 'debate') {
      searchParams.delete('tab')
      setSearchParams(searchParams, { replace: true })
    } else {
      searchParams.set('tab', tabId)
      setSearchParams(searchParams, { replace: true })
    }
  }

  const handleTagClick = (tag: string) => {
    navigate(`/?tags=${encodeURIComponent(tag)}`)
  }

  const handleBookmark = async () => {
    if (!isAuthenticated) {
      setToast({ message: 'Please log in to bookmark policies', variant: 'info' })
      return
    }

    if (!policy) return

    setBookmarkLoading(true)
    try {
      const token = localStorage.getItem('ll_token')
      const res = await fetch(`/api/policies/${policy.id}/bookmark`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (!res.ok) throw new Error('Failed to toggle bookmark')

      setIsBookmarked(!isBookmarked)
      setToast({
        message: isBookmarked ? 'Removed from bookmarks' : 'Added to bookmarks',
        variant: 'success',
      })
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : 'Failed to bookmark',
        variant: 'error',
      })
    } finally {
      setBookmarkLoading(false)
    }
  }

  const handleShare = () => {
    const url = window.location.href
    navigator.clipboard.writeText(url).then(
      () => {
        setToast({ message: 'Link copied to clipboard', variant: 'success' })
      },
      () => {
        setToast({ message: 'Failed to copy link', variant: 'error' })
      }
    )
  }

  const renderTabContent = () => {
    switch (activeTab) {
      case 'debate':
        return (
          <Suspense fallback={<Spinner size="lg" />}>
            <DebateTab policyId={policy?.id || ''} />
          </Suspense>
        )
      case 'research':
        return (
          <Suspense fallback={<Spinner size="lg" />}>
            <ResearchTab policyId={policy?.id || ''} />
          </Suspense>
        )
      case 'representatives':
        return (
          <Suspense fallback={<Spinner size="lg" />}>
            <RepresentativesTab policyId={policy?.id || ''} />
          </Suspense>
        )
      case 'polls':
        return (
          <Suspense fallback={<Spinner size="lg" />}>
            <PollsTab />
          </Suspense>
        )
      case 'draft':
        return (
          <Suspense fallback={<Spinner size="lg" />}>
            <DraftTab />
          </Suspense>
        )
      default:
        return null
    }
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.main}>
          <Skeleton height={40} width="80%" />
          <Skeleton height={60} width="100%" style={{ marginTop: 'var(--ll-space-lg)' }} />
          <Skeleton height={200} width="100%" style={{ marginTop: 'var(--ll-space-lg)' }} />
        </div>
      </div>
    )
  }

  if (error || !policy) {
    return (
      <div className={styles.container}>
        <div className={styles.main}>
          <Card>
            <div className={styles.errorState}>
              <h2>{error === 'Policy not found' ? '404' : 'Error'}</h2>
              <p>{error || 'Policy not found'}</p>
              <Button onClick={() => navigate('/')}>Go to Feed</Button>
            </div>
          </Card>
        </div>
      </div>
    )
  }

  const typeInfo = TYPE_CONFIG[policy.type]

  return (
    <div className={styles.container}>
      <div className={styles.main}>
        {/* Header Section */}
        <header className={styles.header}>
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
          </div>

          <h1 className={styles.title}>{policy.title}</h1>

          {policy.externalUrl && (
            <a
              href={policy.externalUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={styles.externalLink}
            >
              View on {policy.level === 'federal' ? 'Congress.gov' : 'State Website'}
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                <polyline points="15 3 21 3 21 9" />
                <line x1="10" y1="14" x2="21" y2="3" />
              </svg>
            </a>
          )}

          {policy.tags.length > 0 && (
            <div className={styles.tags}>
              {policy.tags.map((tag) => (
                <button
                  key={tag}
                  type="button"
                  className={styles.tag}
                  onClick={() => handleTagClick(tag)}
                >
                  {tag}
                </button>
              ))}
            </div>
          )}

          <div className={styles.meta}>
            <UserBadge username={policy.createdBy} type={policy.creatorType ?? 'human'} />
            <span className={styles.date}>
              {formatDate(policy.createdAt)} ({relativeTime(policy.createdAt)})
            </span>
          </div>

          <div className={styles.actions}>
            <Button
              variant="secondary"
              size="sm"
              onClick={handleBookmark}
              disabled={bookmarkLoading}
            >
              {isBookmarked ? '🔖 Bookmarked' : '🔖 Bookmark'}
            </Button>
            <Button variant="secondary" size="sm" onClick={handleShare}>
              🔗 Share
            </Button>
          </div>

          <div className={styles.stats}>
            <span className={styles.stat}>💬 {policy.engagement.debateCount} debates</span>
            <span className={styles.stat}>🔬 {policy.engagement.researchCount} research</span>
            <span className={styles.stat}>📊 {policy.engagement.pollCount} polls</span>
            <span className={styles.stat}>🔖 {policy.engagement.bookmarkCount} saved</span>
            <span className={styles.stat}>👁️ {policy.engagement.viewCount} views</span>
          </div>
        </header>

        {/* Summary Section */}
        <section className={styles.summary}>
          <h2 className={styles.summaryTitle}>Summary</h2>
          <div className={styles.summaryText}>
            {policy.summary.split('\n').map((paragraph, idx) => (
              <p key={idx}>{paragraph}</p>
            ))}
          </div>
        </section>

        {/* Tabbed Interface */}
        <section className={styles.tabSection}>
          <TabNav
            tabs={TABS}
            activeTab={activeTab}
            onTabChange={handleTabChange}
            className={styles.tabNav}
          />
          <div className={styles.tabContent}>
            {renderTabContent()}
          </div>
        </section>
      </div>

      {/* Sidebar (Desktop only) */}
      <aside className={styles.sidebar}>
        <Card header={<h3>Quick Stats</h3>}>
          <div className={styles.sidebarStats}>
            <div className={styles.sidebarStat}>
              <span className={styles.sidebarStatLabel}>Total Engagement</span>
              <span className={styles.sidebarStatValue}>
                {policy.engagement.debateCount +
                  policy.engagement.researchCount +
                  policy.engagement.pollCount}
              </span>
            </div>
            <div className={styles.sidebarStat}>
              <span className={styles.sidebarStatLabel}>Hot Score</span>
              <span className={styles.sidebarStatValue}>{policy.hotScore.toFixed(1)}</span>
            </div>
            <div className={styles.sidebarStat}>
              <span className={styles.sidebarStatLabel}>Views</span>
              <span className={styles.sidebarStatValue}>{policy.engagement.viewCount}</span>
            </div>
          </div>
        </Card>

        <Card header={<h3>Related Policies</h3>}>
          <div className={styles.placeholderList}>
            <Skeleton height={60} />
            <Skeleton height={60} />
            <Skeleton height={60} />
          </div>
        </Card>

        <Card header={<h3>Top Contributors</h3>}>
          <div className={styles.placeholderList}>
            <Skeleton height={40} />
            <Skeleton height={40} />
            <Skeleton height={40} />
          </div>
        </Card>

        <Card header={<h3>Tags</h3>}>
          <div className={styles.tagCloud}>
            {policy.tags.map((tag) => (
              <button
                key={tag}
                type="button"
                className={styles.tagCloudItem}
                onClick={() => handleTagClick(tag)}
              >
                {tag}
              </button>
            ))}
          </div>
        </Card>
      </aside>

      {/* Toast Notifications */}
      {toast && (
        <div className={styles.toastContainer}>
          <Toast
            variant={toast.variant}
            onClose={() => setToast(null)}
            autoDismiss={3000}
          >
            {toast.message}
          </Toast>
        </div>
      )}
    </div>
  )
}
