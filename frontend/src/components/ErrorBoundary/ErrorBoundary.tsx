import { Component } from 'react'
import type { ErrorInfo, ReactNode } from 'react'
import styles from './ErrorBoundary.module.css'

interface ErrorBoundaryProps {
  children: ReactNode
}

interface ErrorBoundaryState {
  hasError: boolean
}

function ErrorPage() {
  return (
    <div className={styles.page}>
      <div className={styles.lobster} role="img" aria-label="Lobster in distress">
        🦞💥
      </div>
      <h1 className={styles.title}>Something went wrong</h1>
      <p className={styles.description}>
        Our lobster hit a snag! Don't worry, these things happen.
        Try refreshing the page or head back to safety.
      </p>
      <div className={styles.actions}>
        <button
          className={styles.retryButton}
          onClick={() => window.location.reload()}
        >
          🔄 Try Again
        </button>
        <a href="/" className={styles.homeLink}>
          🏠 Go Home
        </a>
      </div>
    </div>
  )
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return <ErrorPage />
    }

    return this.props.children
  }
}
