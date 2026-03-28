package rsync

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type RunRequest struct {
	Source      string
	Destination string
	Archive     bool
	Recursive   bool
	Delete      bool
	Partial     bool
	DryRun      bool
	Includes    []string
	Excludes    []string
	Timeout     time.Duration
}

type Command struct {
	Path string
	Args []string
}

type Runner struct {
	Binary string
}

func NewRunner(binary string) *Runner {
	if binary == "" {
		binary = "rsync"
	}
	return &Runner{Binary: binary}
}

func (r *Runner) Build(req RunRequest) Command {
	args := BuildArgs(req)
	return Command{Path: r.Binary, Args: args}
}

func BuildArgs(req RunRequest) []string {
	args := make([]string, 0, 16)
	if req.Archive {
		args = append(args, "-a")
	}
	if req.Recursive {
		args = append(args, "-r")
	}
	if req.Delete {
		args = append(args, "--delete")
	}
	if req.Partial {
		args = append(args, "--partial")
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	for _, include := range req.Includes {
		args = append(args, "--include="+include)
	}
	for _, exclude := range req.Excludes {
		args = append(args, "--exclude="+exclude)
	}
	args = append(args, req.Source, req.Destination)
	return args
}

func (r *Runner) Run(ctx context.Context, req RunRequest) (*exec.Cmd, error) {
	if r == nil {
		return nil, fmt.Errorf("runner is nil")
	}
	cmd := exec.CommandContext(ctx, r.Binary, BuildArgs(req)...)
	if strings.TrimSpace(req.Source) == "" || strings.TrimSpace(req.Destination) == "" {
		return nil, fmt.Errorf("source and destination are required")
	}
	return cmd, nil
}
