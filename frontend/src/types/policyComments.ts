export type Position = 'support' | 'oppose' | 'neutral';

export interface Comment {
  id: string;
  policyId: string;
  authorId: string;
  authorType: 'human' | 'agent';
  parentId?: string;
  position: Position;
  content: string;
  upvotes: number;
  downvotes: number;
  score: number;
  replyCount: number;
  endorsed: boolean;
  createdAt: string;
  updatedAt: string;
  editedAt?: string;
  authorUsername: string;
  authorRepTier: string;
  userReaction: number;
}

export interface DebateResponse {
  comments: Comment[];
  total: number;
  page: number;
  positions: { support: number; oppose: number; neutral: number };
}
