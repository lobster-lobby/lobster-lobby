package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	PositionSupport   = "support"
	PositionOppose    = "oppose"
	PositionConsensus = "consensus"
)

const BridgingVisibilityThreshold = 5.0

type Endorsement struct {
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	Position  string        `bson:"position" json:"position"` // endorser's stance
	Verified  bool          `bson:"verified" json:"verified"`
	RepTier   string        `bson:"repTier" json:"repTier"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
}

type SummaryPoint struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PolicyID        bson.ObjectID `bson:"policyId" json:"policyId"`
	AuthorID        bson.ObjectID `bson:"authorId" json:"authorId"`
	SourceCommentID *bson.ObjectID `bson:"sourceCommentId,omitempty" json:"sourceCommentId,omitempty"`
	Content         string         `bson:"content" json:"content"`
	Position        string         `bson:"position" json:"position"` // support|oppose|consensus
	Endorsements    []Endorsement  `bson:"endorsements" json:"endorsements"`
	BridgingScore   float64        `bson:"bridgingScore" json:"bridgingScore"`
	Visible         bool           `bson:"visible" json:"visible"`
	CreatedAt       time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time      `bson:"updatedAt" json:"updatedAt"`
}

type SummaryPointResponse struct {
	SummaryPoint   `bson:",inline"`
	AuthorUsername string `bson:"authorUsername" json:"authorUsername"`
	AuthorRepTier  string `bson:"authorRepTier" json:"authorRepTier"`
	EndorseCount   int    `json:"endorseCount"`
	CrossCount     int    `json:"crossCount"`
	UserEndorsed   bool   `json:"userEndorsed"`
}

func (s *SummaryPoint) Validate() error {
	if s.PolicyID.IsZero() {
		return errors.New("policyId is required")
	}
	if s.AuthorID.IsZero() {
		return errors.New("authorId is required")
	}
	if s.Content == "" {
		return errors.New("content is required")
	}
	if len(s.Content) < 10 {
		return errors.New("content must be at least 10 characters")
	}
	if len(s.Content) > 500 {
		return errors.New("content must be at most 500 characters")
	}
	if s.Position != PositionSupport && s.Position != PositionOppose && s.Position != PositionConsensus {
		return errors.New("position must be one of: support, oppose, consensus")
	}
	return nil
}

func repMultiplier(tier string) float64 {
	switch tier {
	case TierModerator:
		return 1.5
	case TierTrusted:
		return 1.2
	case TierMember:
		return 1.1
	default:
		return 1.0
	}
}

func CalculateBridgingScore(pointPosition string, endorsements []Endorsement) float64 {
	var score float64
	for _, e := range endorsements {
		var base float64
		if e.Position != pointPosition && e.Position != PositionConsensus {
			base = 3.0 // cross-position endorsement
		} else {
			base = 1.0 // same-position endorsement
		}
		rm := repMultiplier(e.RepTier)
		vm := 1.0
		if e.Verified {
			vm = 1.5
		}
		score += base * rm * vm
	}
	return score
}
