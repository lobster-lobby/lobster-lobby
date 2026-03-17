import { useState, useEffect, useCallback } from 'react'
import { CrossReferences } from '../components/CrossReferences'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { Card, Spinner, EmptyState } from '../components/ui'
import DebateArgument from '../components/debates/DebateArgument'
import ArgumentComposer from '../components/debates/ArgumentComposer'
import SortControls from '../components/debates/SortControls'
import type { Debate, Argument, DebateSortOption } from '../types/debates'
import { getAccessToken } from '../hooks/useAuth'
import { relativeTime } from '../utils/time'
import styles from './DebateDetail.module.css'

export default function DebateDetail() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()

  const [debate, setDebate] = useState<Debate | null>(null)
  const [arguments_, setArguments] = useState<Argument[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const sortParam = searchParams.get('sort') as DebateSortOption | null
  const sort: DebateSortOption = sortParam && ['newest', 'top', 'controversial'].includes(sortParam)
    ? sortParam
    : 'top'

  const fetchDebate = useCallback(async () => {
    if (!slug) return
    setLoading(true)
    setError(null)

    try {
      const token = getAccessToken()
      const headers: HeadersInit = {}
      if (token) headers['Authorization'] = `Bearer ${token}`

      const params = new URLSearchParams()
      params.set('sort', sort)

      const res = await fetch(`/api/debates/${slug}?${params.toString()}`, { headers })
      if (!res.ok) {
        if (res.status === 404) {
          setError('Debate not found')
        } else {
          setError('Failed to load debate')
        }
        return
      }

      const data = await res.json()
      setDebate(data.debate)
      setArguments(data.arguments)
    } catch {
      setError('Something went wrong')
    } finally {
      setLoading(false)
    }
  }, [slug, sort])

  useEffect(() => {
    fetchDebate()
  }, [fetchDebate])

  function handleSortChange(newSort: DebateSortOption) {
    if (newSort === 'top') {
      searchParams.delete('sort')
    } else {
      searchParams.set('sort', newSort)
    }
    setSearchParams(searchParams, { replace: true })
  }

  function handleArgumentCreated(arg: Argument) {
    setArguments((prev) => [arg, ...prev])
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.loading}><Spinner size="lg" /></div>
      </div>
    )
  }

  if (error || !debate) {
    return (
      <div className={styles.container}>
        <Card>
          <div className={styles.errorState}>
            <h2>{error === 'Debate not found' ? '404' : 'Error'}</h2>
            <p>{error || 'Debate not found'}</p>
            <button onClick={() => navigate('/debates')} className={styles.backBtn}>Back to Debates</button>
          </div>
        </Card>
      </div>
    )
  }

  const proArgs = arguments_.filter((a) => a.side === 'pro')
  const conArgs = arguments_.filter((a) => a.side === 'con')

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <h1 className={styles.title}>{debate.title}</h1>
        {debate.description && <p className={styles.description}>{debate.description}</p>}
        <div className={styles.meta}>
          <span>by {debate.creatorUsername}</span>
          <span>{relativeTime(debate.createdAt)}</span>
          <span>{debate.argumentCount} arguments</span>
          <span className={styles.status}>{debate.status}</span>
        </div>
      </header>

      <ArgumentComposer debateSlug={debate.slug} onArgumentCreated={handleArgumentCreated} />

      <div className={styles.controls}>
        <SortControls current={sort} onChange={handleSortChange} />
      </div>

      {arguments_.length === 0 ? (
        <EmptyState heading="No arguments yet" description="Be the first to make an argument in this debate." />
      ) : (
        <div className={styles.columns}>
          <div className={styles.column}>
            <h3 className={styles.columnTitle} style={{ color: 'var(--ll-support)' }}>Pro ({proArgs.length})</h3>
            {proArgs.length === 0 ? (
              <p className={styles.emptyColumn}>No pro arguments yet</p>
            ) : proArgs.map((a) => (
              <DebateArgument key={a.id} argument={a} debateSlug={debate.slug} />
            ))}
          </div>
          <div className={styles.column}>
            <h3 className={styles.columnTitle} style={{ color: 'var(--ll-oppose)' }}>Con ({conArgs.length})</h3>
            {conArgs.length === 0 ? (
              <p className={styles.emptyColumn}>No con arguments yet</p>
            ) : conArgs.map((a) => (
              <DebateArgument key={a.id} argument={a} debateSlug={debate.slug} />
            ))}
          </div>
        </div>
      )}
      {debate && <CrossReferences entityType="debate" entityId={debate.id} />}
    </div>
  )
}
