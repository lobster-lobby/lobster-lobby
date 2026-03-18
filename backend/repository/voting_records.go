package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/lobster-lobby/lobster-lobby/models"
)

// VotingRecordRepository handles database operations for voting records
type VotingRecordRepository struct {
	coll *mongo.Collection
}

func NewVotingRecordRepository(db *MongoDB) *VotingRecordRepository {
	return &VotingRecordRepository{coll: db.Database.Collection("voting_records")}
}

func (r *VotingRecordRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "representativeId", Value: 1}, {Key: "date", Value: -1}}},
		{Keys: bson.D{{Key: "policyId", Value: 1}}},
		{Keys: bson.D{
			{Key: "representativeId", Value: 1},
			{Key: "policyId", Value: 1},
		}, Options: options.Index().SetUnique(true)},
	})
	return err
}

func (r *VotingRecordRepository) Create(ctx context.Context, vr *models.VotingRecord) error {
	now := time.Now().UTC()
	vr.CreatedAt = now
	vr.UpdatedAt = now
	if vr.ID.IsZero() {
		vr.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, vr)
	return err
}

// VoteListOpts configures pagination for listing voting records
type VoteListOpts struct {
	Page    int
	PerPage int
}

func (r *VotingRecordRepository) FindByRepresentative(ctx context.Context, repID bson.ObjectID, opts VoteListOpts) ([]models.VotingRecord, int64, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 || opts.PerPage > 100 {
		opts.PerPage = 20
	}

	filter := bson.M{"representativeId": repID}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.M{"date": -1}).
		SetSkip(int64((opts.Page - 1) * opts.PerPage)).
		SetLimit(int64(opts.PerPage))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []models.VotingRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}
	if records == nil {
		records = []models.VotingRecord{}
	}
	return records, total, nil
}

func (r *VotingRecordRepository) FindByPolicy(ctx context.Context, policyID bson.ObjectID, opts VoteListOpts) ([]models.VotingRecord, int64, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 || opts.PerPage > 100 {
		opts.PerPage = 20
	}

	filter := bson.M{"policyId": policyID}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.M{"date": -1}).
		SetSkip(int64((opts.Page - 1) * opts.PerPage)).
		SetLimit(int64(opts.PerPage))

	cursor, err := r.coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []models.VotingRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}
	if records == nil {
		records = []models.VotingRecord{}
	}
	return records, total, nil
}

func (r *VotingRecordRepository) GetPolicySummary(ctx context.Context, policyID bson.ObjectID) (*models.VotingSummary, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"policyId": policyID}},
		bson.M{"$group": bson.M{
			"_id":          nil,
			"totalVotes":   bson.M{"$sum": 1},
			"yeaCount":     bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "yea"}}, 1, 0}}},
			"nayCount":     bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "nay"}}, 1, 0}}},
			"abstainCount": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "abstain"}}, 1, 0}}},
			"absentCount":  bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "absent"}}, 1, 0}}},
		}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalVotes   int `bson:"totalVotes"`
		YeaCount     int `bson:"yeaCount"`
		NayCount     int `bson:"nayCount"`
		AbstainCount int `bson:"abstainCount"`
		AbsentCount  int `bson:"absentCount"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	summary := &models.VotingSummary{}
	if len(results) > 0 {
		res := results[0]
		summary.TotalVotes = res.TotalVotes
		summary.YeaCount = res.YeaCount
		summary.NayCount = res.NayCount
		summary.AbstainCount = res.AbstainCount
		summary.AbsentCount = res.AbsentCount
		if summary.TotalVotes > 0 {
			total := float64(summary.TotalVotes)
			summary.YeaPercent = float64(summary.YeaCount) / total * 100
			summary.NayPercent = float64(summary.NayCount) / total * 100
			summary.AbstainPercent = float64(summary.AbstainCount) / total * 100
		}
	}
	return summary, nil
}

func (r *VotingRecordRepository) GetSummary(ctx context.Context, repID bson.ObjectID) (*models.VotingSummary, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"representativeId": repID}},
		bson.M{"$group": bson.M{
			"_id":          nil,
			"totalVotes":   bson.M{"$sum": 1},
			"yeaCount":     bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "yea"}}, 1, 0}}},
			"nayCount":     bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "nay"}}, 1, 0}}},
			"abstainCount": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "abstain"}}, 1, 0}}},
			"absentCount":  bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$vote", "absent"}}, 1, 0}}},
		}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalVotes   int `bson:"totalVotes"`
		YeaCount     int `bson:"yeaCount"`
		NayCount     int `bson:"nayCount"`
		AbstainCount int `bson:"abstainCount"`
		AbsentCount  int `bson:"absentCount"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	summary := &models.VotingSummary{}
	if len(results) > 0 {
		res := results[0]
		summary.TotalVotes = res.TotalVotes
		summary.YeaCount = res.YeaCount
		summary.NayCount = res.NayCount
		summary.AbstainCount = res.AbstainCount
		summary.AbsentCount = res.AbsentCount
		if summary.TotalVotes > 0 {
			total := float64(summary.TotalVotes)
			summary.YeaPercent = float64(summary.YeaCount) / total * 100
			summary.NayPercent = float64(summary.NayCount) / total * 100
			summary.AbstainPercent = float64(summary.AbstainCount) / total * 100
		}
	}
	return summary, nil
}
