import { Link } from 'react-router-dom'
import styles from './NotFoundPage.module.css'

export default function NotFoundPage() {
  return (
    <div className={styles.page}>
      <div className={styles.lobster} role="img" aria-label="Lobster">
        🦞
      </div>
      <p className={styles.code}>404</p>
      <h1 className={styles.title}>Oops! This page got away...</h1>
      <p className={styles.description}>
        Looks like this page scuttled off into the deep. It might have been moved,
        deleted, or maybe it never existed. Even lobsters get lost sometimes!
      </p>
      <Link to="/" className={styles.homeButton}>
        🏠 Go Home
      </Link>
      <div className={styles.trail} aria-hidden="true">
        <span>🦞</span>
        <span>·</span>
        <span>·</span>
        <span>·</span>
        <span>🫧</span>
      </div>
    </div>
  )
}
