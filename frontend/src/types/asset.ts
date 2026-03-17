export type AssetType =
  | 'text_post'
  | 'email_draft'
  | 'social_image'
  | 'infographic'
  | 'flyer'
  | 'letter'
  | 'video_script'
  | 'talking_points'

export interface CampaignAsset {
  id: string
  campaignId: string
  createdBy: string
  createdByUsername: string
  title: string
  type: AssetType
  content?: string
  fileUrl?: string
  fileName?: string
  fileSize?: number
  mimeType?: string
  description: string
  subjectLine?: string
  suggestedRecipients?: string
  upvotes: number
  downvotes: number
  score: number
  downloadCount: number
  shareCount: number
  sharesByPlatform: Record<string, number>
  aiGenerated: boolean
  commentCount: number
  createdAt: string
  updatedAt: string
}

export interface AssetListResponse {
  assets: CampaignAsset[]
  total: number
  page: number
  perPage: number
}

export interface AssetResponse {
  asset: CampaignAsset
  userVote?: number
}

export const ASSET_TYPE_LABELS: Record<AssetType, string> = {
  text_post: 'Text Post',
  email_draft: 'Email Draft',
  social_image: 'Social Image',
  infographic: 'Infographic',
  flyer: 'Flyer',
  letter: 'Letter',
  video_script: 'Video Script',
  talking_points: 'Talking Points',
}

export const TEXT_ASSET_TYPES: AssetType[] = [
  'text_post',
  'email_draft',
  'letter',
  'video_script',
  'talking_points',
]

export const FILE_ASSET_TYPES: AssetType[] = [
  'social_image',
  'infographic',
  'flyer',
]

export function isTextBasedAsset(type: AssetType): boolean {
  return TEXT_ASSET_TYPES.includes(type)
}

export function isFileBasedAsset(type: AssetType): boolean {
  return FILE_ASSET_TYPES.includes(type)
}
