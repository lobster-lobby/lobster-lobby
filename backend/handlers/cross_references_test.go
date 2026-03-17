package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockCrossRefStore struct {
	refs       []*models.CrossReferenceResponse
	created    *models.CrossReferenceResponse
	ref        *models.CrossReference
	createErr  error
	listErr    error
	deleteErr  error
	getByIDErr error

	createCalled bool
	deleteCalled bool
	lastDeleted  bson.ObjectID
}

func (m *mockCrossRefStore) Create(_ context.Context, ref *models.CrossReference) (*models.CrossReferenceResponse, error) {
	m.createCalled = true
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.created != nil {
		return m.created, nil
	}
	return &models.CrossReferenceResponse{
		ID:         ref.ID,
		SourceType: ref.SourceType,
		SourceID:   ref.SourceID,
		TargetType: ref.TargetType,
		TargetID:   ref.TargetID,
		CreatedBy:  ref.CreatedBy,
		CreatedAt:  ref.CreatedAt,
	}, nil
}

func (m *mockCrossRefStore) GetForEntity(_ context.Context, _ string, _ bson.ObjectID) ([]*models.CrossReferenceResponse, error) {
	return m.refs, m.listErr
}

func (m *mockCrossRefStore) Delete(_ context.Context, id bson.ObjectID) error {
	m.deleteCalled = true
	m.lastDeleted = id
	return m.deleteErr
}

func (m *mockCrossRefStore) GetByID(_ context.Context, _ bson.ObjectID) (*models.CrossReference, error) {
	return m.ref, m.getByIDErr
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestCrossRefHandler(store *mockCrossRefStore) *CrossReferenceHandler {
	return &CrossReferenceHandler{
		refs:   store,
		logger: zap.NewNop(),
	}
}

func authedContext(w *httptest.ResponseRecorder) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	return c
}

// ── Tests: Create ─────────────────────────────────────────────────────────────

func TestCreate_Success_Returns201(t *testing.T) {
	store := &mockCrossRefStore{}
	h := newTestCrossRefHandler(store)

	sourceID := bson.NewObjectID()
	targetID := bson.NewObjectID()

	w := httptest.NewRecorder()
	c := authedContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cross-references",
		jsonBody(map[string]string{
			"sourceType": "research",
			"sourceId":   sourceID.Hex(),
			"targetType": "policy",
			"targetId":   targetID.Hex(),
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if !store.createCalled {
		t.Error("expected Create to be called")
	}
}

func TestCreate_DuplicatePrevention_Returns409(t *testing.T) {
	store := &mockCrossRefStore{
		createErr: repository.ErrDuplicateCrossReference,
	}
	h := newTestCrossRefHandler(store)

	sourceID := bson.NewObjectID()
	targetID := bson.NewObjectID()

	w := httptest.NewRecorder()
	c := authedContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cross-references",
		jsonBody(map[string]string{
			"sourceType": "research",
			"sourceId":   sourceID.Hex(),
			"targetType": "policy",
			"targetId":   targetID.Hex(),
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreate_InvalidSourceType_Returns400(t *testing.T) {
	store := &mockCrossRefStore{}
	h := newTestCrossRefHandler(store)

	w := httptest.NewRecorder()
	c := authedContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cross-references",
		jsonBody(map[string]string{
			"sourceType": "invalid",
			"sourceId":   bson.NewObjectID().Hex(),
			"targetType": "policy",
			"targetId":   bson.NewObjectID().Hex(),
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreate_SelfReference_Returns400(t *testing.T) {
	store := &mockCrossRefStore{}
	h := newTestCrossRefHandler(store)

	id := bson.NewObjectID()

	w := httptest.NewRecorder()
	c := authedContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cross-references",
		jsonBody(map[string]string{
			"sourceType": "policy",
			"sourceId":   id.Hex(),
			"targetType": "policy",
			"targetId":   id.Hex(),
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ── Tests: Bidirectional Query ────────────────────────────────────────────────

func TestList_ReturnsBidirectionalRefs(t *testing.T) {
	policyID := bson.NewObjectID()
	researchID := bson.NewObjectID()
	debateID := bson.NewObjectID()

	store := &mockCrossRefStore{
		refs: []*models.CrossReferenceResponse{
			{
				ID:          bson.NewObjectID(),
				SourceType:  "policy",
				SourceID:    policyID,
				SourceTitle: "Test Policy",
				TargetType:  "research",
				TargetID:    researchID,
				TargetTitle: "Test Research",
			},
			{
				ID:          bson.NewObjectID(),
				SourceType:  "debate",
				SourceID:    debateID,
				SourceTitle: "Test Debate",
				TargetType:  "policy",
				TargetID:    policyID,
				TargetTitle: "Test Policy",
			},
		},
	}
	h := newTestCrossRefHandler(store)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/cross-references?type=policy&id="+policyID.Hex(), nil)
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	refs, ok := resp["references"].([]any)
	if !ok || len(refs) != 2 {
		t.Fatalf("expected 2 references, got %v", resp["references"])
	}
}

func TestList_EmptyResult(t *testing.T) {
	store := &mockCrossRefStore{refs: nil}
	h := newTestCrossRefHandler(store)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/cross-references?type=policy&id="+bson.NewObjectID().Hex(), nil)
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	refs := resp["references"].([]any)
	if len(refs) != 0 {
		t.Fatalf("expected empty references array")
	}
}

func TestList_InvalidType_Returns400(t *testing.T) {
	store := &mockCrossRefStore{}
	h := newTestCrossRefHandler(store)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/cross-references?type=invalid&id="+bson.NewObjectID().Hex(), nil)
	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ── Tests: Delete ─────────────────────────────────────────────────────────────

func TestDelete_Success_Returns200(t *testing.T) {
	store := &mockCrossRefStore{}
	h := newTestCrossRefHandler(store)

	refID := bson.NewObjectID()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	c.Params = gin.Params{{Key: "id", Value: refID.Hex()}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/cross-references/"+refID.Hex(), nil)
	h.Delete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !store.deleteCalled {
		t.Error("expected Delete to be called")
	}
	if store.lastDeleted != refID {
		t.Errorf("expected deleted ID %s, got %s", refID.Hex(), store.lastDeleted.Hex())
	}
}

func TestDelete_NotFound_Returns404(t *testing.T) {
	store := &mockCrossRefStore{deleteErr: errors.New("not found")}
	h := newTestCrossRefHandler(store)

	refID := bson.NewObjectID()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	c.Params = gin.Params{{Key: "id", Value: refID.Hex()}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/cross-references/"+refID.Hex(), nil)
	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDelete_InvalidID_Returns400(t *testing.T) {
	store := &mockCrossRefStore{}
	h := newTestCrossRefHandler(store)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	c.Params = gin.Params{{Key: "id", Value: "not-valid"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/cross-references/not-valid", nil)
	h.Delete(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// (duplicate key error is now repository.ErrDuplicateCrossReference)
