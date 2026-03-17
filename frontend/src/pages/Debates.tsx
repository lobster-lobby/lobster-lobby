import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Card, Spinner, EmptyState } from '../components/ui'
import type { Debate } from '../types/debates'
import { relativeTime } from '../utils/time'
import styles from './Debates.module.css'

export default function Debates() {
  const [debates, setDebates] = useState<Debate[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchDebates() {
      try {
        const res = await fetch('/api/debates')
        if (!res.ok) return
        const data = await res.json()
        setDebates(data.debates)
      } catch {
        // Silently fail
      } finally {
        setLoading(false)
      }
    }
    fetchDebates()
  }, [])

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.loading}><Spinner size="lg" /></div>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <h1 className={styles.title}>Debates</h1>
      </header>

      {debates.length === 0 ? (
        <EmptyState heading="No debates yet" description="Start a debate to get the conversation going." />
      ) : (
        <div className={styles.list}>
          {debates.map((d) => (
            <Link key={d.id} to={`/debates/${d.slug}`} className={styles.link}>
              <Card>
                <div className={styles.debateCard}>
                  <h3 className={styles.debateTitle}>{d.title}</h3>
                  {d.description && (
                    <p className={styles.debateDesc}>
                      {d.description.length > 150 ? d.description.slice(0, 150) + '...' : d.description}
                    </p>
                  )}
                  <div className={styles.debateMeta}>
                    <span>by {d.creatorUsername}</span>
                    <span>{relativeTime(d.createdAt)}</span>
                    <span>{d.argumentCount} arguments</span>
                  </div>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
