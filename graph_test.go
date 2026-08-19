package memmesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// capture records the one request a stub server saw, so tests can assert on
// route, query, and body without a live API.
type capture struct {
	method string
	path   string
	query  string
	body   map[string]any
}

func stub(t *testing.T, response string) (*Client, *capture) {
	t.Helper()
	got := &capture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.RawQuery
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&got.body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	return New("sk-test", "proj_1", WithBaseURL(srv.URL)), got
}

// ── graph ───────────────────────────────────────────────────────────────────

func TestGraphStatsParsesCounts(t *testing.T) {
	c, got := stub(t, `{"entityCount":12142,"edgeCount":287698,"memoriesWithEdges":184737,
		"retiredEntities":0,"retiredEdges":2,"entitiesByType":{"concept":5463},
		"extraction":{"platformEnabled":true,"projectEnabled":false}}`)

	st, err := c.Graph.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.EntityCount != 12142 || st.EdgeCount != 287698 || st.MemoriesWithEdges != 184737 {
		t.Fatalf("counts: %+v", st)
	}
	if st.EntitiesByType["concept"] != 5463 {
		t.Fatalf("byType: %+v", st.EntitiesByType)
	}
	if st.Extraction == nil || st.Extraction.ProjectEnabled {
		t.Fatalf("extraction: %+v", st.Extraction)
	}
	if got.path != "/api/v1/projects/proj_1/admin/memory/graph/stats" {
		t.Fatalf("path: %s", got.path)
	}
}

func TestGraphListEntitiesSendsOnlySetFilters(t *testing.T) {
	c, got := stub(t, `[]`)
	if _, err := c.Graph.ListEntities(context.Background(),
		ListEntitiesParams{Search: "Sarah", Limit: 5}); err != nil {
		t.Fatal(err)
	}
	q := got.query
	if !contains(q, "search=Sarah") || !contains(q, "limit=5") {
		t.Fatalf("query missing filters: %s", q)
	}
	if contains(q, "scope=") || contains(q, "offset=") {
		t.Fatalf("query carries unset filters: %s", q)
	}
}

func TestGraphListEntitiesWithoutFiltersSendsNoQuery(t *testing.T) {
	c, got := stub(t, `[]`)
	if _, err := c.Graph.ListEntities(context.Background(), ListEntitiesParams{}); err != nil {
		t.Fatal(err)
	}
	if got.query != "" {
		t.Fatalf("expected no query, got %q", got.query)
	}
}

func TestGraphListEntitiesEscapesFilterValues(t *testing.T) {
	// An unescaped & would truncate the filter server-side and quietly return
	// the wrong page — a correctness test, not a style one.
	c, got := stub(t, `[]`)
	if _, err := c.Graph.ListEntities(context.Background(),
		ListEntitiesParams{Search: "a&b c"}); err != nil {
		t.Fatal(err)
	}
	if !contains(got.query, "search=a%26b+c") {
		t.Fatalf("value not escaped: %s", got.query)
	}
}

