package memmesh

import (
	"context"
	"net/url"
	"strconv"
)

// GraphService reads the knowledge graph built from observed memory — the
// structural half of what MemMesh stores.
//
// Observing text doesn't only produce embeddable rows; extraction also resolves
// entities and writes typed edges between them. That graph is what reaches a
// fact no single memory states outright ("who does Sarah report to?" answered
// from sarah -[member_of]-> team plus team -[led_by]-> priya).
//
// Every route here is admin-tier (/admin/memory/...); a project-scoped key gets
// a 403.
//
// Read-only by design. Entities and edges are written by extraction when you
// call Observe; the server's manual create/retire routes exist for annotation
// tooling, and exposing them here would invite hand-maintained graphs — the
// work the engine exists to do for you.
type GraphService struct{ c *Client }

// MemoryEntity is a resolved thing — person, org, product, concept — filed
// under CanonicalName, with Aliases resolving to it.
type MemoryEntity struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId,omitempty"`
	// BrainID is the brain that FIRST created this entity. Entities dedupe per
	// project, so this is provenance, not an isolation key — brain-scoped graph
	// work filters on the edge's brain, which the read routes apply server-side.
	BrainID        string         `json:"brainId,omitempty"`
	Scope          string         `json:"scope,omitempty"`
	Type           string         `json:"type,omitempty"`
	CanonicalName  string         `json:"canonicalName"`
	Aliases        []string       `json:"aliases,omitempty"`
	Description    string         `json:"description,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	ValidFrom      string         `json:"validFrom,omitempty"`
	ValidTo        string         `json:"validTo,omitempty"` // empty while current
	SupersededByID string         `json:"supersededById,omitempty"`
}

// GraphTraversalEdge is an edge as the READ routes return it — hydrated, not
// the raw memory_edge row. Subject and Object are resolved entities rather than
// ids, plus a Hop counter.
//
// This is the server's GraphTraversalEdge, returned by ListEdges, Traverse, and
// the Edges of GetEntity. The raw row shape (subjectId / objectId) is not
// exposed by any read route, so it is deliberately not modelled — a type
// nothing returns is a trap.
type GraphTraversalEdge struct {
	ID      string       `json:"id"`
	Subject MemoryEntity `json:"subject"`
	// Predicate is the relationship — works_at, owns, located_in, ...
	Predicate string `json:"predicate"`
	// Object is nil when ObjectLiteral carries the value instead.
	Object         *MemoryEntity `json:"object,omitempty"`
	ObjectLiteral  string        `json:"objectLiteral,omitempty"`
	Weight         float64       `json:"weight"`
	ValidFrom      string        `json:"validFrom,omitempty"`
	ValidTo        string        `json:"validTo,omitempty"`
	SourceMemoryID string        `json:"sourceMemoryId,omitempty"`
	// Hop is the distance from the seed entity on a Traverse — 1 for a direct
	// neighbour. ListEdges has no seed, so every edge comes back with Hop 0.
	Hop int `json:"hop"`
}

// ExtractionState reports whether KG extraction is on, platform-wide and for
// this project.
type ExtractionState struct {
	PlatformEnabled bool `json:"platformEnabled"`
	ProjectEnabled  bool `json:"projectEnabled"`
}

// GraphStats holds aggregate graph counts.
//
// MemoriesWithEdges against your total memory count is the useful ratio: it
// says how much of what you remember made it into the graph rather than
// remaining an isolated embedding. A low ratio usually means extraction is off
// — check Extraction before concluding the corpus simply had no relations.
type GraphStats struct {
	EntityCount int64 `json:"entityCount"`
	EdgeCount   int64 `json:"edgeCount"`
	// MemoriesWithEdges counts distinct memories that produced at least one edge.
	MemoriesWithEdges int64            `json:"memoriesWithEdges"`
	RetiredEntities   int64            `json:"retiredEntities"`
	RetiredEdges      int64            `json:"retiredEdges"`
	EntitiesByType    map[string]int64 `json:"entitiesByType,omitempty"`
	Extraction        *ExtractionState `json:"extraction,omitempty"`
}

// EntityWithEdges is an entity plus its 1-hop neighbourhood.
type EntityWithEdges struct {
	Entity *MemoryEntity        `json:"entity"`
	Edges  []GraphTraversalEdge `json:"edges"`
}

// ListEntitiesParams filters ListEntities. Zero values are omitted.
type ListEntitiesParams struct {
	Type  string
	Scope string
	// Search is a substring match against CanonicalName and every alias.
	Search string
	Limit  int
	Offset int
}

// TraverseParams tunes a walk. Zero values are omitted.
type TraverseParams struct {
	// Hops out from the seed entity, 1-3.
	Hops int
	// Predicates restricts the walk, e.g. []string{"member_of", "led_by"}.
	Predicates []string
	AsOf       string
}

// Stats returns aggregate counts for the whole graph.
//
// Prefer this over len(ListEntities(...)) for any "how big is it" question:
// these are SQL COUNT(*)s over the full table, where the list routes page and
// would report the page size as the total.
func (s *GraphService) Stats(ctx context.Context) (*GraphStats, error) {
	var out GraphStats
	err := s.c.do(ctx, "GET", "/admin/memory/graph/stats", nil, nil, &out)
	return &out, err
}

// ListEntities returns entities filtered by type/scope or a substring of name
// or alias.
func (s *GraphService) ListEntities(ctx context.Context, p ListEntitiesParams) ([]MemoryEntity, error) {
	q := url.Values{}
	if p.Type != "" {
		q.Set("type", p.Type)
	}
	if p.Scope != "" {
		q.Set("scope", p.Scope)
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		q.Set("offset", strconv.Itoa(p.Offset))
	}
	var out []MemoryEntity
	err := s.c.do(ctx, "GET", "/admin/memory/entities", q, nil, &out)
	return out, err
}

// GetEntity returns one entity plus its 1-hop neighbourhood. asOf may be empty.
func (s *GraphService) GetEntity(ctx context.Context, entityID, asOf string) (*EntityWithEdges, error) {
	q := url.Values{}
	if asOf != "" {
		q.Set("asOf", asOf)
	}
	var out EntityWithEdges
	err := s.c.do(ctx, "GET", "/admin/memory/entities/"+url.PathEscape(entityID), q, nil, &out)
	return &out, err
}

// ListEdges returns every currently-valid edge. Use it to render a whole small
// graph; for a large one, seed from an entity and Traverse instead. Both
// arguments are optional (empty string / 0).
func (s *GraphService) ListEdges(ctx context.Context, asOf string, limit int) ([]GraphTraversalEdge, error) {
	q := url.Values{}
	if asOf != "" {
		q.Set("asOf", asOf)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var out []GraphTraversalEdge
	err := s.c.do(ctx, "GET", "/admin/memory/graph/edges", q, nil, &out)
	return out, err
}

// Traverse walks out from a seed entity.
//
// This is the multi-hop path: the edges returned here connect facts no single
// memory states together, which is how a question gets answered from a chain
// rather than from one lucky vector hit.
func (s *GraphService) Traverse(ctx context.Context, entityID string, p TraverseParams) ([]GraphTraversalEdge, error) {
	body := map[string]any{"entityId": entityID}
	if p.Hops > 0 {
		body["hops"] = p.Hops
	}
	if len(p.Predicates) > 0 {
		body["predicates"] = p.Predicates
	}
	if p.AsOf != "" {
		body["asOf"] = p.AsOf
	}
	var out []GraphTraversalEdge
	err := s.c.do(ctx, "POST", "/admin/memory/graph/traverse", nil, body, &out)
	return out, err
}
