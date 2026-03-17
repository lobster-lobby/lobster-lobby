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

export interface CivicLookupResponse {
  officials: CivicOfficial[]
}

export interface RepresentativeListResponse {
  representatives: Representative[]
}
