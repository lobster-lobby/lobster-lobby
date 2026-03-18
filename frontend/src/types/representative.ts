export interface CivicOfficial {
  name: string
  title: string
  party: string
  phone?: string
  email?: string
  photoUrl?: string
  urls?: string[]
  socialMedia?: Record<string, string>
}

export interface ContactInfo {
  phone?: string
  email?: string
  website?: string
  office?: string
}

export interface Representative {
  id: string
  name: string
  title: string
  party: string
  state: string
  district: string
  photoUrl?: string
  phone?: string
  email?: string
  website?: string
  bio?: string
  contactInfo?: ContactInfo
  socialMedia?: Record<string, string>
  chamber: 'senate' | 'house' | 'governor' | 'local'
  level: 'federal' | 'state' | 'local'
  externalIds?: {
    bioguideId?: string
    govtrackId?: string
  }
  createdAt: string
  updatedAt: string
}

export interface VotingRecord {
  id: string
  representativeId: string
  policyId: string
  vote: 'yea' | 'nay' | 'abstain' | 'absent'
  date: string
  session: string
  notes?: string
  createdAt: string
  updatedAt: string
}

export interface VotingSummary {
  totalVotes: number
  yeaCount: number
  nayCount: number
  abstainCount: number
  absentCount: number
  yeaPercent: number
  nayPercent: number
  abstainPercent: number
}

export interface CivicLookupResponse {
  officials: CivicOfficial[]
}

export interface RepresentativeListResponse {
  representatives: Representative[]
  total: number
  page: number
  perPage: number
}

export interface RepresentativeDetailResponse {
  representative: Representative
  votingSummary: VotingSummary
}

export interface VotingRecordListResponse {
  votes: VotingRecord[]
  total: number
  page: number
  perPage: number
}
