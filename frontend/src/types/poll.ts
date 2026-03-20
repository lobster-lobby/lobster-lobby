export interface PollOption {
  id: string
  text: string
  votes: number
}

export interface Poll {
  id: string
  policyId: string
  authorId: string
  authorName: string
  question: string
  options: PollOption[]
  multiSelect: boolean
  endsAt?: string
  status: 'active' | 'closed'
  createdAt: string
  updatedAt: string
  totalVotes: number
  userVoteOptionIds?: string[]
}

export interface CreatePollPayload {
  question: string
  options: string[]
  multiSelect: boolean
  endsAt?: string
}
