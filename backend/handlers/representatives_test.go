package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockCivicServer creates a test HTTP server for the Civic API.
func mockCivicServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}

// newTestRepHandler builds a RepresentativeHandler backed by a real service
// pointed at the given mock civic server URL (empty apiKey = no-key path).
func newTestRepHandler(apiKey string) *RepresentativeHandler {
	svc := services.NewRepresentativeService(nil, apiKey)
	return NewRepresentativeHandler(svc)
}

func callList(t *testing.T, handler *RepresentativeHandler, query string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/representatives?"+query, nil)
	// Populate gin query params from the request URL
	c.Request.URL.RawQuery = query
	handler.List(c)
	return w
}

// TestList_NoAPIKey verifies that when no API key is configured the handler
// returns HTTP 200 with an empty officials array (graceful degradation).
func TestList_NoAPIKey(t *testing.T) {
	handler := newTestRepHandler("") // no key
	w := callList(t, handler, "address=123+Main+St")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	officials, ok := body["officials"]
	if !ok {
		t.Fatal("expected 'officials' key in response")
	}
	list, ok := officials.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", officials)
	}
	if len(list) != 0 {
		t.Errorf("expected empty officials, got %d items", len(list))
	}
}

// TestList_NoQueryParams verifies an empty-param request returns an empty result.
func TestList_NoQueryParams(t *testing.T) {
	handler := newTestRepHandler("any-key")
	w := callList(t, handler, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	officials := body["officials"].([]interface{})
	if len(officials) != 0 {
		t.Errorf("expected empty officials array")
	}
}

// TestList_Non200CivicResponse verifies that a non-200 from the Civic API
// results in a graceful HTTP 200 with an empty officials array.
func TestList_Non200CivicResponse(t *testing.T) {
	// Start a mock server that returns 403 (e.g., bad API key or quota exceeded)
	mock := mockCivicServer(http.StatusForbidden, `{"error":{"code":403,"message":"forbidden"}}`)
	defer mock.Close()

	// Build a service that uses the mock server URL as the base
	svc := services.NewRepresentativeServiceWithBaseURL(nil, "test-key", mock.URL)
	handler := NewRepresentativeHandler(svc)

	w := callList(t, handler, "address=1600+Pennsylvania+Ave")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (graceful), got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	officials, ok := body["officials"]
	if !ok {
		t.Fatal("expected 'officials' key")
	}
	list := officials.([]interface{})
	if len(list) != 0 {
		t.Errorf("expected empty officials on non-200, got %d", len(list))
	}
}

// TestList_Non200_500 verifies a 500 from Civic API is also handled gracefully.
func TestList_Non200_500(t *testing.T) {
	mock := mockCivicServer(http.StatusInternalServerError, `internal error`)
	defer mock.Close()

	svc := services.NewRepresentativeServiceWithBaseURL(nil, "test-key", mock.URL)
	handler := NewRepresentativeHandler(svc)

	w := callList(t, handler, "address=anywhere")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (graceful), got %d", w.Code)
	}
}

// TestList_ValidCivicResponse verifies officials are returned on a 200 mock.
func TestList_ValidCivicResponse(t *testing.T) {
	body := `{
		"offices": [{"name":"Mayor","divisionId":"ocd-division/country:us","officialIndices":[0]}],
		"officials": [{"name":"Jane Smith","party":"Democratic Party","phones":["555-0100"]}]
	}`
	mock := mockCivicServer(http.StatusOK, body)
	defer mock.Close()

	svc := services.NewRepresentativeServiceWithBaseURL(nil, "test-key", mock.URL)
	handler := NewRepresentativeHandler(svc)

	w := callList(t, handler, "address=1+Main+St")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Officials []models.CivicOfficial `json:"officials"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Officials) != 1 {
		t.Fatalf("expected 1 official, got %d", len(resp.Officials))
	}
	if resp.Officials[0].Name != "Jane Smith" {
		t.Errorf("unexpected name: %s", resp.Officials[0].Name)
	}
}

// Compile-time check that context is usable (handler uses gin.Context which embeds context.Context).
var _ context.Context = (*gin.Context)(nil)
