import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { useTheme } from '../../contexts/ThemeContext'
import { SearchBar } from '../ui'
import styles from './Header.module.css'

export function Header() {
  const { user, logout } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const isAuthenticated = !!user
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!menuOpen) return
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [menuOpen])

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
            <div className={styles.userMenu} ref={menuRef}>
              <button
                className={styles.avatar}
                onClick={() => setMenuOpen((v) => !v)}
                aria-expanded={menuOpen}
                aria-haspopup="true"
              >
                <span className={styles.avatarText}>
                  {user?.username?.charAt(0).toUpperCase() || 'U'}
                </span>
              </button>
              {menuOpen && (
                <div className={styles.dropdown}>
                  <div className={styles.dropdownUser}>
                    <span className={styles.dropdownUsername}>{user?.username}</span>
                    {user?.email && (
                      <span className={styles.dropdownEmail}>{user.email}</span>
                    )}
                  </div>
                  <div className={styles.dropdownDivider} />
                  <Link
                    to={`/u/${user?.username}`}
                    className={styles.dropdownItem}
                    onClick={() => setMenuOpen(false)}
                  >
                    My Profile
                  </Link>
                  <Link
                    to="/dashboard"
                    className={styles.dropdownItem}
                    onClick={() => setMenuOpen(false)}
                  >
                    Dashboard
                  </Link>
                  <Link
                    to="/bookmarks"
                    className={styles.dropdownItem}
                    onClick={() => setMenuOpen(false)}
                  >
                    Bookmarks
                  </Link>
                  <Link
                    to="/representatives"
                    className={styles.dropdownItem}
                    onClick={() => setMenuOpen(false)}
                  >
                    Representatives
                  </Link>
                  <Link
                    to="/campaigns"
                    className={styles.dropdownItem}
                    onClick={() => setMenuOpen(false)}
                  >
                    Campaigns
                  </Link>
                  <Link
                    to="/settings"
                    className={styles.dropdownItem}
                    onClick={() => setMenuOpen(false)}
                  >
                    Settings
                  </Link>
                  <div className={styles.dropdownDivider} />
                  <button
                    className={styles.dropdownItem}
                    onClick={toggleTheme}
                  >
                    <span className={styles.themeToggleRow}>
                      {theme === 'dark' ? '☀️ Light mode' : '🌙 Dark mode'}
                    </span>
                  </button>
                  <button
                    className={styles.dropdownItem}
                    onClick={() => {
                      setMenuOpen(false)
                      logout()
                    }}
                  >
                    Sign out
                  </button>
                </div>
              )}
            </div>
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
