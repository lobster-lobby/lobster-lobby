import { NavLink } from 'react-router-dom'
import styles from './BottomNav.module.css'

const navItems = [
  { to: '/', label: 'Home', icon: '🏠' },
  { to: '/search', label: 'Search', icon: '🔍' },
  { to: '/policies/new', label: 'Create', icon: '✏️' },
  { to: '/representatives', label: 'Reps', icon: '🏛️' },
  { to: '/dashboard', label: 'Profile', icon: '👤' },
]

export function BottomNav() {
  return (
    <nav className={styles.nav}>
      {navItems.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          className={({ isActive }) =>
            `${styles.item} ${isActive ? styles.active : ''}`
          }
        >
          <span className={styles.icon}>{item.icon}</span>
          <span className={styles.label}>{item.label}</span>
        </NavLink>
      ))}
    </nav>
  )
}
