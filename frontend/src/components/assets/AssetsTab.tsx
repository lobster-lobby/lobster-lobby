import { useState, useEffect, useCallback } from 'react'
import { Button, Spinner, EmptyState, Pagination, Toast } from '../ui'
import { AssetCard } from './AssetCard'
import { AssetForm } from './AssetForm'
import type { CampaignAsset, AssetListResponse } from '../../types/asset'
import { ASSET_TYPE_LABELS } from '../../types/asset'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import styles from './AssetsTab.module.css'

interface AssetsTabProps {
  campaignId: string
}

type SortOption = 'top' | 'newest' | 'most_downloaded' | 'most_shared'

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'top', label: 'Top' },
  { value: 'newest', label: 'Newest' },
  { value: 'most_downloaded', label: 'Most Downloaded' },
  { value: 'most_shared', label: 'Most Shared' },
]

const TYPE_OPTIONS: { value: string; label: string }[] = [
  { value: '', label: 'All Types' },
  ...Object.entries(ASSET_TYPE_LABELS).map(([value, label]) => ({ value, label })),
]

export default function AssetsTab({ campaignId }: AssetsTabProps) {
  const { isAuthenticated } = useAuth()
  const [assets, setAssets] = useState<CampaignAsset[]>([])
  const [userVotes, setUserVotes] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [sort, setSort] = useState<SortOption>('top')
  const [typeFilter, setTypeFilter] = useState<string>('')
  const [showForm, setShowForm] = useState(false)
  const [toast, setToast] = useState<{ message: string; variant: 'success' | 'error' | 'info' } | null>(null)

  const perPage = 10

  const fetchAssets = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const params = new URLSearchParams({
        page: page.toString(),
        perPage: perPage.toString(),
        sort,
      })
      if (typeFilter) {
        params.set('type', typeFilter)
      }

      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }

      const res = await fetch(`/api/campaigns/${campaignId}/assets?${params}`, { headers })
      if (!res.ok) throw new Error('Failed to fetch assets')

      const data: AssetListResponse = await res.json()
      setAssets(data.assets)
      setTotal(data.total)

      // Fetch user votes for each asset if authenticated
      if (token && data.assets.length > 0) {
        const votes: Record<string, number> = {}
        await Promise.all(
          data.assets.map(async (asset) => {
            try {
              const voteRes = await fetch(`/api/campaigns/${campaignId}/assets/${asset.id}`, {
                headers: { 'Authorization': `Bearer ${token}` },
              })
              if (voteRes.ok) {
                const voteData = await voteRes.json()
                votes[asset.id] = voteData.userVote || 0
              }
            } catch {
              // Ignore individual vote fetch errors
            }
          })
        )
        setUserVotes(votes)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }, [campaignId, page, sort, typeFilter])

  useEffect(() => {
    fetchAssets()
  }, [fetchAssets])

  const handleVote = async (assetId: string, value: number) => {
    if (!isAuthenticated) {
      setToast({ message: 'Please log in to vote', variant: 'info' })
      return
    }

    const token = getAccessToken()
    try {
      const res = await fetch(`/api/campaigns/${campaignId}/assets/${assetId}/vote`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ value }),
      })

      if (!res.ok) throw new Error('Failed to vote')

      const data = await res.json()
      setAssets((prev) =>
        prev.map((a) => (a.id === assetId ? data.asset : a))
      )
      setUserVotes((prev) => ({ ...prev, [assetId]: data.userVote }))
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : 'Failed to vote',
        variant: 'error',
      })
    }
  }

  const handleShare = async (assetId: string, platform: string) => {
    if (!isAuthenticated) {
      // Still allow sharing but don't track
      return
    }

    const token = getAccessToken()
    try {
      const res = await fetch(`/api/campaigns/${campaignId}/assets/${assetId}/share`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ platform }),
      })

      if (res.ok) {
        const data = await res.json()
        setAssets((prev) =>
          prev.map((a) => (a.id === assetId ? data.asset : a))
        )
      }
    } catch {
      // Silently fail share tracking
    }
  }

  const handleDownload = async (assetId: string) => {
    const token = getAccessToken()
    try {
      await fetch(`/api/campaigns/${campaignId}/assets/${assetId}/download`, {
        method: 'POST',
        headers: token ? { 'Authorization': `Bearer ${token}` } : {},
      })
      // Refresh to get updated download count
      fetchAssets()
    } catch {
      // Silently fail download tracking
    }
  }

  const handleFormSuccess = () => {
    setShowForm(false)
    setToast({ message: 'Asset created successfully!', variant: 'success' })
    setPage(1)
    fetchAssets()
  }

  const totalPages = Math.ceil(total / perPage)

  if (loading && assets.length === 0) {
    return (
      <div className={styles.loading}>
        <Spinner size="lg" />
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2 className={styles.title}>Campaign Assets ({total})</h2>
        {isAuthenticated && (
          <Button onClick={() => setShowForm(true)}>
            Submit Asset
          </Button>
        )}
      </div>

      {showForm && (
        <div className={styles.formWrapper}>
          <AssetForm
            campaignId={campaignId}
            onSuccess={handleFormSuccess}
            onCancel={() => setShowForm(false)}
          />
        </div>
      )}

      <div className={styles.filters}>
        <div className={styles.filterGroup}>
          <label>Sort by:</label>
          <select
            value={sort}
            onChange={(e) => {
              setSort(e.target.value as SortOption)
              setPage(1)
            }}
            className={styles.select}
          >
            {SORT_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className={styles.filterGroup}>
          <label>Type:</label>
          <select
            value={typeFilter}
            onChange={(e) => {
              setTypeFilter(e.target.value)
              setPage(1)
            }}
            className={styles.select}
          >
            {TYPE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {error && (
        <div className={styles.error}>{error}</div>
      )}

      {assets.length === 0 && !loading ? (
        <EmptyState
          heading="No assets yet"
          description={
            isAuthenticated
              ? 'Be the first to submit an asset for this campaign!'
              : 'No advocacy materials have been submitted yet.'
          }
          action={
            isAuthenticated ? (
              <Button onClick={() => setShowForm(true)}>Submit Asset</Button>
            ) : undefined
          }
        />
      ) : (
        <>
          <div className={styles.assetList}>
            {assets.map((asset) => (
              <AssetCard
                key={asset.id}
                asset={asset}
                userVote={userVotes[asset.id] || 0}
                onVote={handleVote}
                onShare={handleShare}
                onDownload={handleDownload}
              />
            ))}
          </div>

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <Pagination
                currentPage={page}
                totalPages={totalPages}
                onPageChange={setPage}
              />
            </div>
          )}
        </>
      )}

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
