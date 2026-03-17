import { Card } from '../ui'

interface DebateTabProps {
  policyId: string
}

export default function DebateTab({ policyId }: DebateTabProps) {
  return (
    <Card>
      <div style={{ padding: 'var(--ll-space-lg)', textAlign: 'center' }}>
        <h3 style={{ margin: '0 0 var(--ll-space-sm)', fontSize: '1.25rem', color: 'var(--ll-text)' }}>
          Debate Module
        </h3>
        <p style={{ margin: 0, fontSize: '0.875rem', color: 'var(--ll-text-secondary)' }}>
          Coming in a future update. This will display structured debates for policy {policyId}.
        </p>
      </div>
    </Card>
  )
}
