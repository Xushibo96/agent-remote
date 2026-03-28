package budget

import (
	"strings"

	"agent-remote/internal/model"
)

func selectEvents(events []model.ExecEvent, budget model.BudgetPolicy) []model.ExecEvent {
	if len(events) == 0 {
		return nil
	}

	maxLines := budget.MaxLines
	maxBytes := budget.MaxBytes
	windowBytes := budget.WindowBytes

	if maxLines <= 0 {
		maxLines = len(events)
	}
	if maxBytes <= 0 {
		maxBytes = 64 * 1024
	}
	if windowBytes <= 0 {
		windowBytes = maxBytes
	}

	selected := make([]model.ExecEvent, 0, len(events))
	usedBytes := 0

	start := 0
	if windowBytes > 0 {
		for i := len(events) - 1; i >= 0; i-- {
			eventBytes := len(events[i].Payload)
			if eventBytes > windowBytes && len(selected) == 0 {
				start = i
				break
			}
			if usedBytes+eventBytes > windowBytes && len(selected) > 0 {
				break
			}
			usedBytes += eventBytes
			start = i
			if usedBytes >= windowBytes {
				break
			}
		}
	}

	for _, event := range events[start:] {
		if len(selected) >= maxLines {
			break
		}
		if maxBytes > 0 && totalBytes(selected)+len(event.Payload) > maxBytes {
			break
		}
		selected = append(selected, event)
	}

	if budget.KeepLifecycle || budget.KeepErrors {
		selected = mergePriorityEvents(selected, events, budget)
	}

	return selected
}

func totalBytes(events []model.ExecEvent) int {
	total := 0
	for _, event := range events {
		total += len(event.Payload)
	}
	return total
}

func mergePriorityEvents(selected, all []model.ExecEvent, budget model.BudgetPolicy) []model.ExecEvent {
	seen := make(map[int64]struct{}, len(selected))
	for _, event := range selected {
		seen[event.Seq] = struct{}{}
	}

	for _, event := range all {
		if _, ok := seen[event.Seq]; ok {
			continue
		}
		if budget.KeepLifecycle && isLifecycleEvent(event) {
			selected = append(selected, event)
			seen[event.Seq] = struct{}{}
			continue
		}
		if budget.KeepErrors && isErrorEvent(event) {
			selected = append(selected, event)
			seen[event.Seq] = struct{}{}
		}
	}

	if len(selected) > 1 {
		sortEvents(selected)
	}
	return selected
}

func isLifecycleEvent(event model.ExecEvent) bool {
	switch strings.ToLower(event.Type) {
	case "started", "completed", "failed", "stopped", "progress":
		return true
	default:
		return false
	}
}

func isErrorEvent(event model.ExecEvent) bool {
	if strings.ToLower(event.Stream) == "stderr" {
		return true
	}
	return strings.Contains(strings.ToLower(event.Type), "fail")
}

func sortEvents(events []model.ExecEvent) {
	for i := 1; i < len(events); i++ {
		j := i
		for j > 0 && events[j-1].Seq > events[j].Seq {
			events[j-1], events[j] = events[j], events[j-1]
			j--
		}
	}
}

