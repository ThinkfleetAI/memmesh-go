package memmesh

import "testing"

func TestRenderProcedureContent(t *testing.T) {
	got := RenderProcedureContent(ProcedureInput{
		Goal:      "Refund a charge",
		WhenToUse: "double charge",
		Steps: []ProcedureStep{
			{Text: "Find it"},
			{Text: "Refund", Pitfall: "never twice"},
		},
		FailureModes: []string{"wrong card", "  "},
	})
	want := "Goal: Refund a charge\n" +
		"When: double charge\n" +
		"Steps:\n" +
		"1. Find it\n" +
		"2. Refund (watch out: never twice)\n" +
		"Avoid:\n" +
		"- wrong card"
	if got != want {
		t.Fatalf("mismatch:\n got=%q\nwant=%q", got, want)
	}
}
