package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/lobster-lobby/lobster-lobby/models"
)

// getTestDB returns a MongoDB connection for integration tests.
// It uses MONGODB_TEST_URI env var, or skips if not set.
func getTestDB(t *testing.T) *MongoDB {
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	db, err := Connect(uri, "lobster_lobby_test")
	if err != nil {
		t.Skipf("skipping integration test: could not connect to MongoDB: %v", err)
	}

	return db
}

func cleanupComments(t *testing.T, db *MongoDB) {
	ctx := context.Background()
	db.Database.Collection("campaign_comments").Drop(ctx)
	db.Database.Collection("campaign_comment_votes").Drop(ctx)
}

func TestCampaignCommentRepository_Create(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	campaignID := bson.NewObjectID()
	authorID := bson.NewObjectID()

	comment := &models.CampaignComment{
		CampaignID: campaignID,
		AuthorID:   authorID,
		AuthorName: "testuser",
		Body:       "This is a test comment",
	}

	err := repo.Create(ctx, comment)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if comment.ID.IsZero() {
		t.Error("expected ID to be set after Create")
	}
	if comment.Votes != 0 {
		t.Errorf("expected votes to be 0, got %d", comment.Votes)
	}
	if comment.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCampaignCommentRepository_GetByID(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	campaignID := bson.NewObjectID()
	authorID := bson.NewObjectID()

	comment := &models.CampaignComment{
		CampaignID: campaignID,
		AuthorID:   authorID,
		AuthorName: "testuser",
		Body:       "Test comment for GetByID",
	}

	err := repo.Create(ctx, comment)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test GetByID
	found, err := repo.GetByID(ctx, comment.ID.Hex())
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected comment to be found")
	}
	if found.Body != "Test comment for GetByID" {
		t.Errorf("expected body %q, got %q", "Test comment for GetByID", found.Body)
	}

	// Test GetByID with non-existent ID
	notFound, err := repo.GetByID(ctx, bson.NewObjectID().Hex())
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent comment")
	}
}

func TestCampaignCommentRepository_ListByCampaign(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	campaignID := bson.NewObjectID()
	otherCampaignID := bson.NewObjectID()
	authorID := bson.NewObjectID()

	// Create comments for our campaign
	for i := 0; i < 3; i++ {
		comment := &models.CampaignComment{
			CampaignID: campaignID,
			AuthorID:   authorID,
			AuthorName: "testuser",
			Body:       "Comment " + string(rune('A'+i)),
		}
		if err := repo.Create(ctx, comment); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // ensure different timestamps
	}

	// Create comment for different campaign
	otherComment := &models.CampaignComment{
		CampaignID: otherCampaignID,
		AuthorID:   authorID,
		AuthorName: "testuser",
		Body:       "Other campaign comment",
	}
	if err := repo.Create(ctx, otherComment); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List comments for our campaign
	opts := CampaignCommentListOpts{
		CampaignID: campaignID.Hex(),
		Sort:       "newest",
	}
	comments, err := repo.ListByCampaign(ctx, opts)
	if err != nil {
		t.Fatalf("ListByCampaign failed: %v", err)
	}

	if len(comments) != 3 {
		t.Errorf("expected 3 comments, got %d", len(comments))
	}

	// Verify sorted by newest first
	if len(comments) >= 2 {
		if comments[0].CreatedAt.Before(comments[1].CreatedAt) {
			t.Error("expected comments sorted by newest first")
		}
	}
}

func TestCampaignCommentRepository_Update(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	comment := &models.CampaignComment{
		CampaignID: bson.NewObjectID(),
		AuthorID:   bson.NewObjectID(),
		AuthorName: "testuser",
		Body:       "Original body",
	}

	if err := repo.Create(ctx, comment); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the comment
	updates := bson.M{"body": "Updated body"}
	updated, err := repo.Update(ctx, comment.ID.Hex(), updates)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Body != "Updated body" {
		t.Errorf("expected body %q, got %q", "Updated body", updated.Body)
	}
	if updated.UpdatedAt.Equal(comment.UpdatedAt) {
		t.Error("expected UpdatedAt to change")
	}
}

func TestCampaignCommentRepository_Delete(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	comment := &models.CampaignComment{
		CampaignID: bson.NewObjectID(),
		AuthorID:   bson.NewObjectID(),
		AuthorName: "testuser",
		Body:       "Comment to delete",
	}

	if err := repo.Create(ctx, comment); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the comment
	if err := repo.Delete(ctx, comment.ID.Hex()); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's deleted
	found, err := repo.GetByID(ctx, comment.ID.Hex())
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found != nil {
		t.Error("expected comment to be deleted")
	}
}

