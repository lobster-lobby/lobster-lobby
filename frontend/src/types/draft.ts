export type DraftCategory = 'amendment' | 'talking-point' | 'position-statement' | 'full-text'
export type DraftStatus = 'draft' | 'published' | 'archived'

export interface Draft {
  id: string
  policyId: string
  authorId: string
  authorName: string
  title: string
  content: string
  category: DraftCategory
  status: DraftStatus
  endorsements: number
  version: number
  createdAt: string
  updatedAt: string
  userEndorsed?: boolean
}

export interface DraftComment {
  id: string
  draftId: string
  authorId: string
  authorName: string
  content: string
  createdAt: string
}

export interface CreateDraftPayload {
  title: string
  content: string
  category: DraftCategory
  status: DraftStatus
}
