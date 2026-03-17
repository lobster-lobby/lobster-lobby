import { useState, useEffect, useCallback } from 'react'
import { useAuth } from '../../hooks/useAuth'
import { Card, Button, Spinner, EmptyState } from '../ui'
import { ResearchCard } from '../research/ResearchCard'
import { ResearchSubmit } from '../research/ResearchSubmit'
import type { ResearchResponse, ResearchListResponse } from '../../types/research'
import styles from './ResearchTab.module.css'

interface ResearchTabProps {
  policyId: string
}

const typeFilters: { id: string; label: string }[] = [
  { id: '', label: 'All' },
  { id: 'analysis', label: 'Analysis' },
  { id: 'news', label: 'News' },
  { id: 'data', label: 'Data' },
  { id: 'academic', label: 'Academic' },
  { id: 'government', label: 'Government' },
]

const sortOptions = [
  { value: 'newest', label: 'Newest' },
  { value: 'top', label: 'Top Rated' },
  { value: 'most_cited', label: 'Most Cited' },
]

const ITEMS_PER_PAGE = 20

export default function ResearchTab({ policyId }: ResearchTabProps) {
  const { isAuthenticated } = useAuth()
  const [research, setResearch] = useState<ResearchResponse[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [sort, setSort] = useState('newest')
  const [typeFilter, setTypeFilter] = useState('')
  const [loading, setLoading] = useState(true)
  const [showSubmitForm, setShowSubmitForm] = useState(false)

  const fetchResearch = useCallback(async () => {
    setLoading(true)
    try {
      const token = localStorage.getItem('ll_token')
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`

      const params = new URLSearchParams()
      params.set('sort', sort)
      params.set('page', String(page))
      params.set('limit', String(ITEMS_PER_PAGE))
      if (typeFilter) params.set('type', typeFilter)

      const res = await fetch(`/api/policies/${policyId}/research?${params.toString()}`, { headers })
      if (!res.ok) throw new Error('Failed to fetch research')
      const data: ResearchListResponse = await res.json()
      setResearch(data.research || [])
      setTotal(data.total)
    } catch {
      // Silently fail
    } finally {
      setLoading(false)
    }
  }, [policyId, sort, typeFilter, page])

  useEffect(() => {
    fetchResearch()
  }, [fetchResearch])

  const handleVote = async (id: string, value: number) => {
    if (!isAuthenticated) return

    // Optimistic update
    setResearch((prev) =>
      prev.map((r) => {
        if (r.id !== id) return r
        const oldVote = r.userVote
        let upvotes = r.upvotes
        let downvotes = r.downvotes

        // Remove old vote
        if (oldVote === 1) upvotes--
        else if (oldVote === -1) downvotes--

        // Add new vote
        if (value === 1) upvotes++
        else if (value === -1) downvotes++

        return {
          ...r,
          upvotes,
          downvotes,
          score: upvotes - downvotes,
          userVote: value,
        }
      })
    )

    try {
      const token = localStorage.getItem('ll_token')
      await fetch(`/api/policies/${policyId}/research/${id}/react`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ value }),
      })
    } catch {
      // Revert on error - refetch
      fetchResearch()
    }
  }

  const handleResearchSubmitted = (newResearch: ResearchResponse) => {
    setResearch((prev) => [newResearch, ...prev])
    setTotal((t) => t + 1)
    setShowSubmitForm(false)
  }

  const totalPages = Math.ceil(total / ITEMS_PER_PAGE)

  return (
    <div className={styles.wrapper}>
      <Card>
        <div className={styles.header}>
          <h3 className={styles.title}>Research ({total})</h3>
          {isAuthenticated && !showSubmitForm && (
            <Button variant="primary" size="sm" onClick={() => setShowSubmitForm(true)}>
              Submit Research
            </Button>
          )}
        </div>
      </Card>

      {showSubmitForm && (
        <ResearchSubmit
          policyId={policyId}
          onSubmitted={handleResearchSubmitted}
          onCancel={() => setShowSubmitForm(false)}
        />
      )}

      <div className={styles.controls}>
        <div className={styles.filters}>
          {typeFilters.map((f) => (
            <button
              key={f.id}
              className={[styles.filterBtn, typeFilter === f.id && styles.activeFilter]
                .filter(Boolean)
                .join(' ')}
              onClick={() => {
                setTypeFilter(f.id)
                setPage(1)
              }}
              type="button"
            >
              {f.label}
            </button>
          ))}
        </div>
        <select
          className={styles.sortSelect}
          value={sort}
          onChange={(e) => {
            setSort(e.target.value)
            setPage(1)
          }}
        >
          {sortOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {loading ? (
        <div className={styles.loading}>
          <Spinner size="lg" />
        </div>
      ) : research.length === 0 ? (
        <EmptyState
          heading="No research yet"
          description="Be the first to contribute research and analysis on this policy."
        />
      ) : (
        <>
          <div className={styles.list}>
            {research.map((r) => (
              <ResearchCard
                key={r.id}
                research={r}
                onVote={handleVote}
              />
            ))}
          </div>

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <button
                className={styles.pageBtn}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                type="button"
              >
                Previous
              </button>
              <span className={styles.pageInfo}>
                Page {page} of {totalPages}
              </span>
              <button
                className={styles.pageBtn}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                type="button"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
