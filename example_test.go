package memmesh_test

import (
	"context"

	memmesh "github.com/ThinkfleetAI/memmesh-go"
)

func Example() {
	ctx := context.Background()
	mm := memmesh.New("sk-...", "proj_...")

	_, _ = mm.Memory.Observe(ctx, memmesh.Observe{
		Subject: memmesh.Subject{Kind: "contact", ExternalID: "sarah"},
		Content: "Prefers email over phone.",
	})
	_, _ = mm.Memory.Search(ctx, "how to reach sarah", memmesh.SearchOpts{Limit: 5})
	_, _ = mm.Memory.Reflect(ctx, memmesh.ReflectOpts{MaxInsights: 3})
}
