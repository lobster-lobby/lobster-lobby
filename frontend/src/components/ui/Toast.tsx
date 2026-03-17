import type { HTMLAttributes } from 'react'
import { useEffect } from 'react'
import styles from './Toast.module.css'

export interface ToastProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'success' | 'error' | 'info' | 'warning'
  onClose?: () => void
  autoDismiss?: number
}

const icons = {
  success: (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="10" />
      <path d="m9 12 2 2 4-4" />
    </svg>
  ),
  error: (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="10" />
      <path d="m15 9-6 6M9 9l6 6" />
    </svg>
  ),
  info: (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="10" />
      <path d="M12 16v-4M12 8h.01" />
    </svg>
  ),
  warning: (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
      <path d="M12 9v4M12 17h.01" />
    </svg>
  ),
}

export function Toast({
  variant = 'info',
  onClose,
  autoDismiss,
  className = '',
  children,
  ...props
}: ToastProps) {
  useEffect(() => {
    if (autoDismiss && onClose) {
      const timer = setTimeout(onClose, autoDismiss)
      return () => clearTimeout(timer)
    }
  }, [autoDismiss, onClose])

  return (
    <div
      className={[styles.toast, styles[variant], className].filter(Boolean).join(' ')}
      role="alert"
      {...props}
    >
      <span className={styles.icon}>{icons[variant]}</span>
      <span className={styles.content}>{children}</span>
      {onClose && (
        <button className={styles.close} onClick={onClose} aria-label="Close">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M18 6 6 18M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  )
}
