package memmesh_test

import (
	"context"

	memmesh "github.com/ThinkfleetAI/memmesh-go"
)

func Example() {
	ctx := context.Background()
	mm := memmesh.New("sk-...", "proj_...")

	// Primary path: hand the engine the raw turn; it runs the noise filter and
	// returns only what's worth keeping (Saved is empty when the turn was filler).
	_, _ = mm.Memory.Observe(ctx, memmesh.Observe{
		Text: "Sarah told me she prefers email over phone.",
	})
	_, _ = mm.Memory.Search(ctx, "how to reach sarah", memmesh.SearchOpts{Limit: 5})
	_, _ = mm.Memory.Reflect(ctx, memmesh.ReflectOpts{MaxInsights: 3})
}
