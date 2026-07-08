package memmesh

// Subject identifies who/what a memory or prediction is about.
type Subject struct {
	Kind       string `json:"kind"`
	ExternalID string `json:"externalId"`
}

// MemoryItem is a stored memory.
type MemoryItem struct {
	ID             string         `json:"id"`
	PlatformID     string         `json:"platformId"`
	ProjectID      *string        `json:"projectId"`
	Type           string         `json:"type"`
	Content        string         `json:"content"`
	Category       *string        `json:"category"`
	Importance     float64        `json:"importance"`
	Scope          string         `json:"scope"`
	Status         string         `json:"status"`
	Confidence     float64        `json:"confidence"`
	SupersededByID *string        `json:"supersededById"`
	Metadata       map[string]any `json:"metadata"`
	Created        string         `json:"created"`
	Updated        string         `json:"updated"`
}

// SearchResult is one hit from semantic search.
type SearchResult struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Content    string         `json:"content"`
	Category   *string        `json:"category"`
	Similarity float64        `json:"similarity"`
	Metadata   map[string]any `json:"metadata"`
	Scope      string         `json:"scope"`
	Status     string         `json:"status"`
	Importance float64        `json:"importance"`
}

// Insight is a synthesized higher-order memory with provenance.
type Insight struct {
	ID         string   `json:"id"`
	Content    string   `json:"content"`
	SourceIDs  []string `json:"sourceIds"`
	Confidence float64  `json:"confidence"`
}

// GraphEdge is one edge of the temporal knowledge graph.
type GraphEdge struct {
	ID            string  `json:"id"`
	SubjectID     string  `json:"subjectId"`
	Predicate     string  `json:"predicate"`
	ObjectID      *string `json:"objectId"`
	ObjectLiteral *string `json:"objectLiteral"`
	Weight        float64 `json:"weight"`
	ValidFrom     string  `json:"validFrom"`
	ValidTo       *string `json:"validTo"`
}
