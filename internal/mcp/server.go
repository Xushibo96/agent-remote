package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"agent-remote/internal/model"
)

type Server struct {
	adapter *Adapter
}

func NewServer(adapter *Adapter) *Server {
	return &Server{adapter: adapter}
}

func (s *Server) Handle(ctx context.Context, tool string, payload json.RawMessage) (Response, error) {
	if s.adapter == nil {
		return Response{OK: false, State: "error", Error: errorInfo("adapter_missing", "validation", "dispatch", "adapter is not configured")}, nil
	}

	switch tool {
	case "target_add":
		var req model.AddTargetRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return Response{}, err
		}
		return s.adapter.AddTarget(ctx, req)
	case "target_list":
		return s.adapter.ListTargets(ctx)
	case "sync_upload", "sync_download", "sync_bidir":
		var req model.SyncRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return Response{}, err
		}
		if req.Direction == "" {
			switch tool {
			case "sync_upload":
				req.Direction = model.DirectionUpload
			case "sync_download":
				req.Direction = model.DirectionDownload
			case "sync_bidir":
				req.Direction = model.DirectionBidir
			}
		}
		return s.adapter.StartSync(ctx, req)
	case "exec_start":
		var req model.ExecRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return Response{}, err
		}
		return s.adapter.StartExec(ctx, req)
	case "exec_read":
		var req model.ExecReadRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return Response{}, err
		}
		return s.adapter.ReadExec(ctx, req)
	case "exec_stop":
		var req model.ExecStopRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return Response{}, err
		}
		return s.adapter.StopExec(ctx, req)
	case "job_status":
		var req model.JobStatusRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return Response{}, err
		}
		return s.adapter.GetJob(ctx, req)
	default:
		return Response{}, fmt.Errorf("unsupported tool %q", tool)
	}
}
