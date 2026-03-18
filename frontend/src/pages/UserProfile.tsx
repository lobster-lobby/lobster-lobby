import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useAuth, getAccessToken } from '../hooks/useAuth'
import { Button, Spinner } from '../components/ui'
import styles from './UserProfile.module.css'

interface UserProfile {
  id: string
  username: string
  displayName: string
  bio: string
  type: string
  role: string
  reputation: {
    score: number
    contributions: number
    tier: string
  }
  createdAt: string
  stats: {
    policiesCreated: number
    debateComments: number
    researchSubmitted: number
    bookmarks: number
  }
  isOwnProfile: boolean
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'long',
    year: 'numeric',
  })
}

function getInitials(username: string, displayName?: string): string {
  const name = displayName || username
  const parts = name.split(/[\s_-]+/)
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
}

function getTierColor(tier: string): string {
  switch (tier) {
    case 'gold':
      return 'var(--ll-accent)'
    case 'silver':
      return 'var(--ll-text-muted)'
    case 'bronze':
      return '#CD7F32'
    default:
      return 'var(--ll-text-secondary)'
  }
}

export default function UserProfile() {
  const { username } = useParams<{ username: string }>()
  const { user: currentUser } = useAuth()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchProfile() {
      if (!username) return

      setLoading(true)
      setError(null)

      try {
        const token = getAccessToken()
        const headers: HeadersInit = {}
        if (token) {
          headers['Authorization'] = `Bearer ${token}`
        }

        const res = await fetch(`/api/users/${username}`, { headers })
        if (!res.ok) {
          if (res.status === 404) {
            setError('User not found')
          } else {
            setError('Failed to load profile')
          }
          return
        }

        const data = await res.json()
        setProfile(data)
      } catch {
        setError('Failed to load profile')
      } finally {
        setLoading(false)
      }
    }

    fetchProfile()
  }, [username])

  if (loading) {
    return (
      <div className={styles.loadingContainer}>
        <Spinner />
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.errorContainer}>
        <h2 className={styles.errorTitle}>{error}</h2>
        <p className={styles.errorText}>
          {error === 'User not found'
            ? 'The user you are looking for does not exist.'
            : 'There was an error loading this profile. Please try again later.'}
        </p>
        <Link to="/">
          <Button variant="secondary">Go Home</Button>
        </Link>
      </div>
    )
  }

  if (!profile) return null

  const isOwnProfile = currentUser?.username === profile.username

  return (
    <div className={styles.container}>
      {/* Profile Header */}
      <div className={styles.header}>
        <div
          className={styles.avatar}
          style={{ backgroundColor: `var(--ll-primary-tint)` }}
        >
          <span className={styles.avatarText}>
            {getInitials(profile.username, profile.displayName)}
          </span>
        </div>
        <div className={styles.headerInfo}>
          <h1 className={styles.displayName}>
            {profile.displayName || profile.username}
          </h1>
          <p className={styles.username}>@{profile.username}</p>
          {profile.bio && <p className={styles.bio}>{profile.bio}</p>}
          <p className={styles.memberSince}>
            Member since {formatDate(profile.createdAt)}
          </p>
        </div>
        {isOwnProfile && (
          <Link to="/settings" className={styles.editButton}>
            <Button variant="secondary">Edit Profile</Button>
          </Link>
        )}
      </div>

      {/* Stats Grid */}
      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <span className={styles.statValue}>{profile.stats?.policiesCreated || 0}</span>
          <span className={styles.statLabel}>Policies</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statValue}>{profile.stats?.debateComments || 0}</span>
          <span className={styles.statLabel}>Comments</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statValue}>{profile.stats?.researchSubmitted || 0}</span>
          <span className={styles.statLabel}>Research</span>
        </div>
        <div className={styles.statCard}>
          <span className={styles.statValue}>{profile.stats?.bookmarks || 0}</span>
          <span className={styles.statLabel}>Bookmarks</span>
        </div>
      </div>

      {/* Reputation Card */}
      <div className={styles.section}>
        <h2 className={styles.sectionTitle}>Reputation</h2>
        <div className={styles.reputationCard}>
          <div className={styles.reputationScore}>
            <span className={styles.scoreValue}>{profile.reputation?.score || 0}</span>
            <span className={styles.scoreLabel}>points</span>
          </div>
          <div className={styles.reputationDetails}>
            <div className={styles.reputationItem}>
              <span className={styles.reputationLabel}>Tier</span>
              <span
                className={styles.reputationValue}
                style={{ color: getTierColor(profile.reputation?.tier || 'new') }}
              >
                {(profile.reputation?.tier || 'new').charAt(0).toUpperCase() +
                  (profile.reputation?.tier || 'new').slice(1)}
              </span>
            </div>
            <div className={styles.reputationItem}>
              <span className={styles.reputationLabel}>Contributions</span>
              <span className={styles.reputationValue}>
                {profile.reputation?.contributions || 0}
              </span>
            </div>
            <div className={styles.reputationItem}>
              <span className={styles.reputationLabel}>Account Type</span>
              <span className={styles.reputationValue}>
                {(profile.type || 'human').charAt(0).toUpperCase() +
                  (profile.type || 'human').slice(1)}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
