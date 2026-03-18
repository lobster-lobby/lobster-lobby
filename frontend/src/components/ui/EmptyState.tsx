import type { HTMLAttributes, ReactNode } from 'react'
import styles from './EmptyState.module.css'
import { EmptyBoxIcon, EmptySearchIcon, EmptyDocIcon, ChatBubbleIcon } from './Icons'

export interface EmptyStateAction {
  label: string
  onClick: () => void
}

export interface EmptyStateProps extends HTMLAttributes<HTMLDivElement> {
  /** Primary text. `title` and `heading` are interchangeable. */
  title?: string
  /** @deprecated Use `title` instead. */
  heading?: string
  description?: string
  /** Pass a ReactNode for full control, or an EmptyStateAction for a simple button. */
  action?: ReactNode | EmptyStateAction
  /** Preset icon name, emoji string, or any ReactNode. */
  icon?: 'box' | 'search' | 'doc' | 'chat' | ReactNode
}

const ICON_MAP = {
  box: EmptyBoxIcon,
  search: EmptySearchIcon,
  doc: EmptyDocIcon,
  chat: ChatBubbleIcon,
} as const

function resolveIcon(icon: EmptyStateProps['icon']) {
  if (icon === undefined) {
    return (
      <div className={styles.illustration}>
        <EmptyBoxIcon size={56} />
      </div>
    )
  }

  if (typeof icon === 'string' && icon in ICON_MAP) {
    const IconComponent = ICON_MAP[icon as keyof typeof ICON_MAP]
    return (
      <div className={styles.illustration}>
        <IconComponent size={56} />
      </div>
    )
  }

  // String emoji or ReactNode
  if (typeof icon === 'string') {
    return <div className={styles.emoji}>{icon}</div>
  }

  return <div className={styles.illustration}>{icon}</div>
}

function isActionConfig(action: unknown): action is EmptyStateAction {
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
      {resolveIcon(icon)}
      {displayTitle && <h3 className={styles.heading}>{displayTitle}</h3>}
      {description && <p className={styles.description}>{description}</p>}
      {action && (
        <div className={styles.action}>
          {isActionConfig(action) ? (
            <button type="button" className={styles.actionBtn} onClick={action.onClick}>
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
