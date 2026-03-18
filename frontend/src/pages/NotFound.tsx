import { Link } from 'react-router-dom'
import styles from './NotFound.module.css'

export default function NotFound() {
  return (
    <div className={styles.page}>
      <div className={styles.mascot} role="img" aria-label="Lobster">
        🦞
      </div>
      <div className={styles.code}>404</div>
      <h1 className={styles.title}>This page got away...</h1>
      <p className={styles.description}>
        Looks like this page scuttled off into the deep. It may have been moved,
        removed, or never existed in the first place.
      </p>
      <Link to="/" className={styles.homeLink}>
        Go Home
      </Link>
    </div>
  )
}
