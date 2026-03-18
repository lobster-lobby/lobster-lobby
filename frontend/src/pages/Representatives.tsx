import { useState, useEffect, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import type { Representative, CivicOfficial } from '../types/representative'
import { Pagination } from '../components/ui'
import styles from './Representatives.module.css'

const US_STATES = [
  'AL','AK','AZ','AR','CA','CO','CT','DE','FL','GA','HI','ID','IL','IN','IA',
  'KS','KY','LA','ME','MD','MA','MI','MN','MS','MO','MT','NE','NV','NH','NJ',
  'NM','NY','NC','ND','OH','OK','OR','PA','RI','SC','SD','TN','TX','UT','VT',
  'VA','WA','WV','WI','WY','DC',
]

export default function Representatives() {
  const [reps, setReps] = useState<Representative[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [perPage] = useState(20)
  const [loading, setLoading] = useState(true)

  // Filters
  const [search, setSearch] = useState('')
  const [party, setParty] = useState('')
  const [state, setState] = useState('')
  const [chamber, setChamber] = useState('')

  // Address lookup
  const [address, setAddress] = useState('')
  const [officials, setOfficials] = useState<CivicOfficial[]>([])
  const [addressLoading, setAddressLoading] = useState(false)
  const [addressError, setAddressError] = useState<string | null>(null)
  const [hasSearchedAddress, setHasSearchedAddress] = useState(false)
  const [photoErrors, setPhotoErrors] = useState<Record<string, boolean>>({})

  useEffect(() => {
    fetchReps()
  }, [page, party, state, chamber])

  const fetchReps = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('perPage', String(perPage))
      if (search) params.set('search', search)
      if (party) params.set('party', party)
      if (state) params.set('state', state)
      if (chamber) params.set('chamber', chamber)

      const res = await fetch(`/api/representatives?${params}`)
      if (!res.ok) throw new Error('Failed to load representatives')
      const data = await res.json()
      setReps(data.representatives || [])
      setTotal(data.total || 0)
    } catch {
      setReps([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = (e: FormEvent) => {
    e.preventDefault()
    setPage(1)
    fetchReps()
  }

  const handleAddressLookup = async (e: FormEvent) => {
    e.preventDefault()
    if (!address.trim()) return
    setAddressLoading(true)
    setAddressError(null)
    setHasSearchedAddress(true)
    try {
      const params = new URLSearchParams({ address: address.trim() })
      const res = await fetch(`/api/representatives?${params}`)
      if (!res.ok) throw new Error('Failed to look up representatives')
      const data = await res.json()
      setOfficials(data.officials || [])
    } catch (err) {
      setAddressError(err instanceof Error ? err.message : 'An error occurred')
      setOfficials([])
    } finally {
      setAddressLoading(false)
    }
  }

  const getPartyClass = (p: string) => {
    const lower = p?.toLowerCase() || ''
    if (lower.includes('democrat')) return styles.partyDemocratic
    if (lower.includes('republican')) return styles.partyRepublican
    return styles.partyOther
  }

  const getPartyAbbrev = (p: string) => {
    const lower = p?.toLowerCase() || ''
    if (lower.includes('democrat')) return 'D'
    if (lower.includes('republican')) return 'R'
    if (lower.includes('independent')) return 'I'
    if (lower.includes('libertarian')) return 'L'
    return p?.charAt(0) || '?'
  }

  const getChamberLabel = (c: string) => {
    switch (c) {
      case 'senate': return 'Senate'
      case 'house': return 'House'
      case 'governor': return 'Governor'
      case 'local': return 'Local'
      default: return c
    }
  }

  const totalPages = Math.ceil(total / perPage)

  return (
    <div className={styles.page}>
      <header className={styles.pageHeader}>
        <h1 className={styles.pageTitle}>Representatives</h1>
        <p className={styles.pageSubtitle}>
          Browse elected officials and their voting records.
        </p>
      </header>

      {/* Search & Filters */}
      <div className={styles.filters}>
        <form className={styles.searchForm} onSubmit={handleSearch}>
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Search by name..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button type="submit" className={styles.searchBtn}>Search</button>
        </form>

        <div className={styles.filterRow}>
          <select
            className={styles.filterSelect}
            value={party}
            onChange={(e) => { setParty(e.target.value); setPage(1) }}
          >
            <option value="">All Parties</option>
            <option value="Democratic">Democratic</option>
            <option value="Republican">Republican</option>
            <option value="Independent">Independent</option>
          </select>

          <select
            className={styles.filterSelect}
            value={state}
            onChange={(e) => { setState(e.target.value); setPage(1) }}
          >
            <option value="">All States</option>
            {US_STATES.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>

          <select
            className={styles.filterSelect}
            value={chamber}
            onChange={(e) => { setChamber(e.target.value); setPage(1) }}
          >
            <option value="">All Chambers</option>
            <option value="senate">Senate</option>
            <option value="house">House</option>
            <option value="governor">Governor</option>
          </select>
        </div>
      </div>

      {/* Representatives List */}
      {loading ? (
        <div className={styles.loading}><div className={styles.spinner} /></div>
      ) : reps.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyIcon}>🏛️</div>
          <h2 className={styles.emptyTitle}>No representatives found</h2>
          <p className={styles.emptyText}>Try adjusting your search or filters.</p>
        </div>
      ) : (
        <>
          <div className={styles.results}>
            {reps.map((rep) => (
              <Link key={rep.id} to={`/representatives/${rep.id}`} className={styles.repCard}>
                <div className={styles.repHeader}>
                  {rep.photoUrl ? (
                    <img src={rep.photoUrl} alt={rep.name} className={styles.repPhoto} loading="lazy" />
                  ) : (
                    <div className={styles.repAvatar}>{rep.name.charAt(0)}</div>
                  )}
                  <div className={styles.repInfo}>
                    <h3 className={styles.repName}>{rep.name}</h3>
                    <p className={styles.repTitle}>{rep.title || getChamberLabel(rep.chamber)}</p>
                    <div className={styles.repMeta}>
                      <span className={`${styles.repParty} ${getPartyClass(rep.party)}`}>
                        {getPartyAbbrev(rep.party)} - {rep.party}
                      </span>
                      <span className={styles.repState}>{rep.state}{rep.district ? ` - ${rep.district}` : ''}</span>
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
          {totalPages > 1 && (
            <Pagination currentPage={page} totalPages={totalPages} onPageChange={setPage} />
          )}
        </>
      )}

      {/* Address Lookup Section */}
      <section className={styles.addressSection}>
        <h2 className={styles.addressTitle}>Find by Address</h2>
        <p className={styles.addressSubtitle}>
          Enter your address to find representatives at every level of government.
        </p>
        <form className={styles.searchForm} onSubmit={handleAddressLookup}>
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Enter your full address (e.g., 123 Main St, City, State ZIP)"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            disabled={addressLoading}
          />
          <button type="submit" className={styles.searchBtn} disabled={addressLoading || !address.trim()}>
            {addressLoading ? 'Searching...' : 'Lookup'}
          </button>
        </form>

        {addressError && <div className={styles.error}>{addressError}</div>}

        {!addressLoading && hasSearchedAddress && officials.length === 0 && !addressError && (
          <div className={styles.emptyState}>
            <p className={styles.emptyText}>No representatives found for that address.</p>
          </div>
        )}

        {officials.length > 0 && (
          <div className={styles.results}>
            {officials.map((official, idx) => (
              <article key={`${official.name}-${idx}`} className={styles.repCard}>
                <div className={styles.repHeader}>
                  {official.photoUrl && !photoErrors[`${official.name}-${idx}`] ? (
                    <img
                      src={official.photoUrl}
                      alt={official.name}
                      className={styles.repPhoto}
                      onError={() => setPhotoErrors(prev => ({ ...prev, [`${official.name}-${idx}`]: true }))}
                    />
                  ) : (
                    <div className={styles.repAvatar}>{official.name.charAt(0)}</div>
                  )}
                  <div className={styles.repInfo}>
                    <h3 className={styles.repName}>{official.name}</h3>
                    <p className={styles.repTitle}>{official.title}</p>
                    {official.party && (
                      <span className={`${styles.repParty} ${getPartyClass(official.party)}`}>
                        {getPartyAbbrev(official.party)} - {official.party}
                      </span>
                    )}
                  </div>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
