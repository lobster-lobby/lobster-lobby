import type { HTMLAttributes, ReactNode } from 'react'
import styles from './EmptyState.module.css'

export interface EmptyStateProps extends HTMLAttributes<HTMLDivElement> {
  heading: string
  description?: string
  action?: ReactNode
}

export function EmptyState({
  heading,
  description,
  action,
  className = '',
  ...props
}: EmptyStateProps) {
  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} {...props}>
      <h3 className={styles.heading}>{heading}</h3>
      {description && <p className={styles.description}>{description}</p>}
      {action && <div className={styles.action}>{action}</div>}
    </div>
  )
}
