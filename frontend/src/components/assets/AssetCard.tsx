import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { VoteButtons, UserBadge, Badge, Button, Toast } from '../ui'
import type { CampaignAsset } from '../../types/asset'
import { ASSET_TYPE_LABELS, isTextBasedAsset } from '../../types/asset'
import { getAccessToken } from '../../contexts/authTokenStore'
import styles from './AssetCard.module.css'

interface AssetCardProps {
  asset: CampaignAsset
  userVote?: number
  onVote: (assetId: string, value: number) => void
  onShare: (assetId: string, platform: string) => void
  onDownload: (assetId: string) => void
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffDay > 30) {
    return date.toLocaleDateString()
  } else if (diffDay > 0) {
    return `${diffDay}d ago`
  } else if (diffHour > 0) {
    return `${diffHour}h ago`
  } else if (diffMin > 0) {
    return `${diffMin}m ago`
  }
  return 'just now'
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function AssetCard({ asset, userVote = 0, onVote, onShare, onDownload }: AssetCardProps) {
  const [expanded, setExpanded] = useState(false)
  const [toast, setToast] = useState<{ message: string; variant: 'success' | 'error' | 'info' } | null>(null)
  const [shareMenuOpen, setShareMenuOpen] = useState(false)

  const currentVote = userVote === 1 ? 'up' : userVote === -1 ? 'down' : null

  const handleUpvote = () => {
    onVote(asset.id, userVote === 1 ? 0 : 1)
  }

  const handleDownvote = () => {
    onVote(asset.id, userVote === -1 ? 0 : -1)
  }

  const handleCopyToClipboard = async () => {
    if (!asset.content) return
    try {
      await navigator.clipboard.writeText(asset.content)
      setToast({ message: 'Copied to clipboard!', variant: 'success' })
    } catch {
      setToast({ message: 'Failed to copy', variant: 'error' })
    }
  }

  const handleEmailAction = () => {
    if (!asset.content) return
    const subject = encodeURIComponent(asset.subjectLine || asset.title)
    const body = encodeURIComponent(asset.content)
    const recipients = asset.suggestedRecipients ? encodeURIComponent(asset.suggestedRecipients) : ''
    window.open(`mailto:${recipients}?subject=${subject}&body=${body}`, '_blank')
    onShare(asset.id, 'email')
  }

  const handleShare = (platform: string) => {
    onShare(asset.id, platform)
    setShareMenuOpen(false)

    const shareUrl = window.location.href
    const text = encodeURIComponent(asset.title)

    switch (platform) {
      case 'twitter':
        window.open(`https://twitter.com/intent/tweet?text=${text}&url=${encodeURIComponent(shareUrl)}`, '_blank')
        break
      case 'facebook':
        window.open(`https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(shareUrl)}`, '_blank')
        break
      case 'print':
        window.print()
        break
      default:
        navigator.clipboard.writeText(shareUrl).then(() => {
          setToast({ message: 'Link copied!', variant: 'success' })
        })
    }
  }

  const handleDownload = () => {
    onDownload(asset.id)
    if (asset.fileUrl) {
      // For file assets, trigger download
      const token = getAccessToken()
      const downloadUrl = `/api/campaigns/${asset.campaignId}/assets/${asset.id}/download`

      const link = document.createElement('a')
      link.href = downloadUrl
      if (token) {
        // For authenticated downloads, we need to handle this differently
        fetch(downloadUrl, {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${token}` },
        })
          .then(res => res.blob())
          .then(blob => {
            const url = window.URL.createObjectURL(blob)
            link.href = url
            link.download = asset.fileName || 'download'
            link.click()
            window.URL.revokeObjectURL(url)
          })
          .catch(() => {
            setToast({ message: 'Download failed', variant: 'error' })
          })
      } else {
        link.download = asset.fileName || 'download'
        link.click()
      }
    }
  }

  const isTextBased = isTextBasedAsset(asset.type)

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.titleRow}>
          <h3 className={styles.title}>{asset.title}</h3>
          <div className={styles.badges}>
            <span className={`${styles.typeBadge} ${styles[asset.type]}`}>
              {ASSET_TYPE_LABELS[asset.type]}
            </span>
            {asset.aiGenerated && (
              <Badge variant="agent">AI Generated</Badge>
            )}
          </div>
        </div>

        <div className={styles.meta}>
          <UserBadge username={asset.createdByUsername} type="human" />
          <span className={styles.metaDivider}>|</span>
          <span>{formatRelativeTime(asset.createdAt)}</span>
          {asset.fileSize && (
            <>
              <span className={styles.metaDivider}>|</span>
              <span>{formatFileSize(asset.fileSize)}</span>
            </>
          )}
        </div>

        {asset.description && (
          <p className={styles.description}>{asset.description}</p>
        )}

        <div className={styles.stats}>
          <VoteButtons
            upvotes={asset.upvotes}
            downvotes={asset.downvotes}
            userVote={currentVote}
            onUpvote={handleUpvote}
            onDownvote={handleDownvote}
          />
          <span className={styles.stat}>
            {asset.downloadCount} downloads
          </span>
          <span className={styles.stat}>
            {asset.shareCount} shares
          </span>
          {isTextBased && (
            <button
              className={styles.expandBtn}
              onClick={() => setExpanded(!expanded)}
              type="button"
            >
              {expanded ? 'Collapse' : 'Expand'}
            </button>
          )}
        </div>
      </div>

      {/* Content preview for text assets */}
      {isTextBased && expanded && asset.content && (
        <div className={styles.content}>
          {asset.type === 'email_draft' && asset.subjectLine && (
            <div className={styles.emailSubject}>
              <strong>Subject:</strong> {asset.subjectLine}
            </div>
          )}
          {asset.type === 'email_draft' && asset.suggestedRecipients && (
            <div className={styles.emailRecipients}>
              <strong>To:</strong> {asset.suggestedRecipients}
            </div>
          )}
          <div className={styles.markdown}>
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {asset.content}
            </ReactMarkdown>
          </div>
        </div>
      )}

      {/* File preview for file assets */}
      {!isTextBased && asset.fileUrl && (
        <div className={styles.filePreview}>
          {asset.mimeType?.startsWith('image/') ? (
            <img
              src={asset.fileUrl}
              alt={asset.title}
              className={styles.previewImage}
              onClick={() => setExpanded(!expanded)}
            />
          ) : (
            <div className={styles.pdfPreview}>
              <span className={styles.pdfIcon}>PDF</span>
              <span>{asset.fileName}</span>
            </div>
          )}
        </div>
      )}

      {/* Actions */}
      <div className={styles.actions}>
        {isTextBased && (
          <>
            <Button size="sm" variant="secondary" onClick={handleCopyToClipboard}>
              Copy Text
            </Button>
            {asset.type === 'email_draft' && (
              <Button size="sm" variant="secondary" onClick={handleEmailAction}>
                Open in Email
              </Button>
            )}
          </>
        )}
        {!isTextBased && (
          <Button size="sm" variant="primary" onClick={handleDownload}>
            Download
          </Button>
        )}
        <div className={styles.shareDropdown}>
          <Button
            size="sm"
            variant="secondary"
            onClick={() => setShareMenuOpen(!shareMenuOpen)}
          >
            Share
          </Button>
          {shareMenuOpen && (
            <div className={styles.shareMenu}>
              <button onClick={() => handleShare('twitter')}>Twitter</button>
              <button onClick={() => handleShare('facebook')}>Facebook</button>
              <button onClick={() => handleShare('email')}>Email</button>
              <button onClick={() => handleShare('print')}>Print</button>
              <button onClick={() => handleShare('other')}>Copy Link</button>
            </div>
          )}
        </div>
      </div>

      {toast && (
        <div className={styles.toastWrapper}>
          <Toast variant={toast.variant} onClose={() => setToast(null)} autoDismiss={2000}>
            {toast.message}
          </Toast>
        </div>
      )}
    </div>
  )
}