func TestGraphListEdgesDecodesHydratedShape(t *testing.T) {
	// Regression: the read routes return GraphTraversalEdge, not the raw
	// memory_edge row. There is no subjectId on the wire at all.
	c, _ := stub(t, `[{"id":"g1",
		"subject":{"id":"e1","canonicalName":"NVIDIA CORP","type":"org"},
		"predicate":"reported_metric",
		"object":{"id":"e2","canonicalName":"Cost of Revenue"},
		"objectLiteral":null,"weight":0.85,"sourceMemoryId":"m1","hop":0}]`)

	edges, err := c.Graph.ListEdges(context.Background(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if edges[0].Subject.CanonicalName != "NVIDIA CORP" {
		t.Fatalf("subject: %+v", edges[0].Subject)
	}
	if edges[0].Object == nil || edges[0].Object.CanonicalName != "Cost of Revenue" {
		t.Fatalf("object: %+v", edges[0].Object)
	}
	if edges[0].Hop != 0 || edges[0].Weight != 0.85 {
		t.Fatalf("hop/weight: %+v", edges[0])
	}
}

func TestGraphListEdgesDecodesLiteralObject(t *testing.T) {
	// Object is null when the value is a literal rather than an entity.
	c, _ := stub(t, `[{"id":"g2","subject":{"id":"e1","canonicalName":"NVIDIA CORP"},
		"predicate":"ticker_symbol","object":null,"objectLiteral":"NVDA","weight":0.85,"hop":0}]`)

	edges, err := c.Graph.ListEdges(context.Background(), "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if edges[0].Object != nil {
		t.Fatalf("expected nil object, got %+v", edges[0].Object)
	}
	if edges[0].ObjectLiteral != "NVDA" {
		t.Fatalf("literal: %q", edges[0].ObjectLiteral)
	}
}

func TestGraphTraversePostsEntityIDAndOmitsUnset(t *testing.T) {
	c, got := stub(t, `[]`)
	if _, err := c.Graph.Traverse(context.Background(), "e1",
		TraverseParams{Hops: 2, Predicates: []string{"member_of", "led_by"}}); err != nil {
		t.Fatal(err)
	}
	if got.method != "POST" || got.path != "/api/v1/projects/proj_1/admin/memory/graph/traverse" {
		t.Fatalf("route: %s %s", got.method, got.path)
	}
	if got.body["entityId"] != "e1" {
		t.Fatalf("entityId: %v", got.body["entityId"])
	}
	if got.body["hops"] != float64(2) {
		t.Fatalf("hops: %v", got.body["hops"])
	}
	if _, ok := got.body["asOf"]; ok {
		t.Fatalf("asOf should be omitted: %v", got.body)
	}
}

func TestGraphGetEntityReturnsHydratedEdges(t *testing.T) {
	c, _ := stub(t, `{"entity":{"id":"e1","canonicalName":"Sarah"},
		"edges":[{"id":"g1","subject":{"id":"e1","canonicalName":"Sarah"},
		"predicate":"works_at","object":{"id":"e2","canonicalName":"Acme"},"weight":0.9,"hop":1}]}`)

	hood, err := c.Graph.GetEntity(context.Background(), "e1", "")
	if err != nil {
		t.Fatal(err)
	}
	if hood.Entity == nil || hood.Entity.CanonicalName != "Sarah" {
		t.Fatalf("entity: %+v", hood.Entity)
	}
	if hood.Edges[0].Hop != 1 {
		t.Fatalf("hop: %+v", hood.Edges[0])
	}
}

// ── observe ─────────────────────────────────────────────────────────────────

func TestObserveTextRoutesThroughEngineAndCarriesIdentity(t *testing.T) {
	c, got := stub(t, `{"saved":[{"id":"m1","type":"fact","content":"x"}],"candidateCount":3}`)

	res, err := c.Memory.Observe(context.Background(), Observe{
		Text:      "I just moved to Denver.",
		UserID:    "user-123",
		AgentID:   "agent-9",
		SessionID: "thread-456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.path != "/api/v1/projects/proj_1/memory/observe" {
		t.Fatalf("path: %s", got.path)
	}
	if got.body["text"] != "I just moved to Denver." || got.body["role"] != "user" {
		t.Fatalf("body: %v", got.body)
	}
	if got.body["userId"] != "user-123" || got.body["agentId"] != "agent-9" ||
		got.body["sessionId"] != "thread-456" {
		t.Fatalf("identity not forwarded: %v", got.body)
	}
	if len(res.Saved) != 1 || res.CandidateCount != 3 {
		t.Fatalf("response: %+v", res)
	}
}

func TestObserveTextOmitsIdentityWhenUnset(t *testing.T) {
	c, got := stub(t, `{"saved":[],"candidateCount":0}`)
	if _, err := c.Memory.Observe(context.Background(), Observe{Text: "hello"}); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"userId", "agentId", "sessionId"} {
		if _, ok := got.body[k]; ok {
			t.Fatalf("%s should be omitted: %v", k, got.body)
		}
	}
}

func TestObserveFillerReturnsEmptySavedAsSuccess(t *testing.T) {
	// An empty Saved is the engine working, not an error.
	c, _ := stub(t, `{"saved":[],"candidateCount":0}`)
	res, err := c.Memory.Observe(context.Background(), Observe{Text: "ok thanks"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Saved) != 0 || res.CandidateCount != 0 {
		t.Fatalf("response: %+v", res)
	}
}

func TestObserveLegacyContentStillHitsAdminMemory(t *testing.T) {
	// The verbatim path must keep working, wrapped in the new response shape.
	c, got := stub(t, `{"id":"m1","type":"event","content":"Ordered pizza"}`)
	res, err := c.Memory.Observe(context.Background(), Observe{
		Subject: Subject{Kind: "contact", ExternalID: "sarah"},
		Content: "Ordered pizza",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.path != "/api/v1/projects/proj_1/admin/memory" {
		t.Fatalf("path: %s", got.path)
	}
	if len(res.Saved) != 1 || res.CandidateCount != 1 {
		t.Fatalf("response: %+v", res)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle ||
		len(needle) > 0 && indexOf(haystack, needle) >= 0)
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}
