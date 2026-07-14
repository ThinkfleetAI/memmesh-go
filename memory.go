package memmesh

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
)

// MemoryService is the primary surface: ingest, recall, and the admin +
// SOTA operations (reflect, prefetch-related, dedup, ...).
type MemoryService struct{ c *Client }

// Observe records that something happened. The engine mines, wires the graph,
// and revises beliefs server-side.
type Observe struct {
	Subject      Subject        `json:"-"`
	Content      string         `json:"-"`
	Type         string         `json:"-"` // default "event"
	Scope        string         `json:"-"` // default "project"
	Importance   int            `json:"-"` // default 5
	Category     string         `json:"-"`
	ActivityType string         `json:"-"`
	OccurredAt   string         `json:"-"`
	Metadata     map[string]any `json:"-"`
}

func (o Observe) body() map[string]any {
	md := map[string]any{"subject": o.Subject}
	if o.ActivityType != "" {
		md["eventType"] = o.ActivityType
	}
	if o.OccurredAt != "" {
		md["occurredAt"] = o.OccurredAt
	}
	for k, v := range o.Metadata {
		md[k] = v
	}
	typ := o.Type
	if typ == "" {
		typ = "event"
	}
	scope := o.Scope
	if scope == "" {
		scope = "project"
	}
	imp := o.Importance
	if imp == 0 {
		imp = 5
	}
	b := map[string]any{"content": o.Content, "type": typ, "scope": scope, "importance": imp, "source": "admin_created", "metadata": md}
	if o.OccurredAt != "" {
		// Event time, not ingest time. validFrom is the field behavior mining
		// buckets day-of-week / hour-of-day on, so this is what makes a backfill
		// work: without it every historical row lands at the moment of import and
		// the mined patterns describe the import job rather than the data. The
		// metadata copy above is kept only for readers that already look for it.
		b["validFrom"] = o.OccurredAt
	}
	if o.Category != "" {
		b["category"] = o.Category
	}
	return b
}

// Observe ingests an event-shaped memory (the primary agent ingestion call).
func (s *MemoryService) Observe(ctx context.Context, o Observe) (*MemoryItem, error) {
	var out MemoryItem
	err := s.c.do(ctx, "POST", "/admin/memory", nil, o.body(), &out)
	return &out, err
}

// IngestMedia ingests an image / audio / document. The engine extracts text
// (vision, transcription, or OCR via LiteLLM) and runs it through the observe
// pipeline, so the result is real memories, not just a stored file. Requires
// multimodal to be enabled on the engine.
type IngestMedia struct {
	Media     []byte `json:"-"`
	MimeType  string `json:"-"`
	UserID    string `json:"-"`
	AgentID   string `json:"-"`
	SessionID string `json:"-"`
	Source    string `json:"-"`
}

func (m IngestMedia) body() map[string]any {
	b := map[string]any{
		"dataBase64": base64.StdEncoding.EncodeToString(m.Media),
		"mimeType":   m.MimeType,
	}
	if m.UserID != "" {
		b["userId"] = m.UserID
	}
	if m.AgentID != "" {
		b["agentId"] = m.AgentID
	}
	if m.SessionID != "" {
		b["sessionId"] = m.SessionID
	}
	if m.Source != "" {
		b["source"] = m.Source
	}
	return b
}

// IngestMedia sends media to the engine and returns the extracted memories.
func (s *MemoryService) IngestMedia(ctx context.Context, in IngestMedia) (*IngestMediaResult, error) {
	var out IngestMediaResult
	err := s.c.do(ctx, "POST", "/memory/media", nil, in.body(), &out)
	return &out, err
}

