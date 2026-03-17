import { Link } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { SearchBar } from '../ui'
import styles from './Header.module.css'

export function Header() {
  const { isAuthenticated, user } = useAuth()

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <Link to="/" className={styles.logo}>
          <span className={styles.logoIcon}>🦞</span>
          <span className={styles.logoText}>Lobster Lobby</span>
        </Link>

        <div className={styles.search}>
          <SearchBar placeholder="Search policies..." />
        </div>

        <nav className={styles.nav}>
          {isAuthenticated ? (
            <Link to="/dashboard" className={styles.avatar}>
              <span className={styles.avatarText}>
                {user?.username?.charAt(0).toUpperCase() || 'U'}
              </span>
            </Link>
          ) : (
            <div className={styles.authLinks}>
              <Link to="/login" className={styles.link}>
                Log in
              </Link>
              <Link to="/register" className={styles.registerLink}>
                Register
              </Link>
            </div>
          )}
        </nav>
      </div>
    </header>
  )
}
