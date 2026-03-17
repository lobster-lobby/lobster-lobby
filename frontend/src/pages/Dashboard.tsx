import { useAuth } from '../hooks/useAuth'
import { Link, useNavigate } from 'react-router-dom'
import './Dashboard.css'

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
    <div className="dashboard">
      <div className="dashboard-header">
        <div>
          <h1 className="dashboard-title">Dashboard</h1>
          <p className="dashboard-subtitle">Welcome back{user?.username ? `, ${user.username}` : ''}!</p>
        </div>
        <div className="dashboard-header-actions">
          <Link to="/settings" className="dashboard-settings-link">
            ⚙️ Settings
          </Link>
          <button className="dashboard-logout-btn" onClick={handleLogout}>
            Sign Out
          </button>
        </div>
      </div>

      <div className="dashboard-grid">
        {sections.map((s) => (
          <Link to={s.link} key={s.title} className="dashboard-card">
            <span className="dashboard-card-icon">{s.icon}</span>
            <div>
              <h2 className="dashboard-card-title">{s.title}</h2>
              <p className="dashboard-card-desc">{s.description}</p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
