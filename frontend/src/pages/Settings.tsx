import { useTheme } from '../contexts/ThemeContext'
import styles from './Settings.module.css'

export default function Settings() {
  const { theme, setTheme } = useTheme()

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Settings</h1>
      <p className={styles.subtitle}>Manage your account settings.</p>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Appearance</h2>
        <div className={styles.card}>
          <div className={styles.settingRow}>
            <div className={styles.settingInfo}>
              <span className={styles.settingLabel}>Theme</span>
              <span className={styles.settingDesc}>
                Choose how Lobster Lobby looks to you.
              </span>
            </div>
            <div className={styles.themeSwitch}>
              <button
                className={`${styles.themeOption} ${theme === 'light' ? styles.themeOptionActive : ''}`}
                onClick={() => setTheme('light')}
                aria-pressed={theme === 'light'}
              >
                ☀️ Light
              </button>
              <button
                className={`${styles.themeOption} ${theme === 'dark' ? styles.themeOptionActive : ''}`}
                onClick={() => setTheme('dark')}
                aria-pressed={theme === 'dark'}
              >
                🌙 Dark
              </button>
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}