// Create seeds a memory directly (knowledge the agent should have without
// observing it first).
type Create struct {
	Content    string         `json:"content"`
	Type       string         `json:"type,omitempty"`
	Scope      string         `json:"scope,omitempty"`
	Importance int            `json:"importance,omitempty"`
	Category   string         `json:"category,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

func (s *MemoryService) Create(ctx context.Context, in Create) (*MemoryItem, error) {
	var out MemoryItem
	err := s.c.do(ctx, "POST", "/admin/memory", nil, in, &out)
	return &out, err
}

// SearchOpts narrows a semantic search.
type SearchOpts struct {
	Limit  int
	Offset int
	Scope  string
	Status string
}

// Search runs hybrid semantic + keyword search across everything the project
// can see. Page by bumping Offset.
func (s *MemoryService) Search(ctx context.Context, query string, opts SearchOpts) ([]SearchResult, error) {
	body := map[string]any{"query": query}
	if opts.Limit > 0 {
		body["limit"] = opts.Limit
	}
	if opts.Offset > 0 {
		body["offset"] = opts.Offset
	}
	if opts.Scope != "" {
		body["scope"] = opts.Scope
	}
	if opts.Status != "" {
		body["status"] = opts.Status
	}
	var out []SearchResult
	err := s.c.do(ctx, "POST", "/admin/memory/search", nil, body, &out)
	return out, err
}

// ListOpts filters a listing.
type ListOpts struct {
	Type   string
	Scope  string
	Status string
	Limit  int
	Offset int
}

func (s *MemoryService) List(ctx context.Context, opts ListOpts) ([]MemoryItem, error) {
	q := url.Values{}
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	if opts.Scope != "" {
		q.Set("scope", opts.Scope)
	}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}
	var out []MemoryItem
	err := s.c.do(ctx, "GET", "/admin/memory", q, nil, &out)
	return out, err
}

// Get fetches a single memory by id.
//
// The point-lookup counterpart to List/Search: without it, a caller holding a
// memory id (from a pattern's sourceMemoryIds, an audit log, a webhook) had no
// way to resolve it and had to page List hoping the row was still on one.
func (s *MemoryService) Get(ctx context.Context, id string) (*MemoryItem, error) {
	var out MemoryItem
	if err := s.c.do(ctx, "GET", "/admin/memory/"+id, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update patches a memory (nil/empty fields are omitted).
func (s *MemoryService) Update(ctx context.Context, id string, patch map[string]any) (*MemoryItem, error) {
	var out MemoryItem
	err := s.c.do(ctx, "PATCH", "/admin/memory/"+id, nil, patch, &out)
	return &out, err
}

// Delete hard-deletes a memory.
func (s *MemoryService) Delete(ctx context.Context, id string) error {
	return s.c.do(ctx, "DELETE", "/admin/memory/"+id, nil, nil, nil)
}

// Stats returns aggregate counts for the admin dashboard.
func (s *MemoryService) Stats(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "GET", "/admin/memory/stats", nil, nil, &out)
	return out, err
}

// Confirm approves ("confirmed") or rejects ("rejected") a review-queue item.
func (s *MemoryService) Confirm(ctx context.Context, id, status, comment string) (*MemoryItem, error) {
	body := map[string]any{"status": status}
	if comment != "" {
		body["comment"] = comment
	}
	var out MemoryItem
	err := s.c.do(ctx, "POST", fmt.Sprintf("/admin/memory/%s/confirm", id), nil, body, &out)
	return &out, err
}

// Promote moves a memory to a broader scope.
func (s *MemoryService) Promote(ctx context.Context, id, targetScope string) (*MemoryItem, error) {
	var out MemoryItem
	err := s.c.do(ctx, "POST", fmt.Sprintf("/admin/memory/%s/promote", id), nil, map[string]any{"targetScope": targetScope}, &out)
	return &out, err
}

// Feedback reinforces (positive) or flags (negative) a memory.
func (s *MemoryService) Feedback(ctx context.Context, id, rating, comment string) error {
	body := map[string]any{"memoryId": id, "rating": rating}
	if comment != "" {
		body["comment"] = comment
	}
	return s.c.do(ctx, "POST", "/memory/feedback", nil, body, nil)
}

// ── Admin / maintenance ──────────────────────────────────────────────────

// DedupResult is the outcome of a semantic dedup pass.
type DedupResult struct {
	Scanned    int `json:"scanned"`
	Groups     int `json:"groups"`
	Superseded int `json:"superseded"`
}

// Dedup collapses near-duplicate memories (cosine >= threshold).
func (s *MemoryService) Dedup(ctx context.Context, threshold float64, scanLimit int) (*DedupResult, error) {
	body := map[string]any{}
	if threshold > 0 {
		body["threshold"] = threshold
	}
	if scanLimit > 0 {
		body["scanLimit"] = scanLimit
	}
	var out DedupResult
	err := s.c.do(ctx, "POST", "/admin/memory/dedup", nil, body, &out)
	return &out, err
}

// BackfillEmbeddings vectorizes items missing an embedding. Call until Embedded==0.
func (s *MemoryService) BackfillEmbeddings(ctx context.Context, batch int) (embedded int, err error) {
	body := map[string]any{}
	if batch > 0 {
		body["batch"] = batch
	}
	var out struct {
		Embedded int `json:"embedded"`
	}
	err = s.c.do(ctx, "POST", "/admin/memory/embeddings/backfill", nil, body, &out)
	return out.Embedded, err
}

// ReflectResult holds synthesized insights.
type ReflectResult struct {
	Insights          []Insight `json:"insights"`
	SourcesConsidered int       `json:"sourcesConsidered"`
	DryRun            bool      `json:"dryRun"`
}

// ReflectOpts tunes a reflection pass.
type ReflectOpts struct {
	UserID      string
	MaxSources  int
	MaxInsights int
	DryRun      bool
}

// Reflect synthesizes higher-order insight memories from recent confirmed
// memories, each provenanced back to its sources.
func (s *MemoryService) Reflect(ctx context.Context, opts ReflectOpts) (*ReflectResult, error) {
	body := map[string]any{"dryRun": opts.DryRun}
	if opts.UserID != "" {
		body["userId"] = opts.UserID
	}
	if opts.MaxSources > 0 {
		body["maxSources"] = opts.MaxSources
	}
	if opts.MaxInsights > 0 {
		body["maxInsights"] = opts.MaxInsights
	}
	var out ReflectResult
	err := s.c.do(ctx, "POST", "/admin/memory/reflect", nil, body, &out)
	return &out, err
}

// PrefetchRelated returns memories linked to the same graph entities as the
// seeds — the context most likely needed next (spreading activation).
func (s *MemoryService) PrefetchRelated(ctx context.Context, seedMemoryIDs []string, limit int) ([]MemoryItem, error) {
	body := map[string]any{"seedMemoryIds": seedMemoryIDs}
	if limit > 0 {
		body["limit"] = limit
	}
	var out []MemoryItem
	err := s.c.do(ctx, "POST", "/admin/memory/prefetch-related", nil, body, &out)
	return out, err
}
