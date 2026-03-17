import { Link } from 'react-router-dom'
import { LobsterMascot } from '../components/ui'

export default function NotFound() {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '60vh',
      textAlign: 'center',
      padding: 'var(--ll-space-xl)',
    }}>
      <LobsterMascot pose="thinking" width={160} height={160} />
      <h1 style={{ marginTop: 'var(--ll-space-lg)', marginBottom: 'var(--ll-space-sm)' }}>
        404 - Page Not Found
      </h1>
      <p style={{ color: 'var(--ll-text-secondary)', marginBottom: 'var(--ll-space-lg)' }}>
        The page you're looking for doesn't exist.
      </p>
      <Link
        to="/"
        style={{
          color: 'var(--ll-primary)',
          textDecoration: 'none',
          fontWeight: 500,
        }}
      >
        Go home
      </Link>
    </div>
  )
}
