import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { TabNav, SearchBar, Button, Spinner, EmptyState, Skeleton } from '../components/ui'
import { PolicyCard } from '../components/PolicyCard'
import type { Policy } from '../components/PolicyCard'
import styles from './PolicyFeed.module.css'

const SORT_TABS = [
  { id: 'hot', label: '🔥 Hot' },
  { id: 'new', label: '🆕 New' },
  { id: 'top', label: '📈 Top' },
  { id: 'debated', label: '🗣️ Most Debated' },
]

const TOP_RANGES = [
  { id: 'week', label: 'This Week' },
  { id: 'month', label: 'This Month' },
  { id: 'all', label: 'All Time' },
]

const TYPE_OPTIONS = [
  { value: 'existing_law', label: 'Existing Law' },
  { value: 'active_bill', label: 'Active Bill' },
  { value: 'proposed', label: 'Proposed' },
]

const LEVEL_OPTIONS = [
  { value: 'federal', label: 'Federal' },
  { value: 'state', label: 'State' },
]

const POPULAR_TAGS = [
  'ai', 'healthcare', 'climate', 'education', 'privacy',
  'housing', 'immigration', 'taxes', 'defense', 'infrastructure',
]

const US_STATES = [
  'AL','AK','AZ','AR','CA','CO','CT','DE','FL','GA',
  'HI','ID','IL','IN','IA','KS','KY','LA','ME','MD',
  'MA','MI','MN','MS','MO','MT','NE','NV','NH','NJ',
  'NM','NY','NC','ND','OH','OK','OR','PA','RI','SC',
  'SD','TN','TX','UT','VT','VA','WA','WV','WI','WY',
]

const PER_PAGE = 20

interface FeedState {
  policies: Policy[]
  total: number
  loading: boolean
  loadingMore: boolean
  error: string | null
}

