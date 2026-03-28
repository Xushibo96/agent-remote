package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agent-remote/internal/app"
	"agent-remote/internal/budget"
	"agent-remote/internal/connection"
	"agent-remote/internal/config"
	"agent-remote/internal/credential"
	execengine "agent-remote/internal/exec"
	"agent-remote/internal/mcp"
	"agent-remote/internal/model"
	"agent-remote/internal/rsync"
	"agent-remote/internal/secret"
	"agent-remote/internal/session"
	"agent-remote/internal/store"
	syncengine "agent-remote/internal/sync"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	service := buildService(cfg)
	adapter := mcp.NewAdapter(service)
	if len(os.Args) < 2 {
		printJSON(map[string]any{
			"ok":        true,
			"state":     "ready",
			"configDir": cfg.ConfigDir,
			"commands":  []string{"target", "sync", "exec", "job"},
		})
		return
	}

	if err := run(context.Background(), adapter, os.Args[1:]); err != nil {
		printJSON(adapter.ErrorResponse(err))
		os.Exit(1)
	}
}

func buildService(cfg config.Config) mcp.Service {
	cfgStore := store.NewFileConfigStore(cfg.ConfigFile)
	keyManager := secret.NewKeyManager(secret.OSKeyring{}, "", "")
	vault := credential.NewVault(cfgStore, keyManager)
	jobs, err := session.NewStoreWithPath(filepath.Join(cfg.ConfigDir, "sessions.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "session store error: %v\n", err)
		os.Exit(1)
	}
	budgeter := budget.New()
	connections := connection.NewManager()
	exec := execengine.NewEngine(connections, jobs, budgeter)
	sync := syncengine.NewRemoteEngine(connections, rsync.NewRunner(""))
	return app.NewOrchestrator(vault, jobs, sync, exec)
}

func run(ctx context.Context, adapter *mcp.Adapter, args []string) error {
	switch args[0] {
	case "target":
		return runTarget(ctx, adapter, args[1:])
	case "sync":
		return runSync(ctx, adapter, args[1:])
	case "exec":
		return runExec(ctx, adapter, args[1:])
	case "job":
		return runJob(ctx, adapter, args[1:])
	default:
		return fmt.Errorf("unsupported command %q", args[0])
	}
}

func runTarget(ctx context.Context, adapter *mcp.Adapter, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing target subcommand")
	}
	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("target add", flag.ContinueOnError)
		req := model.AddTargetRequest{}
		fs.StringVar(&req.ID, "id", "", "")
		fs.StringVar(&req.Name, "name", "", "")
		fs.StringVar(&req.Host, "host", "", "")
		fs.IntVar(&req.Port, "port", 0, "")
		fs.StringVar(&req.User, "user", "", "")
		fs.StringVar(&req.AuthMode, "auth-mode", "", "")
		fs.StringVar(&req.Password, "password", "", "")
		fs.StringVar(&req.PrivateKeyPath, "private-key-path", "", "")
		fs.StringVar(&req.KnownHostsPolicy, "known-hosts-policy", "", "")
		fs.StringVar(&req.DefaultBaseDir, "default-base-dir", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		resp, err := adapter.AddTarget(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(resp)
	case "list":
		resp, err := adapter.ListTargets(ctx)
		if err != nil {
			return err
		}
		return printJSON(resp)
	default:
		return fmt.Errorf("unsupported target subcommand %q", args[0])
	}
}

func runSync(ctx context.Context, adapter *mcp.Adapter, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing sync subcommand")
	}
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	req := model.SyncRequest{}
	fs.StringVar(&req.TargetID, "target", "", "")
	fs.StringVar(&req.LocalPath, "local", "", "")
	fs.StringVar(&req.RemotePath, "remote", "", "")
	fs.StringVar(&req.ConflictPolicy, "conflict", "", "")
	fs.StringVar(&req.BackendPreference, "backend", "", "")
	fs.BoolVar(&req.CreateDirs, "create-dirs", true, "")
	fs.BoolVar(&req.Resume, "resume", false, "")
	overwrite := true
	fs.BoolVar(&overwrite, "overwrite", true, "")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	req.Direction = strings.TrimSpace(args[0])
	req.Overwrite = &overwrite
	resp, err := adapter.StartSync(ctx, req)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func runExec(ctx context.Context, adapter *mcp.Adapter, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing exec subcommand")
	}
	switch args[0] {
	case "start":
		fs := flag.NewFlagSet("exec start", flag.ContinueOnError)
		req := model.ExecRequest{}
		fs.StringVar(&req.TargetID, "target", "", "")
		fs.StringVar(&req.Command, "command", "", "")
		fs.StringVar(&req.WorkDir, "workdir", "", "")
		fs.BoolVar(&req.PTY, "pty", false, "")
		fs.IntVar(&req.TimeoutSeconds, "timeout", 0, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		resp, err := adapter.StartExec(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(resp)
	case "read":
		fs := flag.NewFlagSet("exec read", flag.ContinueOnError)
		req := model.ExecReadRequest{}
		fs.StringVar(&req.SessionID, "session", "", "")
		fs.StringVar(&req.Cursor, "cursor", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		resp, err := adapter.ReadExec(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(resp)
	case "stop":
		fs := flag.NewFlagSet("exec stop", flag.ContinueOnError)
		req := model.ExecStopRequest{}
		fs.StringVar(&req.SessionID, "session", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		resp, err := adapter.StopExec(ctx, req)
		if err != nil {
			return err
		}
		return printJSON(resp)
	default:
		return fmt.Errorf("unsupported exec subcommand %q", args[0])
	}
}

func runJob(ctx context.Context, adapter *mcp.Adapter, args []string) error {
	if len(args) == 0 || args[0] != "status" {
		return fmt.Errorf("unsupported job subcommand")
	}
	fs := flag.NewFlagSet("job status", flag.ContinueOnError)
	req := model.JobStatusRequest{}
	fs.StringVar(&req.ID, "id", "", "")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	resp, err := adapter.GetJob(ctx, req)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
