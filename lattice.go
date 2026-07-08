package memmesh

import "context"

// LatticeService is behavioral intelligence: mine patterns, forecast, profile.
type LatticeService struct{ c *Client }

// Predict forecasts a target for a subject from its observation history —
// calibrated, provenanced, and abstaining when signal is thin. `target`
// describes what to predict (e.g. {"kind":"event_occurrence","event":"churn"}).
func (s *LatticeService) Predict(ctx context.Context, subject Subject, target map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/lattice/predict", nil, map[string]any{"subject": subject, "target": target}, &out)
	return out, err
}

// Mine runs pattern extraction over recent activity.
func (s *LatticeService) Mine(ctx context.Context, subject *Subject) (map[string]any, error) {
	body := map[string]any{}
	if subject != nil {
		body["subject"] = *subject
	}
	var out map[string]any
	err := s.c.do(ctx, "POST", "/lattice/patterns/extract", nil, body, &out)
	return out, err
}

// Profile returns a behavioral snapshot (RFM, top entity, cadence, risks).
func (s *LatticeService) Profile(ctx context.Context, subject Subject) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/lattice/profile", nil, map[string]any{"subject": subject}, &out)
	return out, err
}

// PredictByCohort forecasts from "people like this one".
func (s *LatticeService) PredictByCohort(ctx context.Context, subject Subject, target map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "POST", "/lattice/cohort/predict", nil, map[string]any{"subject": subject, "target": target}, &out)
	return out, err
}

// Calibration reports how honest the engine's confidence is (buckets → hit rate).
func (s *LatticeService) Calibration(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := s.c.do(ctx, "GET", "/lattice/calibration", nil, nil, &out)
	return out, err
}
