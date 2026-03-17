package handlers

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

// ResolveCampaign looks up a campaign by ID or slug.
// If idOrSlug is a valid ObjectID, it looks up by ID; otherwise by slug.
func ResolveCampaign(ctx context.Context, campaigns *repository.CampaignRepository, idOrSlug string) (*models.Campaign, error) {
	if oid, parseErr := bson.ObjectIDFromHex(idOrSlug); parseErr == nil {
		return campaigns.GetByID(ctx, oid.Hex())
	}
	return campaigns.FindBySlug(ctx, idOrSlug)
}
