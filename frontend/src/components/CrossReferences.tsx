import { useEffect, useState, useCallback } from 'react'
import { getAccessToken } from '../hooks/useAuth'

export interface CrossReference {
  id: string
  sourceType: string
  sourceId: string
  sourceTitle: string
  targetType: string
  targetId: string
  targetTitle: string
  createdBy: string
  createdAt: string
}

interface Props {
  entityType: 'research' | 'debate' | 'policy'
  entityId: string
}

const TYPE_LABELS: Record<string, string> = {
  research: 'Research',
  debate: 'Debate',
  policy: 'Policy',
}

const TYPE_COLORS: Record<string, string> = {
  research: '#2563eb',
  debate: '#7c3aed',
  policy: '#059669',
}

export function CrossReferences({ entityType, entityId }: Props) {
  const token = getAccessToken()
  const [refs, setRefs] = useState<CrossReference[]>([])
  const [loading, setLoading] = useState(true)
  const [showSearch, setShowSearch] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<Array<{ id: string; title: string; type: string }>>([])
  const [searching, setSearching] = useState(false)

  const fetchRefs = useCallback(async () => {
    try {
      const res = await fetch(`/api/cross-references?type=${entityType}&id=${entityId}`)
      if (res.ok) {
        const data = await res.json()
        setRefs(data.references || [])
      }
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }, [entityType, entityId])

  useEffect(() => {
    fetchRefs()
  }, [fetchRefs])

  const handleDelete = async (refId: string) => {
    if (!token) return
    try {
      await fetch(`/api/cross-references/${refId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      })
      setRefs(prev => prev.filter(r => r.id !== refId))
    } catch {
      // ignore
    }
  }

  const handleSearch = async (q: string) => {
    setSearchQuery(q)
    if (q.trim().length < 2) {
      setSearchResults([])
      return
    }
    setSearching(true)
    try {
      const res = await fetch(`/api/search?q=${encodeURIComponent(q)}&types=research,debate,policy`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
      if (res.ok) {
        const data = await res.json()
        const results: Array<{ id: string; title: string; type: string }> = []
        ;(data.policies || []).forEach((p: { id: string; title: string }) =>
          results.push({ id: p.id, title: p.title, type: 'policy' })
        )
        ;(data.research || []).forEach((r: { id: string; title: string }) =>
          results.push({ id: r.id, title: r.title, type: 'research' })
        )
        ;(data.debates || []).forEach((d: { id: string; title: string }) =>
          results.push({ id: d.id, title: d.title, type: 'debate' })
        )
        setSearchResults(results.filter(r => !(r.id === entityId && r.type === entityType)))
      }
    } catch {
      // ignore
    } finally {
      setSearching(false)
    }
  }

  const handleAddRef = async (target: { id: string; type: string; title: string }) => {
    if (!token) return
    try {
      const res = await fetch('/api/cross-references', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({
          sourceType: entityType,
          sourceId: entityId,
          targetType: target.type,
          targetId: target.id,
        }),
      })
      if (res.ok) {
        setShowSearch(false)
        setSearchQuery('')
        setSearchResults([])
        fetchRefs()
      } else if (res.status === 409) {
        alert('This reference already exists.')
      }
    } catch {
      // ignore
    }
  }

  const getLinkedItem = (ref: CrossReference) => {
    if (ref.sourceType === entityType && ref.sourceId === entityId) {
      return { id: ref.targetId, type: ref.targetType, title: ref.targetTitle }
    }
    return { id: ref.sourceId, type: ref.sourceType, title: ref.sourceTitle }
  }

  if (loading) return null

  return (
    <div style={{ marginTop: '2rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
        <h3 style={{ margin: 0, fontSize: '1.1rem' }}>Related</h3>
        {token && (
          <button
            onClick={() => setShowSearch(!showSearch)}
            style={{ fontSize: '0.85rem', padding: '0.25rem 0.75rem', cursor: 'pointer' }}
          >
            + Add Reference
          </button>
        )}
      </div>

      {showSearch && (
        <div style={{ marginBottom: '1rem', padding: '0.75rem', background: '#f9fafb', borderRadius: '0.5rem', border: '1px solid #e5e7eb' }}>
          <input
            type="text"
            placeholder="Search by title..."
            value={searchQuery}
            onChange={e => handleSearch(e.target.value)}
            style={{ width: '100%', padding: '0.5rem', border: '1px solid #d1d5db', borderRadius: '0.375rem', fontSize: '0.9rem' }}
            autoFocus
          />
          {searching && <p style={{ margin: '0.5rem 0 0', fontSize: '0.85rem', color: '#6b7280' }}>Searching…</p>}
          {searchResults.length > 0 && (
            <ul style={{ margin: '0.5rem 0 0', padding: 0, listStyle: 'none', maxHeight: '200px', overflowY: 'auto' }}>
              {searchResults.map(r => (
                <li key={`${r.type}-${r.id}`} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.4rem 0', borderBottom: '1px solid #e5e7eb' }}>
                  <span>
                    <span style={{ fontSize: '0.75rem', background: TYPE_COLORS[r.type], color: '#fff', borderRadius: '0.25rem', padding: '0.1rem 0.4rem', marginRight: '0.5rem' }}>
                      {TYPE_LABELS[r.type]}
                    </span>
                    {r.title}
                  </span>
                  <button onClick={() => handleAddRef(r)} style={{ fontSize: '0.8rem', cursor: 'pointer', padding: '0.2rem 0.5rem' }}>
                    Add
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {refs.length === 0 ? (
        <p style={{ color: '#9ca3af', fontSize: '0.9rem' }}>No related items yet.</p>
      ) : (
        <ul style={{ padding: 0, listStyle: 'none', margin: 0 }}>
          {refs.map(ref => {
            const item = getLinkedItem(ref)
            return (
              <li key={ref.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid #e5e7eb' }}>
                <span>
                  <span style={{ fontSize: '0.75rem', background: TYPE_COLORS[item.type], color: '#fff', borderRadius: '0.25rem', padding: '0.1rem 0.4rem', marginRight: '0.5rem' }}>
                    {TYPE_LABELS[item.type]}
                  </span>
                  {item.title || `${item.type}/${item.id}`}
                </span>
                {token && (
                  <button
                    onClick={() => handleDelete(ref.id)}
                    title="Remove reference"
                    style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#9ca3af', fontSize: '1rem', lineHeight: 1 }}
                  >
                    ×
                  </button>
                )}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
