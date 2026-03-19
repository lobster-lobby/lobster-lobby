import { useAuth } from '../hooks/useAuth'
import { Link, useNavigate } from 'react-router-dom'
import styles from './Dashboard.module.css'

const sections = [
  {
    icon: '🔖',
    title: 'Bookmarks',
    description: 'Policies and debates you saved for later.',
    link: '/bookmarks',
  },
  {
    icon: '💬',
    title: 'Activity',
    description: 'Your comments, reactions, and endorsements.',
    link: '/activity',
  },
  {
    icon: '⭐',
    title: 'Reputation & Stats',
    description: 'Your contribution score and community standing.',
    link: '/reputation',
  },
  {
    icon: '🏛️',
    title: 'Your Representatives',
    description: 'Track votes and contact your elected officials.',
    link: '/representatives',
  },
]

export default function Dashboard() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <div className={styles.dashboard}>
      <div className={styles['dashboard-header']}>
        <div>
          <h1 className={styles['dashboard-title']}>Dashboard</h1>
          <p className={styles['dashboard-subtitle']}>Welcome back{user?.username ? `, ${user.username}` : ''}!</p>
        </div>
        <div className={styles['dashboard-header-actions']}>
          <Link to="/settings" className={styles['dashboard-settings-link']}>
            ⚙️ Settings
          </Link>
          <button className={styles['dashboard-logout-btn']} onClick={handleLogout}>
            Sign Out
          </button>
        </div>
      </div>

      <div className={styles['dashboard-grid']}>
        {sections.map((s) => (
          <Link to={s.link} key={s.title} className={styles['dashboard-card']}>
            <span className={styles['dashboard-card-icon']}>{s.icon}</span>
            <div>
              <h2 className={styles['dashboard-card-title']}>{s.title}</h2>
              <p className={styles['dashboard-card-desc']}>{s.description}</p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
