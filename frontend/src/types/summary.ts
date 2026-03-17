export type SummaryPosition = 'support' | 'oppose' | 'consensus';

export interface Endorsement {
  userId: string;
  position: string;
  verified: boolean;
  repTier: string;
  createdAt: string;
}

export interface SummaryPoint {
  id: string;
  policyId: string;
  authorId: string;
  sourceCommentId?: string;
  content: string;
  position: SummaryPosition;
  endorsements: Endorsement[];
  bridgingScore: number;
  visible: boolean;
  createdAt: string;
  updatedAt: string;
  authorUsername: string;
  authorRepTier: string;
  endorseCount: number;
  crossCount: number;
  userEndorsed: boolean;
}

export interface SummaryResponse {
  support: SummaryPoint[];
  oppose: SummaryPoint[];
  consensus: SummaryPoint[];
}
