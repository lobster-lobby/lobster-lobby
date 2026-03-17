import type { HTMLAttributes, ReactNode } from 'react'
import styles from './EmptyState.module.css'
import { EmptyBoxIcon, EmptySearchIcon, EmptyDocIcon, ChatBubbleIcon } from './Icons'

export interface EmptyStateProps extends HTMLAttributes<HTMLDivElement> {
  heading: string
  description?: string
  action?: ReactNode
  icon?: 'box' | 'search' | 'doc' | 'chat'
}

function EmptyIllustration({ icon = 'box' }: { icon?: EmptyStateProps['icon'] }) {
  const iconMap = {
    box: EmptyBoxIcon,
    search: EmptySearchIcon,
    doc: EmptyDocIcon,
    chat: ChatBubbleIcon,
  }
  const IconComponent = iconMap[icon!] ?? EmptyBoxIcon

  return (
    <div className={styles.illustration}>
      <IconComponent size={56} />
    </div>
  )
}

export function EmptyState({
  heading,
  description,
  action,
  icon = 'box',
  className = '',
  ...props
}: EmptyStateProps) {
  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} {...props}>
      <EmptyIllustration icon={icon} />
      <h3 className={styles.heading}>{heading}</h3>
      {description && <p className={styles.description}>{description}</p>}
      {action && <div className={styles.action}>{action}</div>}
    </div>
  )
}
