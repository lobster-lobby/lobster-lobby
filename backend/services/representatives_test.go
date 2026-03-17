package services

import (
	"testing"

	"github.com/lobster-lobby/lobster-lobby/models"
)

func newSvc() *RepresentativeService {
	return &RepresentativeService{}
}

func TestFlattenCivicResponse_Empty(t *testing.T) {
	svc := newSvc()
	result := svc.flattenCivicResponse(civicResponse{})
	if result == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(result))
	}
}

func TestFlattenCivicResponse_NilSlices(t *testing.T) {
	svc := newSvc()
	resp := civicResponse{
		Offices:   nil,
		Officials: nil,
	}
	result := svc.flattenCivicResponse(resp)
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

func TestFlattenCivicResponse_IndexOutOfBounds(t *testing.T) {
	svc := newSvc()
	resp := civicResponse{
		Offices: []civicOffice{
			{Name: "Senate", OfficialIndices: []int{0, 5, -1}}, // 5 and -1 are out of bounds
		},
		Officials: []civicOfficial{
			{Name: "Alice"},
		},
	}
	result := svc.flattenCivicResponse(resp)
	if len(result) != 1 {
		t.Fatalf("expected 1 valid official, got %d", len(result))
	}
	if result[0].Name != "Alice" {
		t.Errorf("expected Alice, got %s", result[0].Name)
	}
}

func TestFlattenCivicResponse_ChannelMapping(t *testing.T) {
	svc := newSvc()
	resp := civicResponse{
		Offices: []civicOffice{
			{Name: "Mayor", OfficialIndices: []int{0}},
		},
		Officials: []civicOfficial{
			{
				Name: "Bob",
				Channels: []civicChannel{
					{Type: "Twitter", ID: "bob_tweets"},
					{Type: "Facebook", ID: "bob.fb"},
				},
			},
		},
	}
	result := svc.flattenCivicResponse(resp)
	if len(result) != 1 {
		t.Fatalf("expected 1 official, got %d", len(result))
	}
	if result[0].SocialMedia["Twitter"] != "bob_tweets" {
		t.Errorf("expected twitter id bob_tweets, got %s", result[0].SocialMedia["Twitter"])
	}
	if result[0].SocialMedia["Facebook"] != "bob.fb" {
		t.Errorf("expected facebook id bob.fb, got %s", result[0].SocialMedia["Facebook"])
	}
}

func TestFlattenCivicResponse_NoChannels(t *testing.T) {
	svc := newSvc()
	resp := civicResponse{
		Offices: []civicOffice{
			{Name: "Governor", OfficialIndices: []int{0}},
		},
		Officials: []civicOfficial{
			{Name: "Carol", Channels: nil},
		},
	}
	result := svc.flattenCivicResponse(resp)
	if len(result) != 1 {
		t.Fatalf("expected 1 official")
	}
	if result[0].SocialMedia != nil {
		t.Errorf("expected nil SocialMedia for official with no channels")
	}
}

func TestFlattenCivicResponse_PhoneAndEmail(t *testing.T) {
	svc := newSvc()
	resp := civicResponse{
		Offices: []civicOffice{
			{Name: "Rep", OfficialIndices: []int{0}},
		},
		Officials: []civicOfficial{
			{
				Name:   "Dave",
				Phones: []string{"555-1234", "555-5678"},
				Emails: []string{"dave@gov.gov", "dave2@gov.gov"},
			},
		},
	}
	result := svc.flattenCivicResponse(resp)
	if result[0].Phone != "555-1234" {
		t.Errorf("expected first phone, got %s", result[0].Phone)
	}
	if result[0].Email != "dave@gov.gov" {
		t.Errorf("expected first email, got %s", result[0].Email)
	}
}

func TestFlattenCivicResponse_MultipleOffices(t *testing.T) {
	svc := newSvc()
	resp := civicResponse{
		Offices: []civicOffice{
			{Name: "Senate", OfficialIndices: []int{0, 1}},
			{Name: "House", OfficialIndices: []int{2}},
		},
		Officials: []civicOfficial{
			{Name: "Sen A"},
			{Name: "Sen B"},
			{Name: "Rep C"},
		},
	}
	result := svc.flattenCivicResponse(resp)
	if len(result) != 3 {
		t.Fatalf("expected 3 officials, got %d", len(result))
	}
	titles := map[string]string{}
	for _, o := range result {
		titles[o.Name] = o.Title
	}
	if titles["Sen A"] != "Senate" || titles["Sen B"] != "Senate" {
		t.Error("expected Senate title for senators")
	}
	if titles["Rep C"] != "House" {
		t.Error("expected House title for rep")
	}
}

// Ensure return type is always []models.CivicOfficial (not nil)
var _ []models.CivicOfficial = (func() []models.CivicOfficial {
	svc := &RepresentativeService{}
	return svc.flattenCivicResponse(civicResponse{})
})()
