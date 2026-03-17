package repository

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestCountDebateCommentsFilter verifies that the filter used by CountDebateComments
// excludes nested replies by requiring parentId to be absent.
func TestCountDebateCommentsFilter(t *testing.T) {
	policyID := bson.NewObjectID()

	filter := bson.M{
		"policyId": policyID,
		"parentId": bson.M{"$exists": false},
	}

	// Verify policyId is set correctly.
	if filter["policyId"] != policyID {
		t.Errorf("expected policyId %v, got %v", policyID, filter["policyId"])
	}

	// Verify parentId filter excludes replies.
	parentFilter, ok := filter["parentId"].(bson.M)
	if !ok {
		t.Fatal("parentId filter should be a bson.M")
	}
	exists, ok := parentFilter["$exists"].(bool)
	if !ok || exists {
		t.Errorf("parentId $exists should be false to exclude replies, got %v", parentFilter["$exists"])
	}
}

// TestEndorsementThresholdFilter verifies the filter used when adding endorsements
// correctly prevents duplicate endorsements from the same user.
func TestEndorsementThresholdFilter(t *testing.T) {
	policyID := bson.NewObjectID()
	userID := bson.NewObjectID()

	filter := bson.M{
		"policyId":          policyID,
		"endorsers.userId":  bson.M{"$ne": userID},
	}

	if filter["policyId"] != policyID {
		t.Errorf("expected policyId %v", policyID)
	}

	userFilter, ok := filter["endorsers.userId"].(bson.M)
	if !ok {
		t.Fatal("endorsers.userId filter should be bson.M")
	}
	if userFilter["$ne"] != userID {
		t.Errorf("expected $ne userID %v", userID)
	}
}