export default function PolicyFeed() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [feedState, setFeedState] = useState<FeedState>({
    policies: [],
    total: 0,
    loading: true,
    loadingMore: false,
    error: null,
  })
  const pageRef = useRef(1)
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  // Read filter state from URL
  const sort = searchParams.get('sort') || 'hot'
  const topRange = searchParams.get('topRange') || 'week'
  const search = searchParams.get('q') || ''
  const selectedTypes = useMemo(
    () => searchParams.get('type')?.split(',').filter(Boolean) || [],
    [searchParams],
  )
  const selectedLevels = useMemo(
    () => searchParams.get('level')?.split(',').filter(Boolean) || [],
    [searchParams],
  )
  const selectedState = searchParams.get('state') || ''
  const selectedTags = useMemo(
    () => searchParams.get('tags')?.split(',').filter(Boolean) || [],
    [searchParams],
  )

  // Debounced search for API requests
  const [debouncedSearch, setDebouncedSearch] = useState(search)
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300)
    return () => clearTimeout(timer)
  }, [search])

  const hasActiveFilters = selectedTypes.length > 0 || selectedLevels.length > 0 ||
    selectedState !== '' || selectedTags.length > 0 || search !== ''

  // URL update helper — preserves existing params, removes empty ones
  const updateParams = useCallback((updates: Record<string, string>) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      for (const [key, value] of Object.entries(updates)) {
        if (value) {
          next.set(key, value)
        } else {
          next.delete(key)
        }
      }
      return next
    }, { replace: true })
  }, [setSearchParams])

  // Fetch policies from API
  const fetchPolicies = useCallback(async (page: number, append: boolean) => {
    setFeedState((s) => ({
      ...s,
      loading: !append,
      loadingMore: append,
      error: null,
    }))

    const params = new URLSearchParams()
    params.set('page', String(page))
    params.set('perPage', String(PER_PAGE))
    params.set('sort', sort === 'top' ? `top_${topRange}` : sort)
    if (selectedTypes.length > 0) params.set('type', selectedTypes.join(','))
    if (selectedLevels.length > 0) params.set('level', selectedLevels.join(','))
    if (selectedState) params.set('state', selectedState)
    if (selectedTags.length > 0) params.set('tags', selectedTags.join(','))
    if (debouncedSearch) params.set('q', debouncedSearch)

    try {
      const res = await fetch(`/api/policies?${params.toString()}`)
      if (!res.ok) throw new Error('Failed to fetch policies')
      const data = await res.json()
      const fetched: Policy[] = data.policies || []

      setFeedState((s) => ({
        policies: append ? [...s.policies, ...fetched] : fetched,
        total: data.total ?? 0,
        loading: false,
        loadingMore: false,
        error: null,
      }))
    } catch (err) {
      setFeedState((s) => ({
        ...s,
        loading: false,
        loadingMore: false,
        error: err instanceof Error ? err.message : 'Something went wrong',
      }))
    }
  }, [sort, topRange, selectedTypes, selectedLevels, selectedState, selectedTags, debouncedSearch])

  // Reset and fetch on filter/sort change
  useEffect(() => {
    pageRef.current = 1
    fetchPolicies(1, false)
  }, [fetchPolicies])

  // Load more handler
  const handleLoadMore = () => {
    const nextPage = pageRef.current + 1
    pageRef.current = nextPage
    fetchPolicies(nextPage, true)
  }

  const canLoadMore = feedState.policies.length < feedState.total

  // Sort handlers
  const handleSortChange = (tabId: string) => {
    updateParams({ sort: tabId === 'hot' ? '' : tabId, topRange: '' })
  }

  const handleTopRangeChange = (range: string) => {
    updateParams({ sort: 'top', topRange: range === 'week' ? '' : range })
  }

  // Search handler (debounced)
  const handleSearchChange = (value: string) => {
    clearTimeout(searchTimerRef.current)
    searchTimerRef.current = setTimeout(() => {
      updateParams({ q: value })
    }, 300)
  }

  // Filter toggle helpers
  const toggleArrayParam = (key: string, value: string, current: string[]) => {
    const next = current.includes(value)
      ? current.filter((v) => v !== value)
      : [...current, value]
    updateParams({ [key]: next.join(',') })
  }

  const handleTagClick = (tag: string) => {
    toggleArrayParam('tags', tag, selectedTags)
  }

  const removeFilter = (key: string, value?: string) => {
    if (value && (key === 'type' || key === 'level' || key === 'tags')) {
      const current = key === 'type' ? selectedTypes : key === 'level' ? selectedLevels : selectedTags
      const next = current.filter((v) => v !== value)
      updateParams({ [key]: next.join(',') })
    } else {
      updateParams({ [key]: '' })
    }
  }

  const clearAllFilters = () => {
    updateParams({ type: '', level: '', state: '', tags: '', q: '' })
  }

  // Filter panel content (shared between sidebar and drawer)
  const filterContent = (
    <>
      <div className={styles.filterGroup}>
        <span className={styles.filterLabel}>Type</span>
        <div className={styles.checkboxGroup}>
          {TYPE_OPTIONS.map((opt) => (
            <label key={opt.value} className={styles.checkbox}>
              <input
                type="checkbox"
                checked={selectedTypes.includes(opt.value)}
                onChange={() => toggleArrayParam('type', opt.value, selectedTypes)}
              />
              {opt.label}
            </label>
          ))}
        </div>
      </div>

      <div className={styles.filterGroup}>
        <span className={styles.filterLabel}>Level</span>
        <div className={styles.checkboxGroup}>
          {LEVEL_OPTIONS.map((opt) => (
            <label key={opt.value} className={styles.checkbox}>
              <input
                type="checkbox"
                checked={selectedLevels.includes(opt.value)}
                onChange={() => toggleArrayParam('level', opt.value, selectedLevels)}
              />
              {opt.label}
            </label>
          ))}
        </div>
      </div>

      {selectedLevels.includes('state') && (
        <div className={styles.filterGroup}>
          <span className={styles.filterLabel}>State</span>
          <select
            className={styles.stateSelect}
            value={selectedState}
            onChange={(e) => updateParams({ state: e.target.value })}
          >
            <option value="">All States</option>
            {US_STATES.map((st) => (
              <option key={st} value={st}>{st}</option>
            ))}
          </select>
        </div>
      )}

      <div className={styles.filterGroup}>
        <span className={styles.filterLabel}>Tags</span>
        <div className={styles.tagChips}>
          {POPULAR_TAGS.map((tag) => (
            <button
              key={tag}
              type="button"
              className={`${styles.tagChip} ${selectedTags.includes(tag) ? styles.tagChipActive : ''}`}
              onClick={() => handleTagClick(tag)}
            >
              {tag}
            </button>
          ))}
        </div>
      </div>
    </>
  )

  // Build active filter chips
  const activeFilterChips: { key: string; value?: string; label: string }[] = []
  if (search) activeFilterChips.push({ key: 'q', label: `"${search}"` })
  for (const t of selectedTypes) {
    const opt = TYPE_OPTIONS.find((o) => o.value === t)
    activeFilterChips.push({ key: 'type', value: t, label: opt?.label || t })
  }
  for (const l of selectedLevels) {
    const opt = LEVEL_OPTIONS.find((o) => o.value === l)
    activeFilterChips.push({ key: 'level', value: l, label: opt?.label || l })
  }
  if (selectedState) activeFilterChips.push({ key: 'state', label: selectedState })
  for (const tag of selectedTags) {
    activeFilterChips.push({ key: 'tags', value: tag, label: `#${tag}` })
  }

  return (
    <div className={styles.page}>
      <div className={styles.pageHeader}>
        <h1 className={styles.pageTitle}>Your Feed</h1>
        <p className={styles.pageSubtitle}>Policies tailored to your interests.</p>
      </div>

      {/* Search */}
      <div className={styles.searchRow}>
        <SearchBar
          placeholder="Search policies..."
          defaultValue={search}
          onChange={(e) => handleSearchChange(e.target.value)}
        />
      </div>

      {/* Sort tabs + mobile filter button */}
      <div className={styles.sortRow}>
        <TabNav
          tabs={SORT_TABS}
          activeTab={sort}
          onTabChange={handleSortChange}
        />
        {sort === 'top' && (
          <div className={styles.topSubmenu}>
            {TOP_RANGES.map((r) => (
              <button
                key={r.id}
                type="button"
                className={`${styles.topSubmenuBtn} ${topRange === r.id ? styles.topSubmenuBtnActive : ''}`}
                onClick={() => handleTopRangeChange(r.id)}
              >
                {r.label}
              </button>
            ))}
          </div>
        )}
        <button
          type="button"
          className={styles.mobileFilterBtn}
          onClick={() => setDrawerOpen(true)}
        >
          ☰ Filters{hasActiveFilters ? ` (${activeFilterChips.length})` : ''}
        </button>
      </div>

      {/* Active filter chips */}
      {activeFilterChips.length > 0 && (
        <div className={styles.activeFilters}>
          {activeFilterChips.map((chip, i) => (
            <button
              key={`${chip.key}-${chip.value || ''}-${i}`}
              type="button"
              className={styles.activeFilterChip}
              onClick={() => removeFilter(chip.key, chip.value)}
            >
              {chip.label} ✕
            </button>
          ))}
          <button type="button" className={styles.clearAll} onClick={clearAllFilters}>
            Clear all
          </button>
        </div>
      )}

      {/* Main layout */}
      <div className={styles.layout}>
        {/* Desktop sidebar */}
        <aside className={styles.sidebar}>
          <div className={styles.sidebarCard}>
            <h2 className={styles.sidebarTitle}>Filters</h2>
            {filterContent}
          </div>
        </aside>

        {/* Feed */}
        <div className={styles.feed}>
          {feedState.loading ? (
            <>
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className={styles.skeletonCard}>
                  <div className={styles.skeletonRow}>
                    <Skeleton width={90} height={20} variant="rect" />
                    <Skeleton width={72} height={20} variant="rect" />
                  </div>
                  <Skeleton width="80%" height={22} />
                  <Skeleton width="100%" height={16} />
                  <Skeleton width="60%" height={16} />
                  <div className={styles.skeletonRow}>
                    <Skeleton width={60} height={14} />
                    <Skeleton width={60} height={14} />
                    <Skeleton width={60} height={14} />
                    <Skeleton width={60} height={14} />
                  </div>
                </div>
              ))}
            </>
          ) : feedState.error ? (
            <EmptyState
              heading="Something went wrong"
              description={feedState.error}
              action={
                <Button variant="secondary" onClick={() => fetchPolicies(1, false)}>
                  Try again
                </Button>
              }
            />
          ) : feedState.policies.length === 0 ? (
            hasActiveFilters ? (
              <EmptyState
                heading="No policies match your filters"
                description="Try broadening your search or clearing some filters."
                action={
                  <Button variant="secondary" onClick={clearAllFilters}>
                    Clear all filters
                  </Button>
                }
              />
            ) : (
              <EmptyState
                heading="Be the first to add a policy!"
                description="No policies have been created yet. Start the conversation."
                action={
                  <Link to="/policies/new">
                    <Button variant="primary">Create Policy</Button>
                  </Link>
                }
              />
            )
          ) : (
            <>
              {feedState.policies.map((policy) => (
                <PolicyCard
                  key={policy.id}
                  policy={policy}
                  onTagClick={handleTagClick}
                />
              ))}
              {canLoadMore && (
                <div className={styles.loadMoreRow}>
                  {feedState.loadingMore ? (
                    <Spinner size="md" />
                  ) : (
                    <Button variant="secondary" onClick={handleLoadMore}>
                      Load more
                    </Button>
                  )}
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* Mobile filter drawer */}
      {drawerOpen && (
        <>
          <div className={styles.drawerOverlay} onClick={() => setDrawerOpen(false)} />
          <div className={styles.drawer}>
            <div className={styles.drawerHeader}>
              <h2 className={styles.drawerTitle}>Filters</h2>
              <button
                type="button"
                className={styles.drawerClose}
                onClick={() => setDrawerOpen(false)}
                aria-label="Close filters"
              >
                ✕
              </button>
            </div>
            {filterContent}
          </div>
        </>
      )}
    </div>
  )
}
