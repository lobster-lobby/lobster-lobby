import styles from './SourceBadge.module.css'

interface SourceBadgeProps {
  institutional: boolean
  url: string
}

function extractDomain(url: string): string {
  try {
    const parsed = new URL(url)
    return parsed.hostname.replace(/^www\./, '')
  } catch {
    return url
  }
}

export function SourceBadge({ institutional, url }: SourceBadgeProps) {
  const domain = extractDomain(url)

  return (
    <span
      className={[styles.badge, institutional && styles.institutional]
        .filter(Boolean)
        .join(' ')}
    >
      {institutional && <span className={styles.label}>Institutional</span>}
      <span className={styles.domain}>{domain}</span>
    </span>
  )
}
