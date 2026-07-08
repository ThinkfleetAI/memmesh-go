# memmesh-go

Official Go client for **[MemMesh](https://memmesh.ai)** — memory + prediction
for AI agents. Semantic recall, a bi-temporal knowledge graph, belief revision,
reflection, and calibrated forecasting behind an idiomatic Go client.

```bash
go get github.com/ThinkfleetAI/memmesh-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"

	memmesh "github.com/ThinkfleetAI/memmesh-go"
)

func main() {
	ctx := context.Background()
	mm := memmesh.New("sk-...", "proj_...") // apiKey, projectId

	// Remember something
	mm.Memory.Observe(ctx, memmesh.Observe{
		Subject: memmesh.Subject{Kind: "contact", ExternalID: "sarah"},
		Content: "Prefers email over phone.",
	})

	// Recall it, semantically
	hits, _ := mm.Memory.Search(ctx, "how to reach sarah", memmesh.SearchOpts{Limit: 5})
	for _, h := range hits {
		fmt.Println(h.Content)
	}

	// Synthesize higher-order insights, with provenance
	res, _ := mm.Memory.Reflect(ctx, memmesh.ReflectOpts{MaxInsights: 3})
	for _, in := range res.Insights {
		fmt.Printf("%s  (%.0f%%)\n", in.Content, in.Confidence*100)
	}

	// Point-in-time knowledge graph — "what did we believe on date X"
	edges, _ := mm.Context.QueryGraph(ctx, memmesh.GraphOpts{AsOf: "2026-03-01T00:00:00Z"})
	_ = edges
}
```

## Surface

| Service | Highlights |
| --- | --- |
| `mm.Memory` | Observe, Create, Search, List, Update, Delete, Stats, Confirm, Promote, Feedback, **Reflect**, **PrefetchRelated**, Dedup, BackfillEmbeddings |
| `mm.Lattice` | Predict (any target, calibrated + abstaining), Mine, Profile, PredictByCohort, Calibration |
| `mm.Context` | Build, **BatchBuild** (bulk, ≤500), **QueryGraph** (point-in-time) |
| `mm.Events` · `mm.Alerts` | Emit/Poll events; alert rules + fires |
| `mm.Learning` | RecordDecision → RecordOutcome → Effectiveness (the self-improving loop) |
| `mm.Typed` · `mm.Compliance` · `mm.Health` | Typed attributes; export/erase/audit/packs; healthcare pack |

## Config

- `memmesh.New(apiKey, projectID, opts...)`
- `memmesh.WithBaseURL("https://your-host")` — self-hosted or a different region
- `memmesh.WithHTTPClient(&http.Client{...})` — timeouts, proxies, retries

Errors from the API are `*memmesh.APIError` (`.Status`, `.Code`).

Apache-2.0 · [memmesh.ai](https://memmesh.ai) · [docs](https://docs.memmesh.ai)
