package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- Mock RepresentativeStore ---

type mockRepStore struct {
	reps      []models.Representative
	rep       *models.Representative
	total     int64
	createErr error
	findErr   error
	updateErr error
	deleteErr error
	listErr   error

	createdRep *models.Representative
	updatedID  bson.ObjectID
	updatedM   bson.M
	deletedID  bson.ObjectID
}

func (m *mockRepStore) Create(_ context.Context, rep *models.Representative) error {
	m.createdRep = rep
	if rep.ID.IsZero() {
		rep.ID = bson.NewObjectID()
	}
	return m.createErr
}

func (m *mockRepStore) FindByID(_ context.Context, id bson.ObjectID) (*models.Representative, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.rep, nil
}

func (m *mockRepStore) Update(_ context.Context, id bson.ObjectID, updates bson.M) error {
	m.updatedID = id
	m.updatedM = updates
	return m.updateErr
}

func (m *mockRepStore) Delete(_ context.Context, id bson.ObjectID) error {
	m.deletedID = id
	return m.deleteErr
}

func (m *mockRepStore) List(_ context.Context, opts repository.RepListOpts) ([]models.Representative, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.reps, m.total, nil
}

// --- Mock VotingRecordStore ---

type mockVoteStore struct {
	records   []models.VotingRecord
	total     int64
	summary   *models.VotingSummary
	createErr error
	findErr   error
	summErr   error

	createdVR *models.VotingRecord
}

func (m *mockVoteStore) Create(_ context.Context, vr *models.VotingRecord) error {
	m.createdVR = vr
	if vr.ID.IsZero() {
		vr.ID = bson.NewObjectID()
	}
	return m.createErr
}

func (m *mockVoteStore) FindByRepresentative(_ context.Context, _ bson.ObjectID, _ repository.VoteListOpts) ([]models.VotingRecord, int64, error) {
	if m.findErr != nil {
		return nil, 0, m.findErr
	}
	return m.records, m.total, nil
}

func (m *mockVoteStore) GetSummary(_ context.Context, _ bson.ObjectID) (*models.VotingSummary, error) {
	if m.summErr != nil {
		return nil, m.summErr
	}
	return m.summary, nil
}

// --- Mock CivicLookupService ---

type mockCivicSvc struct {
	officials []models.CivicOfficial
	err       error
}

func (m *mockCivicSvc) LookupByAddress(_ context.Context, _ string) ([]models.CivicOfficial, error) {
	return m.officials, m.err
}

// --- Helper ---

func newRepHandler(reps *mockRepStore, votes *mockVoteStore, civic CivicLookupService) *RepresentativeHandler {
	return NewRepresentativeHandler(reps, votes, civic, zap.NewNop())
}

// --- Tests ---

func TestRepList_DBListing(t *testing.T) {
	repID := bson.NewObjectID()
	store := &mockRepStore{
		reps: []models.Representative{
			{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
		},
		total: 1,
	}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives?party=Democratic&state=CA&page=1", nil)

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Representatives []models.Representative `json:"representatives"`
		Total           int64                    `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Representatives) != 1 {
		t.Errorf("expected 1 rep, got %d", len(resp.Representatives))
	}
	if resp.Total != 1 {
		t.Errorf("expected total=1, got %d", resp.Total)
	}
}

func TestRepList_AddressLookup(t *testing.T) {
	civic := &mockCivicSvc{
		officials: []models.CivicOfficial{
			{Name: "John Smith", Title: "Mayor", Party: "Republican"},
		},
	}
	h := newRepHandler(&mockRepStore{}, &mockVoteStore{}, civic)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives?address=123+Main+St", nil)

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Officials []models.CivicOfficial `json:"officials"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Officials) != 1 {
		t.Fatalf("expected 1 official, got %d", len(resp.Officials))
	}
	if resp.Officials[0].Name != "John Smith" {
		t.Errorf("unexpected name: %s", resp.Officials[0].Name)
	}
}

