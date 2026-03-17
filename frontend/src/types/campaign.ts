export interface CampaignComment {
  id: string
  campaignId: string
  parentId?: string
  authorId: string
  authorName: string
  body: string
  votes: number
  createdAt: string
  updatedAt: string
}

export interface CampaignCommentListResponse {
  comments: CampaignComment[]
  userVotes: Record<string, number>
}

export type CampaignEventType =
  | 'created'
  | 'asset_added'
  | 'milestone'
  | 'status_change'
  | 'comment_milestone'

export interface CampaignEvent {
  id: string
  campaignId: string
  type: CampaignEventType
  title: string
  description: string
  metadata?: Record<string, unknown>
  createdAt: string
}

export interface CampaignEventListResponse {
  events: CampaignEvent[]
}

export interface DailyActivity {
  date: string
  count: number
}

export interface CampaignMetrics {
  totalDownloads: number
  totalShares: number
  sharesByPlatform: Record<string, number>
  uniqueParticipants: number
  assetCount: number
  commentCount: number
}

export interface CampaignActivityResponse {
  dailyActivity: DailyActivity[]
  recentEvents: CampaignEvent[]
  metrics: CampaignMetrics
}

export const EVENT_TYPE_CONFIG: Record<
  CampaignEventType,
  { label: string; color: string; icon: string }
> = {
  created: { label: 'Created', color: 'var(--ll-success)', icon: '+' },
  asset_added: { label: 'Asset Added', color: 'var(--ll-primary)', icon: 'A' },
  milestone: { label: 'Milestone', color: 'var(--ll-accent)', icon: 'M' },
  status_change: { label: 'Status Change', color: 'var(--ll-warning)', icon: 'S' },
  comment_milestone: { label: 'Comment Milestone', color: 'var(--ll-secondary)', icon: 'C' },
}
