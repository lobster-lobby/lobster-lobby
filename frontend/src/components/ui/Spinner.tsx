import type { HTMLAttributes } from 'react'
import styles from './Spinner.module.css'

export interface SpinnerProps extends HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg'
}

export function Spinner({ size = 'md', className = '', ...props }: SpinnerProps) {
  return (
    <div
      className={[styles.spinner, styles[size], className].filter(Boolean).join(' ')}
      role="status"
      aria-label="Loading"
      {...props}
    >
      <span className={styles.srOnly}>Loading...</span>
    </div>
  )
}
