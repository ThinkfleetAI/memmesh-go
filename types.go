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

// IngestMediaResult is the outcome of ingesting one media item: the memories
// extracted from it plus the text the model read and where the bytes were kept.
type IngestMediaResult struct {
	Saved          []MemoryItem `json:"saved"`
	CandidateCount int          `json:"candidateCount"`
	ExtractedText  string       `json:"extractedText"`
	Modality       string       `json:"modality"`
	BlobURI        string       `json:"blobUri"`
}

// Insight is a synthesized higher-order memory with provenance.
type Insight struct {
	ID         string   `json:"id"`
	Content    string   `json:"content"`
	SourceIDs  []string `json:"sourceIds"`
	Confidence float64  `json:"confidence"`
}

// Memory item types. `type` is a free-form string on the wire; these are the
// well-known values.
const (
	TypeProcedure = "procedure"
)

// Provenance tiers for the precedence policy, strongest→weakest by default.
const (
	TierHumanVerified = "human_verified"
	TierLocal         = "local"
	TierLicensedBrain = "licensed_brain"
	TierBase          = "base"
)

// Review-queue reasons (highest priority first).
const (
	ReviewPending       = "pending"
	ReviewFlagged       = "flagged"
	ReviewLowConfidence = "low_confidence"
	ReviewStale         = "stale"
)

// ReviewQueueItem is a memory in the adjudication queue plus why it's there.
type ReviewQueueItem struct {
	MemoryItem
	ReviewReason string `json:"reviewReason"`
}

// ProcedureStep is one step of a procedure. Pitfall is an optional warning.
type ProcedureStep struct {
	Text    string `json:"text"`
	Pitfall string `json:"pitfall,omitempty"`
}

// ProcedureInput authors a procedure — "how this job is done here."
type ProcedureInput struct {
	Goal         string          `json:"goal"`
	WhenToUse    string          `json:"whenToUse,omitempty"`
	Steps        []ProcedureStep `json:"steps"`
	FailureModes []string        `json:"failureModes,omitempty"`
	// Category is the heading used when the procedure is injected.
	Category   string `json:"-"`
	Scope      string `json:"-"`
	Importance int    `json:"-"`
}

// PrecedenceOverride is a category-level exception: for this category, this tier wins.
type PrecedenceOverride struct {
	Category    string `json:"category"`
	WinningTier string `json:"winningTier"`
}

// PrecedencePolicy decides which memory wins when two disagree.
type PrecedencePolicy struct {
	DefaultOrder     []string             `json:"defaultOrder"`
	ScopeNearestWins bool                 `json:"scopeNearestWins"`
	Overrides        []PrecedenceOverride `json:"overrides"`
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

// ObserveResponse is what Observe returns: the memories the engine chose to
// keep (empty when the turn was filler — still a success) plus how many
// candidates extraction found before the dedupe/budget pass.
// len(Saved) <= CandidateCount.
type ObserveResponse struct {
	Saved          []MemoryItem `json:"saved"`
	CandidateCount int          `json:"candidateCount"`
}
