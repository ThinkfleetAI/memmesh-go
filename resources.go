package memmesh

import (
	"context"
	"net/url"
	"strconv"
)

// EventsService — the proactive event log (alerts evaluate synchronously).
type EventsService struct{ c *Client }

// Emit appends an event. Idempotent on (projectId, dedupeKey).
func (s *EventsService) Emit(ctx context.Context, event map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/events", nil, event, &out)
	return out, err
}

// Poll drains queued events since a cursor.
func (s *EventsService) Poll(ctx context.Context, since string, limit int) (map[string]any, error) {
	q := url.Values{}
	if since != "" {
		q.Set("since", since)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var out map[string]any
	err := s.c.do(ctx, "GET", "/events", q, nil, &out)
	return out, err
}

// AlertsService — durable alert rules over the event/value streams.
type AlertsService struct{ c *Client }

func (s *AlertsService) Create(ctx context.Context, rule map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/alerts", nil, rule, &out)
	return out, err
}
func (s *AlertsService) List(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	err := s.c.do(ctx, "GET", "/alerts", nil, nil, &out)
	return out, err
}
func (s *AlertsService) Get(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "GET", "/alerts/"+id, nil, nil, &out)
	return out, err
}
func (s *AlertsService) Update(ctx context.Context, id string, patch map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "PATCH", "/alerts/"+id, nil, patch, &out)
	return out, err
}
func (s *AlertsService) Delete(ctx context.Context, id string) error {
	return s.c.do(ctx, "DELETE", "/alerts/"+id, nil, nil, nil)
}
func (s *AlertsService) ListFires(ctx context.Context, ruleID string) ([]map[string]any, error) {
	q := url.Values{}
	if ruleID != "" {
		q.Set("ruleId", ruleID)
	}
	var out []map[string]any
	err := s.c.do(ctx, "GET", "/alerts/fires", q, nil, &out)
	return out, err
}

// LearningService — the closed decision→outcome loop that recalibrates patterns.
type LearningService struct{ c *Client }

func (s *LearningService) RecordDecision(ctx context.Context, decision map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/learning/decisions", nil, decision, &out)
	return out, err
}
func (s *LearningService) RecordOutcome(ctx context.Context, outcome map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/learning/outcomes", nil, outcome, &out)
	return out, err
}
func (s *LearningService) GetOutcomes(ctx context.Context, subject Subject) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/learning/outcomes/query", nil, map[string]any{"subject": subject}, &out)
	return out, err
}
func (s *LearningService) GetEffectiveness(ctx context.Context, query map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/learning/effectiveness", nil, query, &out)
	return out, err
}

// TypedService — structured/numeric attributes the engine reasons over.
type TypedService struct{ c *Client }

func (s *TypedService) RegisterAttribute(ctx context.Context, def map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/typed/attributes", nil, def, &out)
	return out, err
}
func (s *TypedService) Ingest(ctx context.Context, observations []map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/typed/observations", nil, map[string]any{"observations": observations}, &out)
	return out, err
}
func (s *TypedService) QueryObservations(ctx context.Context, query map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/typed/observations/query", nil, query, &out)
	return out, err
}
func (s *TypedService) Accumulator(ctx context.Context, subject Subject, attribute string) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/typed/accumulator", nil, map[string]any{"subject": subject, "attribute": attribute}, &out)
	return out, err
}

// ComplianceService — governance: export, erase, audit, packs.
type ComplianceService struct{ c *Client }

func (s *ComplianceService) ExportSubject(ctx context.Context, subject Subject) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/admin/memory/compliance/export", nil, map[string]any{"subject": subject}, &out)
	return out, err
}
func (s *ComplianceService) HardDeleteSubject(ctx context.Context, subject Subject, reason string, dryRun bool) (map[string]any, error) {
	var out map[string]any
	body := map[string]any{"subject": subject, "reason": reason, "dryRun": dryRun}
	err := s.c.do(ctx, "POST", "/admin/memory/compliance/erase", nil, body, &out)
	return out, err
}
func (s *ComplianceService) ListAuditEvents(ctx context.Context, limit int) ([]map[string]any, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var out []map[string]any
	err := s.c.do(ctx, "GET", "/admin/memory/compliance/audit", q, nil, &out)
	return out, err
}
func (s *ComplianceService) ListPacks(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	err := s.c.do(ctx, "GET", "/admin/memory/compliance/packs", nil, nil, &out)
	return out, err
}

// HealthService — biomarker/demographics ingest + health profile (requires the
// @thinkfleet/pack-healthcare pack).
type HealthService struct{ c *Client }

func (s *HealthService) RecordBiomarker(ctx context.Context, subject Subject, biomarker map[string]any) (map[string]any, error) {
	var out map[string]any
	body := map[string]any{"subject": subject}
	for k, v := range biomarker {
		body[k] = v
	}
	err := s.c.do(ctx, "POST", "/health/biomarkers", nil, body, &out)
	return out, err
}
func (s *HealthService) GetProfile(ctx context.Context, subject Subject) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/health/profile", nil, map[string]any{"subject": subject}, &out)
	return out, err
}
func (s *HealthService) GetCohortRisk(ctx context.Context, subject Subject) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/health/cohort-risk", nil, map[string]any{"subject": subject}, &out)
	return out, err
}
