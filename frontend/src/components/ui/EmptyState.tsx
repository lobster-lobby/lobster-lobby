import type { HTMLAttributes, ReactNode } from 'react'
import styles from './EmptyState.module.css'
import { EmptyBoxIcon, EmptySearchIcon, EmptyDocIcon, ChatBubbleIcon } from './Icons'

export interface EmptyStateAction {
  label: string
  onClick: () => void
}

export interface EmptyStateProps extends HTMLAttributes<HTMLDivElement> {
  /** Primary heading text. Alias: `heading` */
  title?: string
  /** @deprecated Use `title` instead */
  heading?: string
  description?: string
  action?: EmptyStateAction | ReactNode
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

function isActionObject(action: unknown): action is EmptyStateAction {
  return (
    typeof action === 'object' &&
    action !== null &&
    'label' in action &&
    'onClick' in action
  )
}

export function EmptyState({
  title,
  heading,
  description,
  action,
  icon = 'box',
  className = '',
  ...props
}: EmptyStateProps) {
  const displayTitle = title ?? heading

  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} {...props}>
      <div className={styles.lobsterDecor} aria-hidden="true">🦞</div>
      <EmptyIllustration icon={icon} />
      {displayTitle && <h3 className={styles.heading}>{displayTitle}</h3>}
      {description && <p className={styles.description}>{description}</p>}
      {action && (
        <div className={styles.action}>
          {isActionObject(action) ? (
            <button className={styles.actionButton} onClick={action.onClick}>
              {action.label}
            </button>
          ) : (
            action
          )}
        </div>
      )}
    </div>
  )
}
