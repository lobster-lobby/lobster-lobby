import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { PolicyCard } from '../components/PolicyCard'
import type { Policy } from '../components/PolicyCard'
import { Spinner } from '../components/ui'
import styles from './PublicFeed.module.css'

export default function PublicFeed() {
  const { isAuthenticated, user } = useAuth()
  const [policies, setPolicies] = useState<Policy[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchPolicies() {
      try {
        const res = await fetch('/api/policies?sort=hot&perPage=10')
        if (!res.ok) throw new Error('Failed to load policies')
        const data = await res.json()
        setPolicies(data.policies || [])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Something went wrong')
      } finally {
        setLoading(false)
      }
    }
    fetchPolicies()
  }, [])

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Link to="/" className={styles.logo}>
          <span className={styles.logoIcon}>🦞</span>
          <span className={styles.logoText}>Lobster Lobby</span>
        </Link>
        <nav className={styles.nav}>
          {isAuthenticated ? (
            <>
              <span className={styles.navGreeting}>Hi, {user?.username}</span>
              <Link to="/policies" className={styles.navLink}>Browse</Link>
              <Link to="/dashboard" className={`${styles.navLink} ${styles.navLinkPrimary}`}>Dashboard</Link>
            </>
          ) : (
            <>
              <Link to="/login" className={styles.navLink}>Sign in</Link>
              <Link to="/register" className={`${styles.navLink} ${styles.navLinkPrimary}`}>Join free</Link>
            </>
          )}
        </nav>
      </header>

      {!isAuthenticated && (
        <section className={styles.hero}>
          <h1 className={styles.heroTitle}>Policy debate for everyone</h1>
          <p className={styles.heroSubtitle}>
            Browse trending legislation, join debates, and make your voice heard — no government ID required.
          </p>
          <div className={styles.heroCtas}>
            <Link to="/register" className={styles.ctaPrimary}>Get started free</Link>
            <Link to="/login" className={styles.ctaSecondary}>Sign in</Link>
          </div>
        </section>
      )}

      <main className={styles.main}>
        <div className={styles.feedHeader}>
          <h2 className={styles.feedTitle}>Trending Policies</h2>
          {isAuthenticated && (
            <Link to="/policies" className={styles.viewAll}>View all →</Link>
          )}
        </div>

        {loading ? (
          <div className={styles.center}>
            <Spinner size="lg" />
          </div>
        ) : error ? (
          <div className={styles.emptyState}>
            <p className={styles.emptyIcon}>⚠️</p>
            <h3 className={styles.emptyTitle}>Could not load policies</h3>
            <p className={styles.emptyBody}>{error}</p>
            <button
              className={styles.ctaPrimary}
              onClick={() => window.location.reload()}
            >
              Try again
            </button>
          </div>
        ) : policies.length === 0 ? (
          <div className={styles.emptyState}>
            <p className={styles.emptyIcon}>🗳️</p>
            <h3 className={styles.emptyTitle}>No policies yet — be the first!</h3>
            <p className={styles.emptyBody}>
              Lobster Lobby is just getting started. Create the first policy proposal and kick off the debate.
            </p>
            {isAuthenticated ? (
              <Link to="/policies/new" className={styles.ctaPrimary}>Create a policy</Link>
            ) : (
              <Link to="/register" className={styles.ctaPrimary}>Join and create a policy</Link>
            )}
          </div>
        ) : (
          <div className={styles.feed}>
            {policies.map((policy) => (
              <PolicyCard key={policy.id} policy={policy} />
            ))}
            <div className={styles.feedFooter}>
              {isAuthenticated ? (
                <Link to="/policies" className={styles.ctaSecondary}>Browse all policies →</Link>
              ) : (
                <div className={styles.joinBanner}>
                  <p className={styles.joinBannerText}>Want to debate, vote, and track your representatives?</p>
                  <Link to="/register" className={styles.ctaPrimary}>Join Lobster Lobby free</Link>
                </div>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
