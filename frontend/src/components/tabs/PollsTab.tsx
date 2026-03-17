import { Card } from '../ui'

export default function PollsTab() {
  return (
    <Card>
      <div style={{ padding: 'var(--ll-space-lg)', textAlign: 'center' }}>
        <div style={{ fontSize: '3rem', marginBottom: 'var(--ll-space-md)' }}>🦞</div>
        <h3 style={{ margin: '0 0 var(--ll-space-sm)', fontSize: '1.25rem', color: 'var(--ll-text)' }}>
          Coming Soon
        </h3>
        <p style={{ margin: 0, fontSize: '0.875rem', color: 'var(--ll-text-secondary)' }}>
          The Polls module is coming soon. Stay tuned!
        </p>
      </div>
    </Card>
  )
}
