import type { HTMLAttributes } from 'react'
import styles from './TabNav.module.css'

export interface Tab {
  id: string
  label: string
}

export interface TabNavProps extends HTMLAttributes<HTMLDivElement> {
  tabs: Tab[]
  activeTab: string
  onTabChange: (tabId: string) => void
}

export function TabNav({
  tabs,
  activeTab,
  onTabChange,
  className = '',
  ...props
}: TabNavProps) {
  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} role="tablist" {...props}>
      {tabs.map((tab) => (
        <button
          key={tab.id}
          role="tab"
          aria-selected={activeTab === tab.id}
          className={[styles.tab, activeTab === tab.id && styles.active]
            .filter(Boolean)
            .join(' ')}
          onClick={() => onTabChange(tab.id)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}
