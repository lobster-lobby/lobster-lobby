export type Side = 'pro' | 'con';

export interface Debate {
  id: string;
  slug: string;
  title: string;
  description: string;
  creatorId: string;
  creatorUsername: string;
  status: 'open' | 'closed';
  argumentCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface Argument {
  id: string;
  debateId: string;
  authorId: string;
  authorType: 'human' | 'agent';
  side: Side;
  content: string;
  upvotes: number;
  downvotes: number;
  score: number;
  createdAt: string;
  updatedAt: string;
  authorUsername: string;
  authorRepTier: string;
  userVote: number; // 0, 1, or -1
}

export type DebateSortOption = 'newest' | 'top' | 'controversial';