func TestCampaignCommentRepository_ToggleVote(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	comment := &models.CampaignComment{
		CampaignID: bson.NewObjectID(),
		AuthorID:   bson.NewObjectID(),
		AuthorName: "testuser",
		Body:       "Comment for voting",
	}

	if err := repo.Create(ctx, comment); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	userID := bson.NewObjectID().Hex()

	// Test upvote
	newVote, err := repo.ToggleVote(ctx, comment.ID.Hex(), userID, 1)
	if err != nil {
		t.Fatalf("ToggleVote failed: %v", err)
	}
	if newVote != 1 {
		t.Errorf("expected vote value 1, got %d", newVote)
	}

	// Verify comment votes updated
	updated, _ := repo.GetByID(ctx, comment.ID.Hex())
	if updated.Votes != 1 {
		t.Errorf("expected comment votes 1, got %d", updated.Votes)
	}

	// Toggle same vote (should remove)
	newVote, err = repo.ToggleVote(ctx, comment.ID.Hex(), userID, 1)
	if err != nil {
		t.Fatalf("ToggleVote failed: %v", err)
	}
	if newVote != 0 {
		t.Errorf("expected vote value 0 after toggle, got %d", newVote)
	}

	updated, _ = repo.GetByID(ctx, comment.ID.Hex())
	if updated.Votes != 0 {
		t.Errorf("expected comment votes 0 after toggle, got %d", updated.Votes)
	}

	// Test change from upvote to downvote
	repo.ToggleVote(ctx, comment.ID.Hex(), userID, 1) // upvote first
	newVote, err = repo.ToggleVote(ctx, comment.ID.Hex(), userID, -1)
	if err != nil {
		t.Fatalf("ToggleVote failed: %v", err)
	}
	if newVote != -1 {
		t.Errorf("expected vote value -1, got %d", newVote)
	}

	updated, _ = repo.GetByID(ctx, comment.ID.Hex())
	if updated.Votes != -1 {
		t.Errorf("expected comment votes -1, got %d", updated.Votes)
	}
}

func TestCampaignCommentRepository_GetUserVote(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	comment := &models.CampaignComment{
		CampaignID: bson.NewObjectID(),
		AuthorID:   bson.NewObjectID(),
		AuthorName: "testuser",
		Body:       "Comment for GetUserVote",
	}

	if err := repo.Create(ctx, comment); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	userID := bson.NewObjectID().Hex()

	// No vote yet
	vote, err := repo.GetUserVote(ctx, comment.ID.Hex(), userID)
	if err != nil {
		t.Fatalf("GetUserVote failed: %v", err)
	}
	if vote != 0 {
		t.Errorf("expected vote 0 for no vote, got %d", vote)
	}

	// Add vote
	repo.ToggleVote(ctx, comment.ID.Hex(), userID, 1)

	vote, err = repo.GetUserVote(ctx, comment.ID.Hex(), userID)
	if err != nil {
		t.Fatalf("GetUserVote failed: %v", err)
	}
	if vote != 1 {
		t.Errorf("expected vote 1, got %d", vote)
	}
}

func TestCampaignCommentRepository_GetBatchUserVotes(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	campaignID := bson.NewObjectID()
	authorID := bson.NewObjectID()
	userID := bson.NewObjectID().Hex()

	// Create multiple comments
	var commentIDs []string
	for i := 0; i < 3; i++ {
		comment := &models.CampaignComment{
			CampaignID: campaignID,
			AuthorID:   authorID,
			AuthorName: "testuser",
			Body:       "Comment " + string(rune('A'+i)),
		}
		if err := repo.Create(ctx, comment); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		commentIDs = append(commentIDs, comment.ID.Hex())
	}

	// Vote on first and third comments
	repo.ToggleVote(ctx, commentIDs[0], userID, 1)
	repo.ToggleVote(ctx, commentIDs[2], userID, -1)

	// Get batch votes
	votes, err := repo.GetBatchUserVotes(ctx, commentIDs, userID)
	if err != nil {
		t.Fatalf("GetBatchUserVotes failed: %v", err)
	}

	if votes[commentIDs[0]] != 1 {
		t.Errorf("expected vote 1 for comment 0, got %d", votes[commentIDs[0]])
	}
	if votes[commentIDs[1]] != 0 {
		t.Errorf("expected vote 0 for comment 1, got %d", votes[commentIDs[1]])
	}
	if votes[commentIDs[2]] != -1 {
		t.Errorf("expected vote -1 for comment 2, got %d", votes[commentIDs[2]])
	}
}

func TestCampaignCommentRepository_CountByCampaign(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	campaignID := bson.NewObjectID()
	authorID := bson.NewObjectID()

	// Create comments
	for i := 0; i < 5; i++ {
		comment := &models.CampaignComment{
			CampaignID: campaignID,
			AuthorID:   authorID,
			AuthorName: "testuser",
			Body:       "Comment " + string(rune('A'+i)),
		}
		if err := repo.Create(ctx, comment); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	count, err := repo.CountByCampaign(ctx, campaignID.Hex())
	if err != nil {
		t.Fatalf("CountByCampaign failed: %v", err)
	}

	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestCampaignCommentRepository_DeleteRemovesVotes(t *testing.T) {
	db := getTestDB(t)
	defer db.Disconnect()
	cleanupComments(t, db)

	repo := NewCampaignCommentRepository(db)
	ctx := context.Background()

	comment := &models.CampaignComment{
		CampaignID: bson.NewObjectID(),
		AuthorID:   bson.NewObjectID(),
		AuthorName: "testuser",
		Body:       "Comment with votes",
	}

	if err := repo.Create(ctx, comment); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Add some votes
	user1 := bson.NewObjectID().Hex()
	user2 := bson.NewObjectID().Hex()
	repo.ToggleVote(ctx, comment.ID.Hex(), user1, 1)
	repo.ToggleVote(ctx, comment.ID.Hex(), user2, -1)

	// Delete the comment
	if err := repo.Delete(ctx, comment.ID.Hex()); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify votes are also deleted
	vote1, _ := repo.GetUserVote(ctx, comment.ID.Hex(), user1)
	vote2, _ := repo.GetUserVote(ctx, comment.ID.Hex(), user2)

	if vote1 != 0 || vote2 != 0 {
		t.Errorf("expected votes to be removed after comment deletion")
	}
}
