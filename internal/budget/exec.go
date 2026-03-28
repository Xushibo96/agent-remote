package budget

import (
	"agent-remote/internal/model"
	"agent-remote/internal/session"
)

type Budgeter struct{}

func New() *Budgeter {
	return &Budgeter{}
}

func (b *Budgeter) BudgetSync(result model.SyncRunResult, _ model.BudgetPolicy) model.SyncRunResult {
	return result
}

func (b *Budgeter) BudgetExec(events []model.ExecEvent, budget model.BudgetPolicy, cursor string, summary model.JobSummary) model.ExecReadResult {
	selected := selectEvents(events, budget)
	if len(selected) == 0 {
		nextCursor := cursor
		if len(events) > 0 {
			nextCursor = session.EncodeCursor(events[len(events)-1].Seq)
		}
		return model.ExecReadResult{
			ID:        summary.ID,
			Cursor:    nextCursor,
			Truncated: len(events) > 0,
			Summary:   summary,
		}
	}

	nextCursor := session.EncodeCursor(selected[len(selected)-1].Seq)
	if nextCursor == "" {
		nextCursor = cursor
	}

	result := model.ExecReadResult{
		ID:        summary.ID,
		Events:    selected,
		Cursor:    nextCursor,
		Truncated: len(selected) < len(events),
		Summary:   summary,
	}
	if result.Truncated {
		result.Summary.TruncatedEvents += int64(len(events) - len(selected))
	}
	return result
}
