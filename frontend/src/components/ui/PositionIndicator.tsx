import type { HTMLAttributes } from 'react'
import styles from './PositionIndicator.module.css'

export interface PositionIndicatorProps extends HTMLAttributes<HTMLSpanElement> {
  position: 'support' | 'oppose' | 'neutral'
}

const labels = {
  support: 'Support',
  oppose: 'Oppose',
  neutral: 'Neutral',
}

export function PositionIndicator({
  position,
  className = '',
  ...props
}: PositionIndicatorProps) {
  return (
    <span
      className={[styles.indicator, styles[position], className].filter(Boolean).join(' ')}
      {...props}
    >
      {labels[position]}
    </span>
  )
}
