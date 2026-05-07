package symphony

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveServiceConfig_DefaultsAndResolution(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "secret")
	t.Setenv("WSROOT", "relative/workspaces")

	def := WorkflowDefinition{Config: map[string]any{
		"tracker": map[string]any{
			"kind":         "linear",
			"project_slug": "proj",
			"api_key":      "$LINEAR_API_KEY",
		},
		"workspace": map[string]any{"root": "$WSROOT"},
		"agent": map[string]any{
			"max_concurrent_agents_by_state": map[string]any{"In Progress": 2, "bad": 0},
		},
	}}

	cfg, err := ResolveServiceConfig(def, filepath.Join("/tmp", "repo", "WORKFLOW.md"))
	if err != nil {
		t.Fatalf("ResolveServiceConfig failed: %v", err)
	}
	if cfg.Tracker.APIKey != "secret" {
		t.Fatalf("expected API key from env, got %q", cfg.Tracker.APIKey)
	}
	if !filepath.IsAbs(cfg.Workspace.Root) {
		t.Fatalf("workspace root must be absolute: %q", cfg.Workspace.Root)
	}
	if cfg.Agent.MaxConcurrentAgentsByState["in progress"] != 2 {
		t.Fatalf("expected normalized state concurrency map")
	}
	if _, ok := cfg.Agent.MaxConcurrentAgentsByState["bad"]; ok {
		t.Fatalf("expected invalid entry to be ignored")
	}
}

func TestValidateDispatchConfig_MissingProjectSlug(t *testing.T) {
	cfg := ServiceConfig{
		Tracker:   TrackerConfig{Kind: "linear", APIKey: "x"},
		Polling:   PollingConfig{Interval: 1},
		Workspace: WorkspaceConfig{Root: filepath.Clean(string(os.PathSeparator))},
		Hooks:     HooksConfig{Timeout: 1},
		Agent:     AgentConfig{MaxConcurrentAgents: 1, MaxTurns: 1, MaxRetryBackoff: 1},
		Codex:     CodexConfig{Command: "codex app-server"},
	}
	if err := ValidateDispatchConfig(cfg); !errors.Is(err, ErrMissingTrackerProject) {
		t.Fatalf("expected ErrMissingTrackerProject, got %v", err)
	}
}
