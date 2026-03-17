import type { DebateSortOption } from '../../types/debates'
import styles from './SortControls.module.css'

interface SortControlsProps {
  current: DebateSortOption
  onChange: (sort: DebateSortOption) => void
}

const SORT_OPTIONS: { id: DebateSortOption; label: string }[] = [
  { id: 'newest', label: 'Newest' },
  { id: 'top', label: 'Top' },
  { id: 'controversial', label: 'Controversial' },
]

export default function SortControls({ current, onChange }: SortControlsProps) {
  return (
    <div className={styles.wrapper}>
      <span className={styles.label}>Sort by:</span>
      <div className={styles.buttons}>
        {SORT_OPTIONS.map((opt) => (
          <button
            key={opt.id}
            type="button"
            className={[styles.btn, current === opt.id && styles.active].filter(Boolean).join(' ')}
            onClick={() => onChange(opt.id)}
          >
            {opt.label}
          </button>
        ))}
      </div>
    </div>
  )
}
