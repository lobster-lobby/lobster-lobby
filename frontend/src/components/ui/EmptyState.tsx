import type { HTMLAttributes, ReactNode } from 'react'
import styles from './EmptyState.module.css'
import { EmptyBoxIcon, EmptySearchIcon, EmptyDocIcon, ChatBubbleIcon } from './Icons'
import { LobsterMascot, type MascotPose } from './LobsterMascot'

export interface EmptyStateProps extends HTMLAttributes<HTMLDivElement> {
  heading: string
  description?: string
  action?: ReactNode
  /**
   * Icon variant to display when no mascot is provided.
   * Mutually exclusive with `mascot` — if `mascot` is set, `icon` is ignored.
   */
  icon?: 'box' | 'search' | 'doc' | 'chat'
  /**
   * Lobster mascot pose to display instead of an icon.
   * Mutually exclusive with `icon` — takes precedence when both are provided.
   */
  mascot?: MascotPose
}

function EmptyIllustration({
  icon = 'box',
  mascot,
}: {
  icon?: EmptyStateProps['icon']
  mascot?: MascotPose
}) {
  if (mascot) {
    return (
      <div className={styles.illustration} style={{ width: 120, height: 120, background: 'transparent' }}>
        <LobsterMascot pose={mascot} width={100} height={100} />
      </div>
    )
  }

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
  mascot,
  className = '',
  ...props
}: EmptyStateProps) {
  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} {...props}>
      <EmptyIllustration icon={icon} mascot={mascot} />
      <h3 className={styles.heading}>{heading}</h3>
      {description && <p className={styles.description}>{description}</p>}
      {action && <div className={styles.action}>{action}</div>}
    </div>
  )
}
