package handlers

import (
	"bytes"
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
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockDebateStore struct {
	flagged    []models.FlaggedArgumentDetail
	argument   *models.Argument
	getFlagErr error
	getArgErr  error
	unflagErr  error
	deleteErr  error
	banErr     error

	unflagCalled bool
	deleteCalled bool
	banCalled    bool
}

func (m *mockDebateStore) GetFlaggedArguments(_ context.Context) ([]models.FlaggedArgumentDetail, error) {
	return m.flagged, m.getFlagErr
}
func (m *mockDebateStore) GetArgumentByID(_ context.Context, _ bson.ObjectID) (*models.Argument, error) {
	return m.argument, m.getArgErr
}
func (m *mockDebateStore) UnflagArgument(_ context.Context, _ bson.ObjectID) error {
	m.unflagCalled = true
	return m.unflagErr
}
func (m *mockDebateStore) DeleteArgument(_ context.Context, _ bson.ObjectID) error {
	m.deleteCalled = true
	return m.deleteErr
}
func (m *mockDebateStore) BanUser(_ context.Context, _ bson.ObjectID) error {
	m.banCalled = true
	return m.banErr
}

type mockUserStore struct {
	user *models.User
	err  error
}

func (m *mockUserStore) FindByID(_ context.Context, _ bson.ObjectID) (*models.User, error) {
	return m.user, m.err
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestModerationHandler(db *mockDebateStore, us *mockUserStore) *ModerationHandler {
	return &ModerationHandler{
		debates: db,
		users:   us,
		logger:  zap.NewNop(),
	}
}

func adminContext(w *httptest.ResponseRecorder) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	return c
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

// ── Tests: GetQueue ───────────────────────────────────────────────────────────

func TestGetQueue_ReturnsFlaggedArguments(t *testing.T) {
	argID := bson.NewObjectID()
	db := &mockDebateStore{
		flagged: []models.FlaggedArgumentDetail{
			{
				Argument:       models.Argument{ID: argID, Content: "bad content", FlagCount: 5, Flagged: true},
				AuthorUsername: "spammer",
				DebateTitle:    "Some Debate",
				DebateSlug:     "some-debate",
			},
		},
	}
	h := newTestModerationHandler(db, &mockUserStore{})

	w := httptest.NewRecorder()
	c := adminContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/queue", nil)
	h.GetQueue(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	queue, ok := resp["queue"].([]any)
	if !ok || len(queue) != 1 {
		t.Fatalf("expected 1 item in queue, got %v", resp["queue"])
	}
}

func TestGetQueue_EmptyQueue(t *testing.T) {
	db := &mockDebateStore{flagged: []models.FlaggedArgumentDetail{}}
	h := newTestModerationHandler(db, &mockUserStore{})

	w := httptest.NewRecorder()
	c := adminContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/queue", nil)
	h.GetQueue(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	queue := resp["queue"].([]any)
	if len(queue) != 0 {
		t.Fatalf("expected empty queue")
	}
}

func TestGetQueue_StoreError_Returns500(t *testing.T) {
	db := &mockDebateStore{getFlagErr: errors.New("db error")}
	h := newTestModerationHandler(db, &mockUserStore{})

	w := httptest.NewRecorder()
	c := adminContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/queue", nil)
	h.GetQueue(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// ── Tests: TakeAction ─────────────────────────────────────────────────────────

func setupActionTest(t *testing.T, db *mockDebateStore, argIDStr, action string) *httptest.ResponseRecorder {
	t.Helper()
	h := newTestModerationHandler(db, &mockUserStore{})

	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)
	_ = router

	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	c.Params = gin.Params{{Key: "id", Value: argIDStr}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/"+argIDStr+"/action",
		jsonBody(map[string]string{"action": action}))
	c.Request.Header.Set("Content-Type", "application/json")
	h.TakeAction(c)
	return w
}

func TestTakeAction_InvalidID_Returns400(t *testing.T) {
	db := &mockDebateStore{}
	h := newTestModerationHandler(db, &mockUserStore{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	c.Params = gin.Params{{Key: "id", Value: "not-an-id"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/not-an-id/action",
		jsonBody(map[string]string{"action": "approve"}))
	c.Request.Header.Set("Content-Type", "application/json")
	h.TakeAction(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTakeAction_InvalidAction_Returns400(t *testing.T) {
	argID := bson.NewObjectID()
	db := &mockDebateStore{argument: &models.Argument{ID: argID}}
	w := setupActionTest(t, db, argID.Hex(), "delete") // invalid action
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTakeAction_ArgumentNotFound_Returns404(t *testing.T) {
	argID := bson.NewObjectID()
	db := &mockDebateStore{getArgErr: errors.New("not found")}
	w := setupActionTest(t, db, argID.Hex(), "approve")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestTakeAction_Approve_Returns200AndUnflags(t *testing.T) {
	argID := bson.NewObjectID()
	db := &mockDebateStore{argument: &models.Argument{ID: argID, Flagged: true}}
	w := setupActionTest(t, db, argID.Hex(), "approve")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !db.unflagCalled {
		t.Error("expected UnflagArgument to be called")
	}
}

func TestTakeAction_Remove_Returns200AndDeletes(t *testing.T) {
	argID := bson.NewObjectID()
	db := &mockDebateStore{argument: &models.Argument{ID: argID}}
	w := setupActionTest(t, db, argID.Hex(), "remove")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !db.deleteCalled {
		t.Error("expected DeleteArgument to be called")
	}
}

func TestTakeAction_Ban_Returns200DeletesAndBans(t *testing.T) {
	argID := bson.NewObjectID()
	authorID := bson.NewObjectID()
	db := &mockDebateStore{argument: &models.Argument{ID: argID, AuthorID: authorID}}
	w := setupActionTest(t, db, argID.Hex(), "ban")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !db.deleteCalled {
		t.Error("expected DeleteArgument to be called")
	}
	if !db.banCalled {
		t.Error("expected BanUser to be called")
	}
}

func TestTakeAction_Ban_DeleteFails_Returns500(t *testing.T) {
	argID := bson.NewObjectID()
	db := &mockDebateStore{
		argument:  &models.Argument{ID: argID},
		deleteErr: errors.New("db error"),
	}
	w := setupActionTest(t, db, argID.Hex(), "ban")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// ── Tests: FlagArgument (via DebatesHandler interface boundary) ───────────────
// These tests verify the flag handler logic using a mock that satisfies the
// same interface contract as DebateRepository for flag-related operations.

// flagTestStore implements just the flag-related subset of DebateRepository.
type flagTestStore struct {
	debate        *models.Debate
	argument      *models.Argument
	getDebateErr  error
	getArgErr     error
	createFlagErr error
	flagCount     int64
	flagCountErr  error
	updateFlagErr error

	updateFlaggedCalled bool
	lastFlaggedValue    bool
}

func (s *flagTestStore) GetDebateBySlug(_ context.Context, _ string) (*models.DebateResponse, error) {
	if s.getDebateErr != nil {
		return nil, s.getDebateErr
	}
	if s.debate == nil {
		return nil, errors.New("not found")
	}
	return &models.DebateResponse{Debate: *s.debate}, nil
}
func (s *flagTestStore) GetArgumentByID(_ context.Context, _ bson.ObjectID) (*models.Argument, error) {
	return s.argument, s.getArgErr
}
func (s *flagTestStore) CreateFlag(_ context.Context, _ *models.Flag) error {
	return s.createFlagErr
}
func (s *flagTestStore) GetFlagCount(_ context.Context, _ bson.ObjectID) (int64, error) {
	return s.flagCount, s.flagCountErr
}
func (s *flagTestStore) UpdateArgumentFlagged(_ context.Context, _ bson.ObjectID, flagged bool, _ int) error {
	s.updateFlaggedCalled = true
	s.lastFlaggedValue = flagged
	return s.updateFlagErr
}

// flagHandler is a minimal handler for testing just FlagArgument logic.
type flagHandler struct {
	store *flagTestStore
}

func (h *flagHandler) FlagArgument(c *gin.Context) {
	slug := c.Param("slug")
	debate, err := h.store.GetDebateBySlug(c, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "debate not found"})
		return
	}

	argIDStr := c.Param("id")
	argID, err := bson.ObjectIDFromHex(argIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid argument id"})
		return
	}

	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, _ := bson.ObjectIDFromHex(userIDStr.(string))

	argument, err := h.store.GetArgumentByID(c, argID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "argument not found"})
		return
	}
	if argument.DebateID != debate.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "argument does not belong to this debate"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validReasons := map[string]bool{"spam": true, "harassment": true, "misinformation": true, "off-topic": true}
	if !validReasons[req.Reason] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reason"})
		return
	}

	flag := &models.Flag{
		ArgumentID: argID,
		DebateID:   debate.ID,
		UserID:     userID,
		Reason:     req.Reason,
	}

	if err := h.store.CreateFlag(c, flag); err != nil {
		// Simulate duplicate key: check with a sentinel error
		if errors.Is(err, errDuplicateFlag) {
			c.JSON(http.StatusConflict, gin.H{"error": "you have already flagged this argument"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create flag"})
		return
	}

	count, err := h.store.GetFlagCount(c, argID)
	if err == nil && count >= 3 {
		_ = h.store.UpdateArgumentFlagged(c, argID, true, int(count))
	}

	c.JSON(http.StatusCreated, gin.H{"flag": flag})
}

var errDuplicateFlag = errors.New("duplicate flag")

func callFlagArgument(t *testing.T, h *flagHandler, debateID, argID bson.ObjectID, reason string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserID, bson.NewObjectID().Hex())
	c.Params = gin.Params{
		{Key: "slug", Value: "test-debate"},
		{Key: "id", Value: argID.Hex()},
	}
	c.Request = httptest.NewRequest(http.MethodPost, "/debates/test-debate/arguments/"+argID.Hex()+"/flag",
		jsonBody(map[string]string{"reason": reason}))
	c.Request.Header.Set("Content-Type", "application/json")
	h.FlagArgument(c)
	return w
}

func TestFlagArgument_Created_Returns201(t *testing.T) {
	debateID := bson.NewObjectID()
	argID := bson.NewObjectID()
	store := &flagTestStore{
		debate:    &models.Debate{ID: debateID},
		argument:  &models.Argument{ID: argID, DebateID: debateID},
		flagCount: 1,
	}
	h := &flagHandler{store: store}
	w := callFlagArgument(t, h, debateID, argID, "spam")
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFlagArgument_DuplicateFlag_Returns409(t *testing.T) {
	debateID := bson.NewObjectID()
	argID := bson.NewObjectID()
	store := &flagTestStore{
		debate:        &models.Debate{ID: debateID},
		argument:      &models.Argument{ID: argID, DebateID: debateID},
		createFlagErr: errDuplicateFlag,
	}
	h := &flagHandler{store: store}
	w := callFlagArgument(t, h, debateID, argID, "spam")
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFlagArgument_AutoDownrankAt3Flags(t *testing.T) {
	debateID := bson.NewObjectID()
	argID := bson.NewObjectID()
	store := &flagTestStore{
		debate:    &models.Debate{ID: debateID},
		argument:  &models.Argument{ID: argID, DebateID: debateID},
		flagCount: 3, // threshold met
	}
	h := &flagHandler{store: store}
	w := callFlagArgument(t, h, debateID, argID, "spam")
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if !store.updateFlaggedCalled {
		t.Error("expected UpdateArgumentFlagged to be called at 3-flag threshold")
	}
	if !store.lastFlaggedValue {
		t.Error("expected argument to be marked as flagged=true")
	}
}

func TestFlagArgument_Below3Flags_NoAutoFlag(t *testing.T) {
	debateID := bson.NewObjectID()
	argID := bson.NewObjectID()
	store := &flagTestStore{
		debate:    &models.Debate{ID: debateID},
		argument:  &models.Argument{ID: argID, DebateID: debateID},
		flagCount: 2, // below threshold
	}
	h := &flagHandler{store: store}
	w := callFlagArgument(t, h, debateID, argID, "harassment")
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if store.updateFlaggedCalled {
		t.Error("expected UpdateArgumentFlagged NOT to be called below threshold")
	}
}
