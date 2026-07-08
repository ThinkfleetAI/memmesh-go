package memmesh

import "context"

// ContextService assembles LLM-ready context bundles and queries the temporal
// knowledge graph.
type ContextService struct{ c *Client }

// ContextOpts tunes a context bundle.
type ContextOpts struct {
	Include           []string
	MaxTokens         int
	MemoryLimit       int
	PredictionLimit   int
	ExcludeCategories []string
}

func (o ContextOpts) apply(m map[string]any) {
	if len(o.Include) > 0 {
		m["include"] = o.Include
	}
	if o.MaxTokens > 0 {
		m["maxTokens"] = o.MaxTokens
	}
	if o.MemoryLimit > 0 {
		m["memoryLimit"] = o.MemoryLimit
	}
	if o.PredictionLimit > 0 {
		m["predictionLimit"] = o.PredictionLimit
	}
	if len(o.ExcludeCategories) > 0 {
		m["excludeCategories"] = o.ExcludeCategories
	}
}

// Build aggregates profile + patterns + predictions + memories + observations
// for one subject into a single provenanced bundle (returned as a generic map).
func (s *ContextService) Build(ctx context.Context, subject Subject, opts ContextOpts) (map[string]any, error) {
	body := map[string]any{"subject": subject}
	opts.apply(body)
	var out map[string]any
	err := s.c.do(ctx, "POST", "/lattice/context", nil, body, &out)
	return out, err
}

// BatchBuild builds bundles for many subjects (<=500) in one call.
func (s *ContextService) BatchBuild(ctx context.Context, subjects []Subject, opts ContextOpts) ([]map[string]any, error) {
	body := map[string]any{"subjects": subjects}
	opts.apply(body)
	var out struct {
		Bundles []map[string]any `json:"bundles"`
	}
	err := s.c.do(ctx, "POST", "/lattice/context/batch", nil, body, &out)
	return out.Bundles, err
}

// GraphOpts filters a point-in-time graph query.
type GraphOpts struct {
	SubjectID string
	Predicate string
	AsOf      string // RFC3339; empty = current graph
	Limit     int
}

// QueryGraph returns the edges valid AT AsOf (or current), filtered by
// subject/predicate.
func (s *ContextService) QueryGraph(ctx context.Context, opts GraphOpts) ([]GraphEdge, error) {
	body := map[string]any{}
	if opts.SubjectID != "" {
		body["subjectId"] = opts.SubjectID
	}
	if opts.Predicate != "" {
		body["predicate"] = opts.Predicate
	}
	if opts.AsOf != "" {
		body["asOf"] = opts.AsOf
	}
	if opts.Limit > 0 {
		body["limit"] = opts.Limit
	}
	var out struct {
		Edges []GraphEdge `json:"edges"`
	}
	err := s.c.do(ctx, "POST", "/lattice/graph/query", nil, body, &out)
	return out.Edges, err
}
