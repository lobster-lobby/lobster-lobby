package models

import (
	"errors"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PolicyType string

const (
	PolicyTypeExistingLaw PolicyType = "existing_law"
	PolicyTypeActiveBill  PolicyType = "active_bill"
	PolicyTypeProposed    PolicyType = "proposed"
)

type PolicyLevel string

const (
	PolicyLevelFederal PolicyLevel = "federal"
	PolicyLevelState   PolicyLevel = "state"
)

type PolicyStatus string

const (
	PolicyStatusActive            PolicyStatus = "active"
	PolicyStatusPassed            PolicyStatus = "passed"
	PolicyStatusFailed            PolicyStatus = "failed"
	PolicyStatusWithdrawn         PolicyStatus = "withdrawn"
	PolicyStatusArchived          PolicyStatus = "archived"
	PolicyStatusReadyForCampaign  PolicyStatus = "ready_for_campaign"
)

type EngagementStats struct {
	DebateCount   int `bson:"debateCount" json:"debateCount"`
	ResearchCount int `bson:"researchCount" json:"researchCount"`
	PollCount     int `bson:"pollCount" json:"pollCount"`
	BookmarkCount int `bson:"bookmarkCount" json:"bookmarkCount"`
	ViewCount     int `bson:"viewCount" json:"viewCount"`
}

type Policy struct {
	ID             bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title          string          `bson:"title" json:"title"`
	Slug           string          `bson:"slug" json:"slug"`
	Summary        string          `bson:"summary" json:"summary"`
	Type           PolicyType      `bson:"type" json:"type"`
	Level          PolicyLevel     `bson:"level" json:"level"`
	State          string          `bson:"state,omitempty" json:"state,omitempty"`
	Status         PolicyStatus    `bson:"status" json:"status"`
	ExternalURL    string          `bson:"externalUrl,omitempty" json:"externalUrl,omitempty"`
	BillNumber     string          `bson:"billNumber,omitempty" json:"billNumber,omitempty"`
	Tags           []string        `bson:"tags" json:"tags"`
	CreatedBy      bson.ObjectID   `bson:"createdBy" json:"createdBy"`
	LinkedPolicies []bson.ObjectID `bson:"linkedPolicies,omitempty" json:"linkedPolicies,omitempty"`
	ParentPolicy   *bson.ObjectID  `bson:"parentPolicy,omitempty" json:"parentPolicy,omitempty"`
	Engagement     EngagementStats `bson:"engagement" json:"engagement"`
	HotScore       float64         `bson:"hotScore" json:"hotScore"`
	CreatedAt      time.Time       `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time       `bson:"updatedAt" json:"updatedAt"`
}

func (p *Policy) Validate() error {
	if len(p.Title) < 5 || len(p.Title) > 300 {
		return errors.New("title must be between 5 and 300 characters")
	}

	if len(p.Summary) < 20 || len(p.Summary) > 5000 {
		return errors.New("summary must be between 20 and 5000 characters")
	}

	if p.Type != PolicyTypeExistingLaw && p.Type != PolicyTypeActiveBill && p.Type != PolicyTypeProposed {
		return errors.New("type must be one of: existing_law, active_bill, proposed")
	}

	if p.Level != PolicyLevelFederal && p.Level != PolicyLevelState {
		return errors.New("level must be one of: federal, state")
	}

	if p.Level == PolicyLevelState && p.State == "" {
		return errors.New("state is required when level is state")
	}

	if (p.Type == PolicyTypeExistingLaw || p.Type == PolicyTypeActiveBill) && p.ExternalURL == "" {
		return errors.New("externalUrl is required for existing_law and active_bill types")
	}

	if p.ExternalURL != "" {
		if _, err := url.ParseRequestURI(p.ExternalURL); err != nil {
			return errors.New("externalUrl must be a valid URL")
		}
	}

	if len(p.Tags) > 10 {
		return errors.New("tags cannot exceed 10 items")
	}

	if p.Status == "" {
		p.Status = PolicyStatusActive
	} else if p.Status != PolicyStatusActive && p.Status != PolicyStatusPassed &&
		p.Status != PolicyStatusFailed && p.Status != PolicyStatusWithdrawn && p.Status != PolicyStatusArchived {
		return errors.New("status must be one of: active, passed, failed, withdrawn, archived")
	}

	return nil
}
