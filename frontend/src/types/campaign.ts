export interface CampaignComment {
  id: string
  campaignId: string
  parentId?: string
  authorId: string
  authorName: string
  body: string
  pinned: boolean
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

export type CampaignActivityType = 'join' | 'share' | 'comment' | 'upload'

export interface CampaignActivity {
  id: string
  type: CampaignActivityType
  userId: string
  campaignId: string
  description: string
  createdAt: string
}

export interface CampaignActivityListResponse {
  activities: CampaignActivity[]
  total: number
  page: number
  limit: number
}

export interface ReachMetrics {
  totalSupporters: number
  totalShares: number
  totalDownloads: number
  uniqueParticipants: number
  trendingScore: number
}

export const ACTIVITY_TYPE_CONFIG: Record<
  CampaignActivityType,
  { label: string; color: string; icon: string }
> = {
  join: { label: 'Joined', color: 'var(--ll-success)', icon: 'J' },
  share: { label: 'Shared', color: 'var(--ll-info)', icon: 'S' },
  comment: { label: 'Commented', color: 'var(--ll-accent)', icon: 'C' },
  upload: { label: 'Uploaded', color: 'var(--ll-primary)', icon: 'U' },
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
