package perf

import (
	"testing"
	"time"

	"agent-remote/internal/budget"
	"agent-remote/internal/model"
)

func BenchmarkBudgetExec(b *testing.B) {
	events := make([]model.ExecEvent, 0, 1000)
	for i := 0; i < 1000; i++ {
		events = append(events, model.ExecEvent{
			Seq:       int64(i + 1),
			Type:      "stdout",
			Stream:    "stdout",
			Payload:   "payload-line",
			Timestamp: time.Now().UTC(),
		})
	}

	budgeter := budget.New()
	policy := model.BudgetPolicy{MaxLines: 100, MaxBytes: 4096, WindowBytes: 4096, KeepErrors: true, KeepLifecycle: true}
	summary := model.JobSummary{ID: "bench"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = budgeter.BudgetExec(events, policy, "", summary)
	}
}
