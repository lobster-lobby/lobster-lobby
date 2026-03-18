import { Component } from 'react'
import type { ErrorInfo, ReactNode } from 'react'
import styles from './ErrorBoundary.module.css'

interface ErrorBoundaryProps {
  children: ReactNode
}

interface ErrorBoundaryState {
  hasError: boolean
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

  handleReset = () => {
    this.setState({ hasError: false })
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className={styles.page}>
          <div className={styles.icon} role="img" aria-label="Error">
            🦞
          </div>
          <div className={styles.code}>500</div>
          <h1 className={styles.title}>Something went wrong</h1>
          <p className={styles.description}>
            Our lobster hit a snag. The error has been noted — try refreshing
            or head back home.
          </p>
          <div className={styles.actions}>
            <button
              type="button"
              className={styles.primaryBtn}
              onClick={this.handleReset}
            >
              Try Again
            </button>
            <a href="/" className={styles.secondaryBtn}>
              Go Home
            </a>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
