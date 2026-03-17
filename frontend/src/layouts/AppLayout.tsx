import { Outlet } from 'react-router-dom'
import { Header } from '../components/layout/Header'
import { BottomNav } from '../components/layout/BottomNav'
import styles from './AppLayout.module.css'

export function AppLayout() {
  return (
    <div className={styles.layout}>
      <Header />
      <main className={styles.main}>
        <div className={styles.content}>
          <Outlet />
        </div>
      </main>
      <BottomNav />
    </div>
  )
}
