import { useState, useEffect, Suspense, lazy } from 'react'
import { useParams, useNavigate, useLocation } from 'react-router-dom'
import { Card, Badge, Button, Spinner, Skeleton, TabNav } from '../components/ui'
import { relativeTime } from '../utils/time'
import styles from './CampaignDetail.module.css'

const AssetsTab = lazy(() => import('../components/assets/AssetsTab'))
const DiscussionTab = lazy(() => import('../components/campaigns/tabs/DiscussionTab'))
const MetricsTab = lazy(() => import('../components/campaigns/tabs/MetricsTab'))
const TimelineTab = lazy(() => import('../components/campaigns/tabs/TimelineTab'))

type TabId = 'assets' | 'discussion' | 'metrics' | 'timeline'

const TABS: { id: TabId; label: string }[] = [
  { id: 'assets', label: 'Assets' },
  { id: 'discussion', label: 'Discussion' },
  { id: 'metrics', label: 'Metrics' },
  { id: 'timeline', label: 'Timeline' },
]

interface Campaign {
  id: string
  title: string
  slug: string
  policyId: string
  createdBy: string
  objective: string
  target: string
  description: string
  status: string
  metrics: {
    totalDownloads: number
    totalShares: number
    sharesByPlatform: Record<string, number>
    uniqueParticipants: number
    assetCount: number
    commentCount: number
  }
  trendingScore: number
  createdAt: string
  updatedAt: string
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  })
}

const STATUS_CONFIG: Record<string, { label: string; variant: 'success' | 'neutral' | 'default' }> = {
  active: { label: 'Active', variant: 'success' },
  paused: { label: 'Paused', variant: 'neutral' },
  completed: { label: 'Completed', variant: 'default' },
  archived: { label: 'Archived', variant: 'default' },
}

function getTabFromHash(hash: string): TabId {
  const tabId = hash.replace('#', '') as TabId
  return TABS.some((t) => t.id === tabId) ? tabId : 'assets'
}

export default function CampaignDetail() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const location = useLocation()

  const [campaign, setCampaign] = useState<Campaign | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabId>(() => getTabFromHash(location.hash))

  // Sync tab with URL hash
  useEffect(() => {
    setActiveTab(getTabFromHash(location.hash))
  }, [location.hash])

  const handleTabChange = (tabId: string) => {
    navigate(`#${tabId}`, { replace: true })
  }

  useEffect(() => {
    const fetchCampaign = async () => {
      if (!slug) return

      setLoading(true)
      setError(null)

      try {
        const res = await fetch(`/api/campaigns/${slug}`)
        if (!res.ok) {
          if (res.status === 404) {
            setError('Campaign not found')
          } else {
            setError('Failed to load campaign')
          }
          setLoading(false)
          return
        }

        const data = await res.json()
        setCampaign(data.campaign)
        document.title = `${data.campaign.title} | Lobster Lobby`
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Something went wrong')
      } finally {
        setLoading(false)
      }
    }

    fetchCampaign()

    return () => {
      document.title = 'Lobster Lobby'
    }
  }, [slug])

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

  if (error || !campaign) {
    return (
      <div className={styles.container}>
        <div className={styles.main}>
          <Card>
            <div className={styles.errorState}>
              <h2>{error === 'Campaign not found' ? '404' : 'Error'}</h2>
              <p>{error || 'Campaign not found'}</p>
              <Button onClick={() => navigate('/')}>Go to Feed</Button>
            </div>
          </Card>
        </div>
      </div>
    )
  }

  const statusInfo = STATUS_CONFIG[campaign.status] || STATUS_CONFIG.active

  return (
    <div className={styles.container}>
      <div className={styles.main}>
        {/* Header Section */}
        <header className={styles.header}>
          <div className={styles.badges}>
            <Badge variant={statusInfo.variant}>{statusInfo.label}</Badge>
          </div>

          <h1 className={styles.title}>{campaign.title}</h1>

          <div className={styles.meta}>
            <span className={styles.date}>
              Created {formatDate(campaign.createdAt)} ({relativeTime(campaign.createdAt)})
            </span>
          </div>

          <div className={styles.stats}>
            <span className={styles.stat}>{campaign.metrics.assetCount} assets</span>
            <span className={styles.stat}>{campaign.metrics.totalDownloads} downloads</span>
            <span className={styles.stat}>{campaign.metrics.totalShares} shares</span>
            <span className={styles.stat}>{campaign.metrics.uniqueParticipants} participants</span>
          </div>
        </header>

        {/* Objective & Target */}
        <section className={styles.section}>
          <div className={styles.infoGrid}>
            <div className={styles.infoCard}>
              <h3>Objective</h3>
              <p>{campaign.objective}</p>
            </div>
            <div className={styles.infoCard}>
              <h3>Target</h3>
              <p>{campaign.target}</p>
            </div>
          </div>
        </section>

        {/* Description */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Description</h2>
          <div className={styles.description}>
            {campaign.description.split('\n').map((paragraph, idx) => (
              <p key={idx}>{paragraph}</p>
            ))}
          </div>
        </section>

        {/* Tab Navigation */}
        <section className={styles.section}>
          <TabNav
            tabs={TABS}
            activeTab={activeTab}
            onTabChange={handleTabChange}
          />
        </section>

        {/* Tab Content */}
        <section className={styles.section}>
          <Suspense fallback={<div className={styles.tabLoading}><Spinner size="lg" /></div>}>
            {activeTab === 'assets' && <AssetsTab campaignId={campaign.id} />}
            {activeTab === 'discussion' && <DiscussionTab campaignId={campaign.id} />}
            {activeTab === 'metrics' && <MetricsTab campaignId={campaign.id} />}
            {activeTab === 'timeline' && <TimelineTab campaignId={campaign.id} />}
          </Suspense>
        </section>
      </div>

      {/* Sidebar */}
      <aside className={styles.sidebar}>
        <Card header={<h3>Campaign Stats</h3>}>
          <div className={styles.sidebarStats}>
            <div className={styles.sidebarStat}>
              <span className={styles.sidebarStatLabel}>Total Assets</span>
              <span className={styles.sidebarStatValue}>{campaign.metrics.assetCount}</span>
            </div>
            <div className={styles.sidebarStat}>
              <span className={styles.sidebarStatLabel}>Downloads</span>
              <span className={styles.sidebarStatValue}>{campaign.metrics.totalDownloads}</span>
            </div>
            <div className={styles.sidebarStat}>
              <span className={styles.sidebarStatLabel}>Shares</span>
              <span className={styles.sidebarStatValue}>{campaign.metrics.totalShares}</span>
            </div>
            <div className={styles.sidebarStat}>
              <span className={styles.sidebarStatLabel}>Trending Score</span>
              <span className={styles.sidebarStatValue}>{campaign.trendingScore.toFixed(1)}</span>
            </div>
          </div>
        </Card>

        {Object.keys(campaign.metrics.sharesByPlatform || {}).length > 0 && (
          <Card header={<h3>Shares by Platform</h3>}>
            <div className={styles.platformStats}>
              {Object.entries(campaign.metrics.sharesByPlatform).map(([platform, count]) => (
                <div key={platform} className={styles.platformStat}>
                  <span className={styles.platformName}>{platform}</span>
                  <span className={styles.platformCount}>{count}</span>
                </div>
              ))}
            </div>
          </Card>
        )}

        <Card header={<h3>Quick Links</h3>}>
          <div className={styles.quickLinks}>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => navigate(`/policies/${campaign.policyId}`)}
            >
              View Related Policy
            </Button>
          </div>
        </Card>
      </aside>
    </div>
  )
}
