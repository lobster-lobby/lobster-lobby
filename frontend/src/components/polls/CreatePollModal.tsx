import { useState } from 'react'
import { getAccessToken } from '../../hooks/useAuth'
import { Modal, Button, Input, Textarea } from '../ui'
import type { Poll } from '../../types/poll'
import styles from './CreatePollModal.module.css'

interface CreatePollModalProps {
  policyId: string
  isOpen: boolean
  onClose: () => void
  onCreated: (poll: Poll) => void
}

export default function CreatePollModal({ policyId, isOpen, onClose, onCreated }: CreatePollModalProps) {
  const [question, setQuestion] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [multiSelect, setMultiSelect] = useState(false)
  const [endsAt, setEndsAt] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  function addOption() {
    if (options.length < 6) setOptions([...options, ''])
  }

  function removeOption(i: number) {
    if (options.length <= 2) return
    setOptions(options.filter((_, idx) => idx !== i))
  }

  function updateOption(i: number, val: string) {
    const next = [...options]
    next[i] = val
    setOptions(next)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const filled = options.filter((o) => o.trim())
    if (!question.trim() || filled.length < 2) {
      setError('A question and at least 2 options are required.')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      const token = getAccessToken()
      const body: Record<string, unknown> = { question: question.trim(), options: filled, multiSelect }
      if (endsAt) body.endsAt = new Date(endsAt).toISOString()
      const res = await fetch(`/api/policies/${policyId}/polls`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) throw new Error('Failed to create poll')
      const poll: Poll = await res.json()
      onCreated(poll)
      onClose()
      setQuestion('')
      setOptions(['', ''])
      setMultiSelect(false)
      setEndsAt('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create poll')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Create Poll">
      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.field}>
          <label className={styles.label}>Question</label>
          <Textarea
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            placeholder="What do you want to ask?"
            maxLength={280}
            rows={2}
            required
          />
          <span className={styles.charCount}>{question.length}/280</span>
        </div>

        <div className={styles.field}>
          <label className={styles.label}>Options</label>
          {options.map((opt, i) => (
            <div key={i} className={styles.optionRow}>
              <Input
                value={opt}
                onChange={(e) => updateOption(i, e.target.value)}
                placeholder={`Option ${i + 1}`}
                maxLength={200}
              />
              {options.length > 2 && (
                <button type="button" className={styles.removeBtn} onClick={() => removeOption(i)}>✕</button>
              )}
            </div>
          ))}
          {options.length < 6 && (
            <button type="button" className={styles.addBtn} onClick={addOption}>+ Add option</button>
          )}
        </div>

        <div className={styles.checkRow}>
          <input
            id="multiSelect"
            type="checkbox"
            checked={multiSelect}
            onChange={(e) => setMultiSelect(e.target.checked)}
          />
          <label htmlFor="multiSelect">Allow multiple selections</label>
        </div>

        <div className={styles.field}>
          <label className={styles.label}>End date (optional)</label>
          <input
            type="datetime-local"
            value={endsAt}
            onChange={(e) => setEndsAt(e.target.value)}
            className={styles.dateInput}
          />
        </div>

        {error && <p className={styles.error}>{error}</p>}

        <div className={styles.actions}>
          <Button type="button" variant="secondary" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={submitting}>{submitting ? 'Creating…' : 'Create Poll'}</Button>
        </div>
      </form>
    </Modal>
  )
}
