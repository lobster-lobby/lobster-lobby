import styles from './PositionBar.module.css'

interface PositionBarProps {
  support: number
  oppose: number
  neutral: number
}

export default function PositionBar({ support, oppose, neutral }: PositionBarProps) {
  const total = support + oppose + neutral

  if (total === 0) {
    return (
      <div className={styles.wrapper}>
        <div className={styles.bar}>
          <div className={styles.empty} style={{ width: '100%' }} />
        </div>
        <div className={styles.labels}>
          <span className={styles.label}>No positions yet</span>
        </div>
      </div>
    )
  }

  const supportPct = (support / total) * 100
  const opposePct = (oppose / total) * 100
  const neutralPct = (neutral / total) * 100

  return (
    <div className={styles.wrapper}>
      <div className={styles.bar}>
        {supportPct > 0 && (
          <div
            className={[styles.segment, styles.support].join(' ')}
            style={{ width: `${supportPct}%` }}
            title={`Support: ${support}`}
          />
        )}
        {opposePct > 0 && (
          <div
            className={[styles.segment, styles.oppose].join(' ')}
            style={{ width: `${opposePct}%` }}
            title={`Oppose: ${oppose}`}
          />
        )}
        {neutralPct > 0 && (
          <div
            className={[styles.segment, styles.neutral].join(' ')}
            style={{ width: `${neutralPct}%` }}
            title={`Neutral: ${neutral}`}
          />
        )}
      </div>
      <div className={styles.labels}>
        <span className={[styles.label, styles.supportLabel].join(' ')}>
          Support {support}
        </span>
        <span className={[styles.label, styles.opposeLabel].join(' ')}>
          Oppose {oppose}
        </span>
        <span className={[styles.label, styles.neutralLabel].join(' ')}>
          Neutral {neutral}
        </span>
      </div>
    </div>
  )
}
