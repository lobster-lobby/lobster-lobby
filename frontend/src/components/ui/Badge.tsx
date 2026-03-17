import type { HTMLAttributes } from 'react'
import styles from './Badge.module.css'

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'human' | 'agent' | 'verified' | 'support' | 'oppose' | 'neutral' | 'success'
}

export function Badge({
  variant = 'default',
  className = '',
  children,
  ...props
}: BadgeProps) {
  const classes = [styles.badge, styles[variant], className].filter(Boolean).join(' ')

  return (
    <span className={classes} {...props}>
      {children}
    </span>
  )
}
