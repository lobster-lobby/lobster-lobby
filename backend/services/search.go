package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/models"
)

const policiesIndex = "policies"

// PolicyDocument is the search document structure for Meilisearch.
type PolicyDocument struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Tags          []string `json:"tags"`
	Type          string   `json:"type"`
	Level         string   `json:"level"`
	State         string   `json:"state"`
	Status        string   `json:"status"`
	BillNumber    string   `json:"billNumber"`
	HotScore      float64  `json:"hotScore"`
	CreatedAt     int64    `json:"createdAt"`
	DebateCount   int      `json:"debateCount"`
	ResearchCount int      `json:"researchCount"`
	BookmarkCount int      `json:"bookmarkCount"`
}

// SearchResult is a single result with highlights.
type SearchResult struct {
	PolicyDocument
	Highlights map[string]interface{} `json:"_highlights,omitempty"`
}

// SearchResponse is the paginated search response.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"perPage"`
	Query   string         `json:"query"`
}

// SearchFilters holds optional search filters.
type SearchFilters struct {
	Type   string
	Level  string
	State  string
	Status string
}

// SearchService wraps Meilisearch operations.
type SearchService struct {
	client meilisearch.ServiceManager
	logger *zap.Logger
	ready  bool
}

// NewSearchService creates a SearchService. Connection failures are non-fatal.
func NewSearchService(url, apiKey string, logger *zap.Logger) *SearchService {
	svc := &SearchService{logger: logger}

	client := meilisearch.New(url, meilisearch.WithAPIKey(apiKey))
	svc.client = client

	if _, err := client.Health(); err != nil {
		logger.Warn("meilisearch not reachable — search disabled", zap.String("url", url), zap.Error(err))
		return svc
	}

	if err := svc.configureIndex(); err != nil {
		logger.Warn("failed to configure meilisearch index", zap.Error(err))
		return svc
	}

	svc.ready = true
	logger.Info("meilisearch connected", zap.String("url", url))
	return svc
}

func strSlicePtr(ss []string) *[]string {
	return &ss
}

func ifaceSlicePtr(ss []string) *[]interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return &out
}

func (s *SearchService) configureIndex() error {
	idx := s.client.Index(policiesIndex)

	if _, err := idx.UpdateSearchableAttributes(strSlicePtr([]string{
		"title", "summary", "tags", "billNumber",
	})); err != nil {
		return fmt.Errorf("searchable attributes: %w", err)
	}

	if _, err := idx.UpdateFilterableAttributes(ifaceSlicePtr([]string{
		"type", "level", "state", "status",
	})); err != nil {
		return fmt.Errorf("filterable attributes: %w", err)
	}

	if _, err := idx.UpdateSortableAttributes(strSlicePtr([]string{
		"hotScore", "createdAt", "debateCount",
	})); err != nil {
		return fmt.Errorf("sortable attributes: %w", err)
	}

	return nil
}

func policyToDoc(p *models.Policy) PolicyDocument {
	return PolicyDocument{
		ID:            p.ID.Hex(),
		Title:         p.Title,
		Summary:       p.Summary,
		Tags:          p.Tags,
		Type:          string(p.Type),
		Level:         string(p.Level),
		State:         p.State,
		Status:        string(p.Status),
		BillNumber:    p.BillNumber,
		HotScore:      p.HotScore,
		CreatedAt:     p.CreatedAt.Unix(),
		DebateCount:   p.Engagement.DebateCount,
		ResearchCount: p.Engagement.ResearchCount,
		BookmarkCount: p.Engagement.BookmarkCount,
	}
}

// IndexPolicy adds or updates a policy document in Meilisearch.
func (s *SearchService) IndexPolicy(_ context.Context, p *models.Policy) error {
	if !s.ready {
		return nil
	}
	pk := "id"
	doc := policyToDoc(p)
	_, err := s.client.Index(policiesIndex).AddDocuments([]PolicyDocument{doc}, &meilisearch.DocumentOptions{PrimaryKey: &pk})
	return err
}

// RemovePolicy removes a policy document from Meilisearch.
func (s *SearchService) RemovePolicy(_ context.Context, id string) error {
	if !s.ready {
		return nil
	}
	_, err := s.client.Index(policiesIndex).DeleteDocument(id, nil)
	return err
}

