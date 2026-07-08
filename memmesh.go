// Package memmesh is the official Go client for MemMesh — memory + prediction
// for AI agents.
//
//	mm := memmesh.New("sk-...", "proj_...")
//	mm.Memory.Observe(ctx, memmesh.Observe{
//	    Subject: memmesh.Subject{Kind: "contact", ExternalID: "sarah"},
//	    Content: "Prefers email over phone.",
//	})
//	hits, _ := mm.Memory.Search(ctx, "how to reach sarah", memmesh.SearchOpts{Limit: 5})
package memmesh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://app.memmesh.ai"

// Client is the entry point. Construct with New. Safe for concurrent use.
type Client struct {
	apiKey    string
	projectID string
	baseURL   string
	http      *http.Client

	Memory     *MemoryService
	Lattice    *LatticeService
	Context    *ContextService
	Events     *EventsService
	Alerts     *AlertsService
	Health     *HealthService
	Learning   *LearningService
	Typed      *TypedService
	Compliance *ComplianceService
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the API base (default https://app.memmesh.ai).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient supplies a custom *http.Client (timeouts, proxies, retries).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

// New creates a client. apiKey is your sk-... key; projectID is the default
// project for every call.
func New(apiKey, projectID string, opts ...Option) *Client {
	c := &Client{
		apiKey:    apiKey,
		projectID: projectID,
		baseURL:   defaultBaseURL,
		http:      &http.Client{Timeout: 60 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	c.Memory = &MemoryService{c: c}
	c.Lattice = &LatticeService{c: c}
	c.Context = &ContextService{c: c}
	c.Events = &EventsService{c: c}
	c.Alerts = &AlertsService{c: c}
	c.Health = &HealthService{c: c}
	c.Learning = &LearningService{c: c}
	c.Typed = &TypedService{c: c}
	c.Compliance = &ComplianceService{c: c}
	return c
}

// APIError is a non-2xx response from the API.
type APIError struct {
	Status int
	Code   string `json:"code"`
	Body   string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("memmesh: %d %s", e.Status, e.Code)
	}
	return fmt.Sprintf("memmesh: %d %s", e.Status, e.Body)
}

// do performs an authenticated request against the project-scoped API and
// decodes the JSON response into out (may be nil). query params are optional.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	u := fmt.Sprintf("%s/api/v1/projects/%s%s", c.baseURL, c.projectID, path)
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ae := &APIError{Status: resp.StatusCode, Body: string(data)}
		_ = json.Unmarshal(data, ae) // best-effort code
		return ae
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}
