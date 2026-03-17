import type { InputHTMLAttributes } from 'react'
import { forwardRef } from 'react'
import styles from './SearchBar.module.css'

export interface SearchBarProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {}

export const SearchBar = forwardRef<HTMLInputElement, SearchBarProps>(
  ({ className = '', ...props }, ref) => {
    return (
      <div className={[styles.wrapper, className].filter(Boolean).join(' ')}>
        <svg
          className={styles.icon}
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
        <input
          ref={ref}
          type="search"
          className={styles.input}
          placeholder="Search..."
          {...props}
        />
      </div>
    )
  }
)

SearchBar.displayName = 'SearchBar'