func TestRepList_AddressLookupNoCivic(t *testing.T) {
	h := newRepHandler(&mockRepStore{}, &mockVoteStore{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives?address=123+Main+St", nil)

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRepGetByID_Success(t *testing.T) {
	repID := bson.NewObjectID()
	store := &mockRepStore{
		rep: &models.Representative{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
	}
	votes := &mockVoteStore{
		summary: &models.VotingSummary{TotalVotes: 10, YeaCount: 7, NayCount: 2, AbstainCount: 1, YeaPercent: 70, NayPercent: 20, AbstainPercent: 10},
	}
	h := newRepHandler(store, votes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives/"+repID.Hex(), nil)
	c.Params = gin.Params{{Key: "id", Value: repID.Hex()}}

	h.GetByID(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Representative models.Representative `json:"representative"`
		VotingSummary  models.VotingSummary   `json:"votingSummary"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Representative.Name != "Jane Doe" {
		t.Errorf("unexpected name: %s", resp.Representative.Name)
	}
	if resp.VotingSummary.TotalVotes != 10 {
		t.Errorf("expected 10 total votes, got %d", resp.VotingSummary.TotalVotes)
	}
}

func TestRepGetByID_NotFound(t *testing.T) {
	store := &mockRepStore{rep: nil}
	h := newRepHandler(store, &mockVoteStore{summary: &models.VotingSummary{}}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	id := bson.NewObjectID()
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives/"+id.Hex(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.Hex()}}

	h.GetByID(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRepGetByID_InvalidID(t *testing.T) {
	h := newRepHandler(&mockRepStore{}, &mockVoteStore{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	h.GetByID(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRepCreate_Success(t *testing.T) {
	store := &mockRepStore{}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	body := `{"name":"John Doe","title":"Senator","party":"Republican","state":"TX","district":"","chamber":"senate","level":"federal"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/representatives", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", w.Code, w.Body.String())
	}
	if store.createdRep == nil {
		t.Fatal("expected rep to be created")
	}
	if store.createdRep.Name != "John Doe" {
		t.Errorf("unexpected name: %s", store.createdRep.Name)
	}
}

func TestRepCreate_ValidationError(t *testing.T) {
	h := newRepHandler(&mockRepStore{}, &mockVoteStore{}, nil)

	body := `{"name":"X","party":"","state":"","chamber":""}` // name too short, party/state/chamber empty
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/representatives", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestRepUpdate_Success(t *testing.T) {
	repID := bson.NewObjectID()
	store := &mockRepStore{
		rep: &models.Representative{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
	}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	body := `{"name":"Jane Smith","bio":"Updated bio"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/representatives/"+repID.Hex(), bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: repID.Hex()}}

	h.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if store.updatedM["name"] != "Jane Smith" {
		t.Errorf("expected name update, got %v", store.updatedM)
	}
	if store.updatedM["bio"] != "Updated bio" {
		t.Errorf("expected bio update, got %v", store.updatedM)
	}
}

func TestRepUpdate_NotFound(t *testing.T) {
	store := &mockRepStore{rep: nil}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	body := `{"name":"Jane Smith"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	id := bson.NewObjectID()
	c.Request = httptest.NewRequest(http.MethodPut, "/api/representatives/"+id.Hex(), bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id.Hex()}}

	h.Update(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRepDelete_Success(t *testing.T) {
	repID := bson.NewObjectID()
	store := &mockRepStore{
		rep: &models.Representative{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
	}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/representatives/"+repID.Hex(), nil)
	c.Params = gin.Params{{Key: "id", Value: repID.Hex()}}

	h.Delete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if store.deletedID != repID {
		t.Errorf("expected delete for %s, got %s", repID.Hex(), store.deletedID.Hex())
	}
}

func TestRepDelete_NotFound(t *testing.T) {
	store := &mockRepStore{rep: nil}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	id := bson.NewObjectID()
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/representatives/"+id.Hex(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.Hex()}}

	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRepListVotes_Success(t *testing.T) {
	repID := bson.NewObjectID()
	repStore := &mockRepStore{
		rep: &models.Representative{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
	}
	voteStore := &mockVoteStore{
		records: []models.VotingRecord{
			{ID: bson.NewObjectID(), RepresentativeID: repID, PolicyID: bson.NewObjectID(), Vote: models.VoteYea, Date: time.Now(), Session: "118th"},
		},
		total: 1,
	}
	h := newRepHandler(repStore, voteStore, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives/"+repID.Hex()+"/votes?page=1&perPage=20", nil)
	c.Params = gin.Params{{Key: "id", Value: repID.Hex()}}

	h.ListVotes(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Votes []models.VotingRecord `json:"votes"`
		Total int64                 `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Votes) != 1 {
		t.Errorf("expected 1 vote, got %d", len(resp.Votes))
	}
}

func TestRepListVotes_RepNotFound(t *testing.T) {
	store := &mockRepStore{rep: nil}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	id := bson.NewObjectID()
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives/"+id.Hex()+"/votes", nil)
	c.Params = gin.Params{{Key: "id", Value: id.Hex()}}

	h.ListVotes(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRepRecordVote_Success(t *testing.T) {
	repID := bson.NewObjectID()
	policyID := bson.NewObjectID()
	repStore := &mockRepStore{
		rep: &models.Representative{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
	}
	voteStore := &mockVoteStore{}
	h := newRepHandler(repStore, voteStore, nil)

	body, _ := json.Marshal(map[string]string{
		"policyId": policyID.Hex(),
		"vote":     "yea",
		"date":     "2025-03-15",
		"session":  "118th Congress",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/representatives/"+repID.Hex()+"/votes", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: repID.Hex()}}

	h.RecordVote(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", w.Code, w.Body.String())
	}
	if voteStore.createdVR == nil {
		t.Fatal("expected vote to be created")
	}
	if voteStore.createdVR.Vote != models.VoteYea {
		t.Errorf("expected yea, got %s", voteStore.createdVR.Vote)
	}
}

func TestRepRecordVote_InvalidDate(t *testing.T) {
	repID := bson.NewObjectID()
	repStore := &mockRepStore{
		rep: &models.Representative{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
	}
	h := newRepHandler(repStore, &mockVoteStore{}, nil)

	body := `{"policyId":"` + bson.NewObjectID().Hex() + `","vote":"yea","date":"not-a-date","session":"118th"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/representatives/"+repID.Hex()+"/votes", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: repID.Hex()}}

	h.RecordVote(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRepRecordVote_InvalidVote(t *testing.T) {
	repID := bson.NewObjectID()
	repStore := &mockRepStore{
		rep: &models.Representative{ID: repID, Name: "Jane Doe", Party: "Democratic", State: "CA", Chamber: "senate"},
	}
	h := newRepHandler(repStore, &mockVoteStore{}, nil)

	body := `{"policyId":"` + bson.NewObjectID().Hex() + `","vote":"invalid","date":"2025-03-15","session":"118th"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/representatives/"+repID.Hex()+"/votes", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: repID.Hex()}}

	h.RecordVote(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestRepRecordVote_RepNotFound(t *testing.T) {
	store := &mockRepStore{rep: nil}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	body := `{"policyId":"` + bson.NewObjectID().Hex() + `","vote":"yea","date":"2025-03-15","session":"118th"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	id := bson.NewObjectID()
	c.Request = httptest.NewRequest(http.MethodPost, "/api/representatives/"+id.Hex()+"/votes", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id.Hex()}}

	h.RecordVote(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRepList_StoreError(t *testing.T) {
	store := &mockRepStore{listErr: errors.New("db error")}
	h := newRepHandler(store, &mockVoteStore{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives", nil)

	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
