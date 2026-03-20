import { useState, useEffect, useCallback } from 'react'
import { useAuth, getAccessToken } from '../../hooks/useAuth'
import { Button, Spinner, EmptyState, TabNav } from '../ui'
import DraftCard from '../drafts/DraftCard'
import DraftForm from '../drafts/DraftForm'
import type { Draft, DraftCategory } from '../../types/draft'
import styles from './DraftTab.module.css'

interface DraftTabProps {
  policyId: string
}

const sortTabs = [
  { id: 'endorsements', label: 'Top Endorsed' },
  { id: 'newest', label: 'Newest' },
]

const CATEGORIES: { value: DraftCategory | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'amendment', label: 'Amendment' },
  { value: 'talking-point', label: 'Talking Point' },
  { value: 'position-statement', label: 'Position Statement' },
  { value: 'full-text', label: 'Full Text' },
]

export default function DraftTab({ policyId }: DraftTabProps) {
  const { isAuthenticated } = useAuth()
  const [drafts, setDrafts] = useState<Draft[]>([])
  const [loading, setLoading] = useState(true)
  const [sort, setSort] = useState('endorsements')
  const [categoryFilter, setCategoryFilter] = useState<DraftCategory | 'all'>('all')
  const [showForm, setShowForm] = useState(false)

  const fetchDrafts = useCallback(async () => {
    setLoading(true)
    try {
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`
      const params = new URLSearchParams({ sort })
      if (categoryFilter !== 'all') params.set('category', categoryFilter)
      const res = await fetch(`/api/policies/${policyId}/drafts?${params.toString()}`, { headers })
      if (!res.ok) throw new Error('Failed to fetch drafts')
      const data = await res.json()
      setDrafts(Array.isArray(data) ? data : data.drafts ?? [])
    } catch {
      // silently fail
    } finally {
      setLoading(false)
    }
  }, [policyId, sort, categoryFilter])

  useEffect(() => {
    fetchDrafts()
  }, [fetchDrafts])

  function handleCreated(draft: Draft) {
    setDrafts((prev) => [draft, ...prev])
    setShowForm(false)
  }

  function handleUpdated(updated: Draft) {
    setDrafts((prev) => prev.map((d) => (d.id === updated.id ? updated : d)))
  }

  function handleDeleted(draftId: string) {
    setDrafts((prev) => prev.filter((d) => d.id !== draftId))
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.topBar}>
        <h3 className={styles.title}>Drafts ({drafts.length})</h3>
        {isAuthenticated && !showForm && (
          <Button size="sm" onClick={() => setShowForm(true)}>+ New Draft</Button>
        )}
      </div>

      {showForm && (
        <div className={styles.formSection}>
          <DraftForm
            policyId={policyId}
            onSaved={handleCreated}
            onCancel={() => setShowForm(false)}
          />
        </div>
      )}

      <div className={styles.controls}>
        <TabNav tabs={sortTabs} activeTab={sort} onTabChange={setSort} />
        <div className={styles.filters}>
          {CATEGORIES.map((c) => (
            <button
              key={c.value}
              type="button"
              className={[styles.filterPill, categoryFilter === c.value && styles.activeFilter].filter(Boolean).join(' ')}
              onClick={() => setCategoryFilter(c.value)}
            >
              {c.label}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className={styles.loading}><Spinner size="lg" /></div>
      ) : drafts.length === 0 ? (
        <EmptyState
          heading="No drafts yet"
          description="Be the first to propose draft language for this policy."
        />
      ) : (
        <div className={styles.list}>
          {drafts.map((draft) => (
            <DraftCard
              key={draft.id}
              draft={draft}
              onUpdated={handleUpdated}
              onDeleted={handleDeleted}
            />
          ))}
        </div>
      )}
    </div>
  )
}
