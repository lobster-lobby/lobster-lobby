import type { HTMLAttributes, KeyboardEvent } from 'react'
import { useRef } from 'react'
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
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([])

  const handleKeyDown = (e: KeyboardEvent<HTMLButtonElement>, index: number) => {
    let nextIndex: number | null = null
    if (e.key === 'ArrowRight') {
      nextIndex = (index + 1) % tabs.length
    } else if (e.key === 'ArrowLeft') {
      nextIndex = (index - 1 + tabs.length) % tabs.length
    }
    if (nextIndex !== null) {
      e.preventDefault()
      tabRefs.current[nextIndex]?.focus()
    }
  }

  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} role="tablist" {...props}>
      {tabs.map((tab, index) => (
        <button
          key={tab.id}
          ref={(el) => { tabRefs.current[index] = el }}
          role="tab"
          aria-selected={activeTab === tab.id}
          tabIndex={activeTab === tab.id ? 0 : -1}
          className={[styles.tab, activeTab === tab.id && styles.active]
            .filter(Boolean)
            .join(' ')}
          onClick={() => onTabChange(tab.id)}
          onKeyDown={(e) => handleKeyDown(e, index)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}
