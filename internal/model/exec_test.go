package model

import "testing"

func TestNormalizeExecRequestRequiresCommand(t *testing.T) {
	_, err := NormalizeExecRequest(ExecRequest{
		TargetID: "prod",
	})
	if err == nil {
		t.Fatal("expected missing command error")
	}
}
