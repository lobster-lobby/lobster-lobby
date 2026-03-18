import { useState, useEffect, useCallback, type FormEvent } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Pagination } from '../components/ui'
import { CampaignCard, type CampaignCardData } from '../components/campaigns/CampaignCard'
import styles from './Campaigns.module.css'

const PER_PAGE = 12

const SORT_OPTIONS = [
  { value: 'trending', label: 'Trending' },
  { value: 'newest', label: 'Newest' },
  { value: 'participants', label: 'Most Supporters' },
  { value: 'shares', label: 'Most Shared' },
]

const STATUS_OPTIONS = [
  { value: '', label: 'All Campaigns' },
  { value: 'active', label: 'Active' },
  { value: 'completed', label: 'Completed' },
  { value: 'paused', label: 'Paused' },
]

export default function Campaigns() {
  const [searchParams, setSearchParams] = useSearchParams()

  const [campaigns, setCampaigns] = useState<CampaignCardData[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Get filter values from URL
  const page = parseInt(searchParams.get('page') || '1', 10)
  const sort = searchParams.get('sort') || 'trending'
  const status = searchParams.get('status') || ''
  const searchQuery = searchParams.get('q') || ''

  const [searchInput, setSearchInput] = useState(searchQuery)

  const updateParams = useCallback(
    (updates: Record<string, string>) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev)
        for (const [key, value] of Object.entries(updates)) {
          if (value) {
            next.set(key, value)
          } else {
            next.delete(key)
          }
        }
        // Reset page when filters change (except page itself)
        if (!('page' in updates)) {
          next.delete('page')
        }
        return next
      }, { replace: true })
    },
    [setSearchParams]
  )

  const fetchCampaigns = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('perPage', String(PER_PAGE))
      params.set('sort', sort)
      if (status) params.set('status', status)
      if (searchQuery) params.set('q', searchQuery)

      const res = await fetch(`/api/campaigns?${params}`)
      if (!res.ok) throw new Error('Failed to load campaigns')

      const data = await res.json()
      setCampaigns(data.campaigns || [])
      setTotal(data.total || 0)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
      setCampaigns([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [page, sort, status, searchQuery])

  useEffect(() => {
    fetchCampaigns()
  }, [fetchCampaigns])

  useEffect(() => {
    setSearchInput(searchQuery)
  }, [searchQuery])

  const handleSearch = (e: FormEvent) => {
    e.preventDefault()
    updateParams({ q: searchInput.trim() })
  }

  const handleSortChange = (newSort: string) => {
    updateParams({ sort: newSort })
  }

  const handleStatusChange = (newStatus: string) => {
    updateParams({ status: newStatus })
  }

  const handlePageChange = (newPage: number) => {
    updateParams({ page: String(newPage) })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const clearFilters = () => {
    setSearchParams({}, { replace: true })
    setSearchInput('')
  }

  const totalPages = Math.ceil(total / PER_PAGE)
  const hasActiveFilters = status || searchQuery

  return (
    <div className={styles.page}>
      <header className={styles.pageHeader}>
        <h1 className={styles.pageTitle}>Campaigns</h1>
        <p className={styles.pageSubtitle}>
          Discover grassroots campaigns advocating for policy change
        </p>
      </header>

      {/* Search & Filters */}
      <div className={styles.filters}>
        <form className={styles.searchForm} onSubmit={handleSearch}>
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Search campaigns..."
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
          <button type="submit" className={styles.searchBtn}>Search</button>
        </form>

        <div className={styles.filterRow}>
          <select
            className={styles.filterSelect}
            value={sort}
            onChange={(e) => handleSortChange(e.target.value)}
            aria-label="Sort campaigns"
          >
            {SORT_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>

          <select
            className={styles.filterSelect}
            value={status}
            onChange={(e) => handleStatusChange(e.target.value)}
            aria-label="Filter by status"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>

          {hasActiveFilters && (
            <button
              type="button"
              className={styles.clearBtn}
              onClick={clearFilters}
            >
              Clear filters
            </button>
          )}
        </div>
      </div>

      {/* Results count */}
      {!loading && !error && (
        <p className={styles.resultsCount}>
          {total} campaign{total !== 1 ? 's' : ''} found
          {searchQuery && <> for &ldquo;{searchQuery}&rdquo;</>}
        </p>
      )}

      {/* Campaign Grid */}
      {loading ? (
        <div className={styles.loading}>
          <div className={styles.spinner} />
        </div>
      ) : error ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>⚠️</div>
          <h2 className={styles.emptyTitle}>Something went wrong</h2>
          <p className={styles.emptyText}>{error}</p>
          <button className={styles.retryBtn} onClick={fetchCampaigns}>Try again</button>
        </div>
      ) : campaigns.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>📢</div>
          <h2 className={styles.emptyTitle}>No campaigns found</h2>
          <p className={styles.emptyText}>
            {hasActiveFilters
              ? 'Try adjusting your search or filters.'
              : 'Be the first to start a campaign!'}
          </p>
          {hasActiveFilters && (
            <button className={styles.clearBtn} onClick={clearFilters}>
              Clear filters
            </button>
          )}
        </div>
      ) : (
        <>
          <div className={styles.grid}>
            {campaigns.map((campaign) => (
              <CampaignCard key={campaign.id} campaign={campaign} />
            ))}
          </div>

          {totalPages > 1 && (
            <Pagination
              currentPage={page}
              totalPages={totalPages}
              onPageChange={handlePageChange}
            />
          )}
        </>
      )}
    </div>
  )
}
