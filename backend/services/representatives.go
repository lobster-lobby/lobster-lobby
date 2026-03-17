package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

// Google Civic API response structures
type civicResponse struct {
	Offices   []civicOffice   `json:"offices"`
	Officials []civicOfficial `json:"officials"`
}

type civicOffice struct {
	Name            string `json:"name"`
	DivisionID      string `json:"divisionId"`
	OfficialIndices []int  `json:"officialIndices"`
}

type civicOfficial struct {
	Name     string          `json:"name"`
	Party    string          `json:"party"`
	Phones   []string        `json:"phones"`
	Emails   []string        `json:"emails"`
	PhotoURL string          `json:"photoUrl"`
	URLs     []string        `json:"urls"`
	Channels []civicChannel  `json:"channels"`
}

type civicChannel struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type RepresentativeService struct {
	repo    *repository.RepresentativeRepository
	apiKey  string
	client  *http.Client
	baseURL string
}

func NewRepresentativeService(repo *repository.RepresentativeRepository, apiKey string) *RepresentativeService {
	return &RepresentativeService{
		repo:    repo,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://civicinfo.googleapis.com/civicinfo/v2",
	}
}

// NewRepresentativeServiceWithBaseURL creates a service with a custom base URL (for testing).
func NewRepresentativeServiceWithBaseURL(repo *repository.RepresentativeRepository, apiKey, baseURL string) *RepresentativeService {
	return &RepresentativeService{
		repo:    repo,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
	}
}

// LookupByAddress queries the Google Civic API for representatives at a given address
// Per PRD-004-R03, addresses are NOT stored
func (s *RepresentativeService) LookupByAddress(ctx context.Context, address string) ([]models.CivicOfficial, error) {
	if s.apiKey == "" {
		// No API key configured - return empty results gracefully
		return []models.CivicOfficial{}, nil
	}

	apiURL := fmt.Sprintf(
		"%s/representatives?address=%s&key=%s",
		s.baseURL,
		url.QueryEscape(address),
		s.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Civic API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// API errors (invalid address, rate limit, etc.) - return empty gracefully
		return []models.CivicOfficial{}, nil
	}

	var civicResp civicResponse
	if err := json.NewDecoder(resp.Body).Decode(&civicResp); err != nil {
		return nil, fmt.Errorf("failed to decode Civic API response: %w", err)
	}

	return s.flattenCivicResponse(civicResp), nil
}

func (s *RepresentativeService) flattenCivicResponse(resp civicResponse) []models.CivicOfficial {
	var officials []models.CivicOfficial

	for _, office := range resp.Offices {
		for _, idx := range office.OfficialIndices {
			if idx < 0 || idx >= len(resp.Officials) {
				continue
			}
			o := resp.Officials[idx]

			official := models.CivicOfficial{
				Name:  o.Name,
				Title: office.Name,
				Party: o.Party,
				URLs:  o.URLs,
			}

			if len(o.Phones) > 0 {
				official.Phone = o.Phones[0]
			}
			if len(o.Emails) > 0 {
				official.Email = o.Emails[0]
			}
			if o.PhotoURL != "" {
				official.PhotoURL = o.PhotoURL
			}

			if len(o.Channels) > 0 {
				official.SocialMedia = make(map[string]string)
				for _, ch := range o.Channels {
					official.SocialMedia[ch.Type] = ch.ID
				}
			}

			officials = append(officials, official)
		}
	}

	if officials == nil {
		officials = []models.CivicOfficial{}
	}
	return officials
}

// GetByState returns representatives from the database cache for a given state
func (s *RepresentativeService) GetByState(ctx context.Context, state string) ([]models.Representative, error) {
	return s.repo.FindByState(ctx, state)
}

// GetByDistrict returns representatives from the database cache for a given district
func (s *RepresentativeService) GetByDistrict(ctx context.Context, district string) ([]models.Representative, error) {
	return s.repo.FindByDistrict(ctx, district)
}

// GetByID returns a representative by ID from the database
func (s *RepresentativeService) GetByID(ctx context.Context, id string) (*models.Representative, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid representative ID: %w", err)
	}
	return s.repo.FindByID(ctx, oid)
}
