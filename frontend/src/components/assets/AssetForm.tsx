import { useState, useRef } from 'react'
import { Button, Input } from '../ui'
import type { AssetType } from '../../types/asset'
import { ASSET_TYPE_LABELS, TEXT_ASSET_TYPES, FILE_ASSET_TYPES } from '../../types/asset'
import { getAccessToken } from '../../contexts/authTokenStore'
import styles from './AssetForm.module.css'

interface AssetFormProps {
  campaignId: string
  onSuccess: () => void
  onCancel: () => void
}

export function AssetForm({ campaignId, onSuccess, onCancel }: AssetFormProps) {
  const [mode, setMode] = useState<'text' | 'file'>('text')
  const [title, setTitle] = useState('')
  const [assetType, setAssetType] = useState<AssetType>('text_post')
  const [content, setContent] = useState('')
  const [description, setDescription] = useState('')
  const [subjectLine, setSubjectLine] = useState('')
  const [suggestedRecipients, setSuggestedRecipients] = useState('')
  const [aiGenerated, setAiGenerated] = useState(false)
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleModeChange = (newMode: 'text' | 'file') => {
    setMode(newMode)
    if (newMode === 'text') {
      setAssetType('text_post')
      setFile(null)
    } else {
      setAssetType('social_image')
      setContent('')
      setSubjectLine('')
      setSuggestedRecipients('')
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (!selectedFile) return

    // Validate file type
    const allowedTypes = ['image/png', 'image/jpeg', 'image/svg+xml', 'application/pdf']
    if (!allowedTypes.includes(selectedFile.type)) {
      setError('File type not allowed. Please use PNG, JPEG, SVG, or PDF.')
      return
    }

    // Validate file size
    const maxSize = selectedFile.type === 'application/pdf' ? 25 * 1024 * 1024 : 10 * 1024 * 1024
    if (selectedFile.size > maxSize) {
      setError(`File too large. Max size: ${maxSize / (1024 * 1024)}MB`)
      return
    }

    setFile(selectedFile)
    setError(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    const token = getAccessToken()
    if (!token) {
      setError('Please log in to submit assets')
      setLoading(false)
      return
    }

    try {
      if (mode === 'text') {
        // Submit text asset
        const res = await fetch(`/api/campaigns/${campaignId}/assets`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({
            title,
            type: assetType,
            content,
            description,
            subjectLine: assetType === 'email_draft' ? subjectLine : undefined,
            suggestedRecipients: assetType === 'email_draft' ? suggestedRecipients : undefined,
            aiGenerated,
          }),
        })

        if (!res.ok) {
          const data = await res.json()
          throw new Error(data.error || 'Failed to create asset')
        }
      } else {
        // Upload file asset
        if (!file) {
          setError('Please select a file')
          setLoading(false)
          return
        }

        const formData = new FormData()
        formData.append('file', file)
        formData.append('title', title)
        formData.append('type', assetType)
        formData.append('description', description)
        formData.append('aiGenerated', aiGenerated.toString())

        const res = await fetch(`/api/campaigns/${campaignId}/assets/upload`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
          body: formData,
        })

        if (!res.ok) {
          const data = await res.json()
          throw new Error(data.error || 'Failed to upload asset')
        }
      }

      onSuccess()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  const availableTypes = mode === 'text' ? TEXT_ASSET_TYPES : FILE_ASSET_TYPES

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <div className={styles.modeToggle}>
        <button
          type="button"
          className={`${styles.modeBtn} ${mode === 'text' ? styles.active : ''}`}
          onClick={() => handleModeChange('text')}
        >
          Text Asset
        </button>
        <button
          type="button"
          className={`${styles.modeBtn} ${mode === 'file' ? styles.active : ''}`}
          onClick={() => handleModeChange('file')}
        >
          Upload File
        </button>
      </div>

      {error && <div className={styles.error}>{error}</div>}

      <div className={styles.field}>
        <label htmlFor="title">Title *</label>
        <Input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Give your asset a descriptive title"
          required
          minLength={3}
          maxLength={200}
        />
      </div>

      <div className={styles.field}>
        <label htmlFor="type">Type *</label>
        <select
          id="type"
          value={assetType}
          onChange={(e) => setAssetType(e.target.value as AssetType)}
          className={styles.select}
          required
        >
          {availableTypes.map((type) => (
            <option key={type} value={type}>
              {ASSET_TYPE_LABELS[type]}
            </option>
          ))}
        </select>
      </div>

      <div className={styles.field}>
        <label htmlFor="description">Description (optional)</label>
        <Input
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Usage instructions or context"
          maxLength={1000}
        />
      </div>

      {mode === 'text' && (
        <>
          {assetType === 'email_draft' && (
            <>
              <div className={styles.field}>
                <label htmlFor="subjectLine">Subject Line *</label>
                <Input
                  id="subjectLine"
                  value={subjectLine}
                  onChange={(e) => setSubjectLine(e.target.value)}
                  placeholder="Email subject line"
                  required={assetType === 'email_draft'}
                />
              </div>

              <div className={styles.field}>
                <label htmlFor="recipients">Suggested Recipients (optional)</label>
                <Input
                  id="recipients"
                  value={suggestedRecipients}
                  onChange={(e) => setSuggestedRecipients(e.target.value)}
                  placeholder="e.g., senator@congress.gov"
                />
              </div>
            </>
          )}

          <div className={styles.field}>
            <label htmlFor="content">Content * (Markdown supported)</label>
            <textarea
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Write your content here... Markdown is supported."
              className={styles.textarea}
              required
              minLength={10}
              maxLength={50000}
              rows={10}
            />
          </div>
        </>
      )}

      {mode === 'file' && (
        <div className={styles.field}>
          <label>File * (PNG, JPEG, SVG, or PDF)</label>
          <div
            className={styles.dropZone}
            onClick={() => fileInputRef.current?.click()}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept="image/png,image/jpeg,image/svg+xml,application/pdf"
              onChange={handleFileChange}
              className={styles.fileInput}
            />
            {file ? (
              <div className={styles.fileInfo}>
                <span className={styles.fileName}>{file.name}</span>
                <span className={styles.fileSize}>
                  ({(file.size / (1024 * 1024)).toFixed(2)} MB)
                </span>
              </div>
            ) : (
              <div className={styles.dropText}>
                Click to select a file or drag and drop
              </div>
            )}
          </div>
          <p className={styles.hint}>
            Max size: 10MB for images, 25MB for PDFs
          </p>
        </div>
      )}

      <div className={styles.field}>
        <label className={styles.checkbox}>
          <input
            type="checkbox"
            checked={aiGenerated}
            onChange={(e) => setAiGenerated(e.target.checked)}
          />
          <span>This asset was AI-generated</span>
        </label>
      </div>

      <div className={styles.actions}>
        <Button type="button" variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={loading}>
          {loading ? 'Submitting...' : mode === 'text' ? 'Submit Asset' : 'Upload Asset'}
        </Button>
      </div>
    </form>
  )
}
