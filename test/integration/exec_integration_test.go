package integration

import (
	"context"
	"os"
	"testing"

	"agent-remote/internal/model"
)

func TestExecIntegrationRequiresOptIn(t *testing.T) {
	if os.Getenv("AGENT_REMOTE_INTEGRATION") == "" {
		t.Skip("set AGENT_REMOTE_INTEGRATION=1 to run real SSH integration")
	}

	_ = context.Background()
	_ = model.ExecRequest{}
}