// SearchPolicies performs a full-text search with optional filters and pagination.
func (s *SearchService) SearchPolicies(_ context.Context, query string, filters SearchFilters, page, perPage int) (*SearchResponse, error) {
	if !s.ready {
		return &SearchResponse{Results: []SearchResult{}, Query: query, Page: page, PerPage: perPage}, nil
	}

	var filterParts []string
	if filters.Type != "" {
		filterParts = append(filterParts, fmt.Sprintf("type = %q", filters.Type))
	}
	if filters.Level != "" {
		filterParts = append(filterParts, fmt.Sprintf("level = %q", filters.Level))
	}
	if filters.State != "" {
		filterParts = append(filterParts, fmt.Sprintf("state = %q", filters.State))
	}
	if filters.Status != "" {
		filterParts = append(filterParts, fmt.Sprintf("status = %q", filters.Status))
	}

	offset := int64((page - 1) * perPage)
	req := &meilisearch.SearchRequest{
		Offset:                offset,
		Limit:                 int64(perPage),
		AttributesToHighlight: []string{"title", "summary"},
		HighlightPreTag:       "<mark>",
		HighlightPostTag:      "</mark>",
	}
	if len(filterParts) > 0 {
		req.Filter = strings.Join(filterParts, " AND ")
	}

	res, err := s.client.Index(policiesIndex).Search(query, req)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(res.Hits))
	for _, hit := range res.Hits {
		var doc PolicyDocument
		if err := hit.Decode(&doc); err != nil {
			continue
		}
		sr := SearchResult{PolicyDocument: doc}

		// Extract _formatted highlights
		if raw, ok := hit["_formatted"]; ok {
			var hl map[string]interface{}
			if err := json.Unmarshal(raw, &hl); err == nil {
				sr.Highlights = hl
			}
		}
		results = append(results, sr)
	}

	return &SearchResponse{
		Results: results,
		Total:   res.EstimatedTotalHits,
		Page:    page,
		PerPage: perPage,
		Query:   query,
	}, nil
}

// SimilarPolicy represents a similar policy match from search.
type SimilarPolicy struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Slug       string  `json:"slug"`
	Similarity float64 `json:"similarity"`
	Type       string  `json:"type"`
}

// FindSimilar searches for policies similar to the given title and summary.
// Returns up to 5 matches with ranking score > 0.3.
func (s *SearchService) FindSimilar(_ context.Context, title, summary string) ([]SimilarPolicy, error) {
	if !s.ready {
		return []SimilarPolicy{}, nil
	}

	// Combine title + first 200 chars of summary for search query
	query := title
	if len(summary) > 200 {
		query += " " + summary[:200]
	} else if summary != "" {
		query += " " + summary
	}

	req := &meilisearch.SearchRequest{
		Limit:            5,
		ShowRankingScore: true,
	}

	res, err := s.client.Index(policiesIndex).Search(query, req)
	if err != nil {
		return nil, err
	}

	var results []SimilarPolicy
	for _, hit := range res.Hits {
		var score float64
		if raw, ok := hit["_rankingScore"]; ok {
			_ = json.Unmarshal(raw, &score)
		}
		if score < 0.3 {
			continue
		}

		var doc PolicyDocument
		if err := hit.Decode(&doc); err != nil {
			continue
		}

		results = append(results, SimilarPolicy{
			ID:         doc.ID,
			Title:      doc.Title,
			Slug:       toSlug(doc.Title),
			Similarity: math.Round(score*100) / 100,
			Type:       doc.Type,
		})
	}

	if results == nil {
		results = []SimilarPolicy{}
	}
	return results, nil
}

// toSlug is a minimal helper to produce a slug from a title for search results.
// The authoritative slug comes from the DB; this is a best-effort fallback.
func toSlug(title string) string {
	s := strings.ToLower(title)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
	s = strings.Trim(s, "-")
	return s
}

// BulkIndex indexes a slice of policies.
func (s *SearchService) BulkIndex(_ context.Context, policies []*models.Policy) error {
	if !s.ready || len(policies) == 0 {
		return nil
	}
	docs := make([]PolicyDocument, len(policies))
	for i, p := range policies {
		docs[i] = policyToDoc(p)
	}
	pk := "id"
	_, err := s.client.Index(policiesIndex).AddDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: &pk})
	return err
}

// Ready returns whether the service is connected.
func (s *SearchService) Ready() bool {
	return s.ready
}
