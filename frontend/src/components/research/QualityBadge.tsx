import styles from './QualityBadge.module.css'

interface QualityBadgeProps {
  score: number
}

function getQualityTier(score: number): { label: string; tier: string } {
  if (score >= 70) return { label: 'High Quality', tier: 'high' }
  if (score >= 40) return { label: 'Good', tier: 'good' }
  if (score >= 20) return { label: 'Fair', tier: 'fair' }
  return { label: 'Unrated', tier: 'unrated' }
}

export function QualityBadge({ score }: QualityBadgeProps) {
  const { label, tier } = getQualityTier(score)

  if (tier === 'unrated') return null

  return (
    <span className={[styles.badge, styles[tier]].filter(Boolean).join(' ')}>
      {label}
    </span>
  )
}
