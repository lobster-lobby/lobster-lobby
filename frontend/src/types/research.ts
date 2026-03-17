export interface Source {
  url: string
  title: string
  publisher?: string
  publishedDate?: string
  institutional: boolean
}

export type ResearchType = 'analysis' | 'news' | 'data' | 'academic' | 'government'

export interface ResearchResponse {
  id: string
  policyId: string
  authorId: string
  authorUsername: string
  authorRepTier: string
  authorType: string
  title: string
  type: ResearchType
  content: string
  sources: Source[]
  upvotes: number
  downvotes: number
  score: number
  citedBy: number
  userVote: number
  createdAt: string
  updatedAt: string
}

export interface ResearchListResponse {
  research: ResearchResponse[]
  total: number
  page: number
}
