import { useState, type FormEvent } from 'react'
import type { CivicOfficial } from '../types/representative'
import styles from './Representatives.module.css'

export default function Representatives() {
  const [address, setAddress] = useState('')
  const [officials, setOfficials] = useState<CivicOfficial[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasSearched, setHasSearched] = useState(false)
  const [photoErrors, setPhotoErrors] = useState<Record<string, boolean>>({})

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!address.trim()) return

    setLoading(true)
    setError(null)
    setHasSearched(true)

    try {
      const params = new URLSearchParams({ address: address.trim() })
      const res = await fetch(`/api/representatives?${params}`)
      if (!res.ok) {
        throw new Error('Failed to look up representatives')
      }
      const data = await res.json()
      setOfficials(data.officials || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
      setOfficials([])
    } finally {
      setLoading(false)
    }
  }

  const getPartyClass = (party: string) => {
    const p = party?.toLowerCase() || ''
    if (p.includes('democrat')) return styles.partyDemocratic
    if (p.includes('republican')) return styles.partyRepublican
    return styles.partyOther
  }

  const getPartyAbbrev = (party: string) => {
    const p = party?.toLowerCase() || ''
    if (p.includes('democrat')) return 'D'
    if (p.includes('republican')) return 'R'
    if (p.includes('independent')) return 'I'
    if (p.includes('libertarian')) return 'L'
    if (p.includes('green')) return 'G'
    return party?.charAt(0) || '?'
  }

  const getSocialUrl = (type: string, id: string) => {
    switch (type.toLowerCase()) {
      case 'twitter':
        return `https://twitter.com/${id}`
      case 'facebook':
        return `https://facebook.com/${id}`
      case 'youtube':
        return `https://youtube.com/${id}`
      case 'instagram':
        return `https://instagram.com/${id}`
      default:
        return null
    }
  }

  const getSocialIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case 'twitter':
        return 'X'
      case 'facebook':
        return 'f'
      case 'youtube':
        return 'YT'
      case 'instagram':
        return 'IG'
      default:
        return type.charAt(0).toUpperCase()
    }
  }

  return (
    <div className={styles.page}>
      <header className={styles.pageHeader}>
        <h1 className={styles.pageTitle}>Find Your Representatives</h1>
        <p className={styles.pageSubtitle}>
          Enter your address to see who represents you at every level of government.
        </p>
      </header>

      <form className={styles.searchForm} onSubmit={handleSubmit}>
        <div className={styles.searchRow}>
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Enter your full address (e.g., 123 Main St, City, State ZIP)"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            disabled={loading}
          />
          <button type="submit" className={styles.searchBtn} disabled={loading || !address.trim()}>
            {loading ? 'Searching...' : 'Search'}
          </button>
        </div>
      </form>

      {loading && (
        <div className={styles.loading}>
          <div className={styles.spinner} />
        </div>
      )}

      {error && <div className={styles.error}>{error}</div>}

      {!loading && !error && hasSearched && officials.length === 0 && (
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>🔍</div>
          <h2 className={styles.emptyTitle}>No representatives found</h2>
          <p className={styles.emptyText}>
            Try entering a more complete address with street, city, state, and ZIP code.
          </p>
        </div>
      )}

      {!loading && officials.length > 0 && (
        <div className={styles.results}>
          {officials.map((official, idx) => (
            <article key={`${official.name}-${idx}`} className={styles.officialCard}>
              <div className={styles.officialHeader}>
                {official.photoUrl && !photoErrors[`${official.name}-${idx}`] ? (
                  <img
                    src={official.photoUrl}
                    alt={official.name}
                    className={styles.officialPhoto}
                    onError={() => setPhotoErrors(prev => ({ ...prev, [`${official.name}-${idx}`]: true }))}
                  />
                ) : (
                  <div className={styles.officialAvatar}>
                    {official.name.charAt(0)}
                  </div>
                )}
                <div className={styles.officialInfo}>
                  <h3 className={styles.officialName}>{official.name}</h3>
                  <p className={styles.officialTitle}>{official.title}</p>
                  {official.party && (
                    <span className={`${styles.officialParty} ${getPartyClass(official.party)}`}>
                      {getPartyAbbrev(official.party)}
                    </span>
                  )}
                </div>
              </div>

              <div className={styles.contactList}>
                {official.phone && (
                  <div className={styles.contactItem}>
                    <span className={styles.contactIcon}>T</span>
                    <a href={`tel:${official.phone}`}>{official.phone}</a>
                  </div>
                )}
                {official.email && (
                  <div className={styles.contactItem}>
                    <span className={styles.contactIcon}>@</span>
                    <a href={`mailto:${official.email}`}>{official.email}</a>
                  </div>
                )}
                {official.urls && official.urls.length > 0 && (
                  <div className={styles.contactItem}>
                    <span className={styles.contactIcon}>W</span>
                    <a href={official.urls[0]} target="_blank" rel="noopener noreferrer">
                      Website
                    </a>
                  </div>
                )}
              </div>

              {official.socialMedia && Object.keys(official.socialMedia).length > 0 && (
                <div className={styles.socialLinks}>
                  {Object.entries(official.socialMedia).map(([type, id]) => {
                    const url = getSocialUrl(type, id)
                    if (!url) return null
                    return (
                      <a
                        key={type}
                        href={url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.socialLink}
                        title={type}
                      >
                        {getSocialIcon(type)}
                      </a>
                    )
                  })}
                </div>
              )}
            </article>
          ))}
        </div>
      )}

      {!hasSearched && (
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>🏛️</div>
          <h2 className={styles.emptyTitle}>Look up your elected officials</h2>
          <p className={styles.emptyText}>
            Enter your address above to find your representatives at the federal, state, and local levels.
          </p>
        </div>
      )}
    </div>
  )
}
