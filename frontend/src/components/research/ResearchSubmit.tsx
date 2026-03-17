import { useState } from 'react'
import { useAuth } from '../../hooks/useAuth'
import { Button, Input, Textarea } from '../ui'
import type { ResearchResponse, ResearchType } from '../../types/research'
import styles from './ResearchSubmit.module.css'

interface ResearchSubmitProps {
  policyId: string
  onSubmitted: (research: ResearchResponse) => void
  onCancel?: () => void
}

const RESEARCH_TYPES: { value: ResearchType; label: string }[] = [
  { value: 'analysis', label: 'Analysis' },
  { value: 'news', label: 'News' },
  { value: 'data', label: 'Data' },
  { value: 'academic', label: 'Academic' },
  { value: 'government', label: 'Government' },
]

interface SourceInput {
  url: string
  title: string
}

export function ResearchSubmit({ policyId, onSubmitted, onCancel }: ResearchSubmitProps) {
  const { isAuthenticated } = useAuth()
  const [title, setTitle] = useState('')
  const [type, setType] = useState<ResearchType>('analysis')
  const [content, setContent] = useState('')
  const [sources, setSources] = useState<SourceInput[]>([{ url: '', title: '' }])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  if (!isAuthenticated) {
    return (
      <div className={styles.wrapper}>
        <div className={styles.header}>
          <h3 className={styles.title}>Submit Research</h3>
        </div>
        <div className={styles.loginPrompt}>
          <p>You must be logged in to submit research.</p>
          <Button variant="primary" onClick={() => window.location.href = '/login'}>
            Log In
          </Button>
        </div>
      </div>
    )
  }

  const handleAddSource = () => {
    setSources([...sources, { url: '', title: '' }])
  }

  const handleRemoveSource = (idx: number) => {
    if (sources.length > 1) {
      setSources(sources.filter((_, i) => i !== idx))
    }
  }

  const handleSourceChange = (idx: number, field: 'url' | 'title', value: string) => {
    const updated = [...sources]
    updated[idx][field] = value
    setSources(updated)
  }

  const validate = (): string | null => {
    if (title.length < 5 || title.length > 300) {
      return 'Title must be between 5 and 300 characters'
    }
    if (content.length < 50) {
      return 'Content must be at least 50 characters'
    }
    const validSources = sources.filter(s => s.url.trim() && s.title.trim())
    if (validSources.length < 1) {
      return 'At least one source with URL and title is required'
    }
    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }

    setLoading(true)
    try {
      const token = localStorage.getItem('ll_token')
      const validSources = sources
        .filter(s => s.url.trim() && s.title.trim())
        .map(s => ({ url: s.url.trim(), title: s.title.trim() }))

      const res = await fetch(`/api/policies/${policyId}/research`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          title: title.trim(),
          type,
          content: content.trim(),
          sources: validSources,
        }),
      })

      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.error || 'Failed to submit research')
      }

      const data = await res.json()
      onSubmitted(data.research)

      // Reset form
      setTitle('')
      setType('analysis')
      setContent('')
      setSources([{ url: '', title: '' }])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit research')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.header}>
        <h3 className={styles.title}>Submit Research</h3>
      </div>
      <form className={styles.form} onSubmit={handleSubmit}>
        {error && <div className={styles.error}>{error}</div>}

        <div className={styles.row}>
          <Input
            label="Title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Research title (5-300 characters)"
            required
          />
          <div>
            <label className={styles.label}>Type</label>
            <select
              className={styles.typeSelect}
              value={type}
              onChange={(e) => setType(e.target.value as ResearchType)}
            >
              {RESEARCH_TYPES.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        <Textarea
          label="Content (Markdown supported)"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Write your research analysis here... (minimum 50 characters)"
          rows={8}
          required
        />

        <div className={styles.sourcesSection}>
          <div className={styles.sourcesHeader}>
            <h4 className={styles.sourcesTitle}>Sources</h4>
          </div>
          {sources.map((source, idx) => (
            <div key={idx} className={styles.sourceItem}>
              <Input
                placeholder="https://example.com/article"
                value={source.url}
                onChange={(e) => handleSourceChange(idx, 'url', e.target.value)}
              />
              <Input
                placeholder="Source title"
                value={source.title}
                onChange={(e) => handleSourceChange(idx, 'title', e.target.value)}
              />
              {sources.length > 1 && (
                <button
                  type="button"
                  className={styles.removeBtn}
                  onClick={() => handleRemoveSource(idx)}
                  aria-label="Remove source"
                >
                  X
                </button>
              )}
            </div>
          ))}
          <button
            type="button"
            className={styles.addSourceBtn}
            onClick={handleAddSource}
          >
            + Add another source
          </button>
        </div>

        <div className={styles.actions}>
          {onCancel && (
            <Button type="button" variant="ghost" onClick={onCancel}>
              Cancel
            </Button>
          )}
          <Button type="submit" variant="primary" disabled={loading}>
            {loading ? 'Submitting...' : 'Submit Research'}
          </Button>
        </div>
      </form>
    </div>
  )
}
