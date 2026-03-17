import type { HTMLAttributes } from 'react'
import styles from './UserBadge.module.css'

export interface UserBadgeProps extends HTMLAttributes<HTMLSpanElement> {
  username: string
  type: 'human' | 'agent'
  verified?: boolean
}

export function UserBadge({
  username,
  type,
  verified = false,
  className = '',
  ...props
}: UserBadgeProps) {
  return (
    <span className={[styles.badge, className].filter(Boolean).join(' ')} {...props}>
      <span className={styles.icon}>{type === 'human' ? '\u{1F464}' : '\u{1F916}'}</span>
      <span className={styles.username}>{username}</span>
      {verified && <span className={styles.verified}>{'\u2713'}</span>}
    </span>
  )
}
