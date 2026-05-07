package symphony

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type TrackerConfig struct {
	Kind           string
	Endpoint       string
	APIKey         string
	ProjectSlug    string
	ActiveStates   []string
	TerminalStates []string
}

type PollingConfig struct {
	Interval time.Duration
}

type HooksConfig struct {
	AfterCreate  string
	BeforeRun    string
	AfterRun     string
	BeforeRemove string
	Timeout      time.Duration
}

type AgentConfig struct {
	MaxConcurrentAgents        int
	MaxTurns                   int
	MaxRetryBackoff            time.Duration
	MaxConcurrentAgentsByState map[string]int
}

type CodexConfig struct {
	Command           string
	ApprovalPolicy    string
	ThreadSandbox     string
	TurnSandboxPolicy string
	TurnTimeout       time.Duration
	ReadTimeout       time.Duration
	StallTimeout      time.Duration
}

type ServiceConfig struct {
	Tracker   TrackerConfig
	Polling   PollingConfig
	Workspace WorkspaceConfig
	Hooks     HooksConfig
	Agent     AgentConfig
	Codex     CodexConfig
}

type WorkspaceConfig struct {
	Root string
}

func ResolveServiceConfig(def WorkflowDefinition, workflowPath string) (ServiceConfig, error) {
	cfg := ServiceConfig{
		Tracker: TrackerConfig{
			Endpoint:       "https://api.linear.app/graphql",
			ActiveStates:   []string{"Todo", "In Progress"},
			TerminalStates: []string{"Closed", "Cancelled", "Canceled", "Duplicate", "Done"},
		},
		Polling:   PollingConfig{Interval: 30 * time.Second},
		Workspace: WorkspaceConfig{Root: filepath.Join(os.TempDir(), "symphony_workspaces")},
		Hooks:     HooksConfig{Timeout: 60 * time.Second},
		Agent: AgentConfig{
			MaxConcurrentAgents:        10,
			MaxTurns:                   20,
			MaxRetryBackoff:            5 * time.Minute,
			MaxConcurrentAgentsByState: map[string]int{},
		},
		Codex: CodexConfig{
			Command:      "codex app-server",
			TurnTimeout:  time.Hour,
			ReadTimeout:  5 * time.Second,
			StallTimeout: 5 * time.Minute,
		},
	}

	tracker := getMap(def.Config, "tracker")
	polling := getMap(def.Config, "polling")
	workspace := getMap(def.Config, "workspace")
	hooks := getMap(def.Config, "hooks")
	agent := getMap(def.Config, "agent")
	codex := getMap(def.Config, "codex")

	cfg.Tracker.Kind = stringOr(tracker, "kind", cfg.Tracker.Kind)
	cfg.Tracker.Endpoint = stringOr(tracker, "endpoint", cfg.Tracker.Endpoint)
	cfg.Tracker.APIKey = resolveEnvToken(stringOr(tracker, "api_key", cfg.Tracker.APIKey))
	if cfg.Tracker.APIKey == "" && strings.EqualFold(cfg.Tracker.Kind, "linear") {
		cfg.Tracker.APIKey = os.Getenv("LINEAR_API_KEY")
	}
	cfg.Tracker.ProjectSlug = stringOr(tracker, "project_slug", cfg.Tracker.ProjectSlug)
	cfg.Tracker.ActiveStates = stringSliceOr(tracker, "active_states", cfg.Tracker.ActiveStates)
	cfg.Tracker.TerminalStates = stringSliceOr(tracker, "terminal_states", cfg.Tracker.TerminalStates)

	cfg.Polling.Interval = millisecondsOr(polling, "interval_ms", cfg.Polling.Interval)
	cfg.Workspace.Root = expandPath(pathValue(workspace, "root", cfg.Workspace.Root), filepath.Dir(workflowPath))

	cfg.Hooks.AfterCreate = stringOr(hooks, "after_create", cfg.Hooks.AfterCreate)
	cfg.Hooks.BeforeRun = stringOr(hooks, "before_run", cfg.Hooks.BeforeRun)
	cfg.Hooks.AfterRun = stringOr(hooks, "after_run", cfg.Hooks.AfterRun)
	cfg.Hooks.BeforeRemove = stringOr(hooks, "before_remove", cfg.Hooks.BeforeRemove)
	cfg.Hooks.Timeout = millisecondsOr(hooks, "timeout_ms", cfg.Hooks.Timeout)

	cfg.Agent.MaxConcurrentAgents = intOr(agent, "max_concurrent_agents", cfg.Agent.MaxConcurrentAgents)
	cfg.Agent.MaxTurns = intOr(agent, "max_turns", cfg.Agent.MaxTurns)
	cfg.Agent.MaxRetryBackoff = millisecondsOr(agent, "max_retry_backoff_ms", cfg.Agent.MaxRetryBackoff)
	cfg.Agent.MaxConcurrentAgentsByState = normalizePerStateConcurrency(getMap(agent, "max_concurrent_agents_by_state"))

	cfg.Codex.Command = stringOr(codex, "command", cfg.Codex.Command)
	cfg.Codex.ApprovalPolicy = stringOr(codex, "approval_policy", cfg.Codex.ApprovalPolicy)
	cfg.Codex.ThreadSandbox = stringOr(codex, "thread_sandbox", cfg.Codex.ThreadSandbox)
	cfg.Codex.TurnSandboxPolicy = stringOr(codex, "turn_sandbox_policy", cfg.Codex.TurnSandboxPolicy)
	cfg.Codex.TurnTimeout = millisecondsOr(codex, "turn_timeout_ms", cfg.Codex.TurnTimeout)
	cfg.Codex.ReadTimeout = millisecondsOr(codex, "read_timeout_ms", cfg.Codex.ReadTimeout)
	cfg.Codex.StallTimeout = millisecondsOr(codex, "stall_timeout_ms", cfg.Codex.StallTimeout)

	if err := ValidateDispatchConfig(cfg); err != nil {
		return ServiceConfig{}, err
	}
	return cfg, nil
}

func ValidateDispatchConfig(cfg ServiceConfig) error {
	if !strings.EqualFold(strings.TrimSpace(cfg.Tracker.Kind), "linear") {
		return ErrUnsupportedTrackerKind
	}
	if strings.TrimSpace(cfg.Tracker.APIKey) == "" {
		return ErrMissingTrackerAPIKey
	}
	if strings.TrimSpace(cfg.Tracker.ProjectSlug) == "" {
		return ErrMissingTrackerProject
	}
	if strings.TrimSpace(cfg.Codex.Command) == "" {
		return ErrMissingCodexCommand
	}
	if cfg.Hooks.Timeout <= 0 {
		return ErrInvalidHookTimeout
	}
	if cfg.Agent.MaxTurns <= 0 {
		return ErrInvalidAgentMaxTurns
	}
	if cfg.Polling.Interval <= 0 {
		return ErrInvalidPollingInterval
	}
	if cfg.Agent.MaxConcurrentAgents <= 0 {
		return ErrInvalidConcurrentAgents
	}
	if cfg.Agent.MaxRetryBackoff <= 0 {
		return ErrInvalidRetryBackoff
	}
	if !filepath.IsAbs(cfg.Workspace.Root) {
		return fmt.Errorf("%w: root must be absolute", ErrInvalidWorkspacePath)
	}
	return nil
}

func (c ServiceConfig) IsActiveState(state string) bool {
	needle := strings.ToLower(strings.TrimSpace(state))
	for _, s := range c.Tracker.ActiveStates {
		if strings.ToLower(s) == needle {
			return true
		}
	}
	return false
}

func (c ServiceConfig) IsTerminalState(state string) bool {
	needle := strings.ToLower(strings.TrimSpace(state))
	for _, s := range c.Tracker.TerminalStates {
		if strings.ToLower(s) == needle {
			return true
		}
	}
	return false
}

func getMap(root map[string]any, key string) map[string]any {
	v, ok := root[key]
	if !ok {
		return map[string]any{}
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func stringOr(m map[string]any, key, dflt string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return dflt
}

func pathValue(m map[string]any, key, dflt string) string {
	v := stringOr(m, key, dflt)
	if strings.HasPrefix(v, "$") {
		return os.Getenv(strings.TrimPrefix(v, "$"))
	}
	return v
}

func resolveEnvToken(v string) string {
	if strings.HasPrefix(v, "$") {
		return os.Getenv(strings.TrimPrefix(v, "$"))
	}
	return v
}

func intOr(m map[string]any, key string, dflt int) int {
	v, ok := m[key]
	if !ok {
		return dflt
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(n))
		if err == nil {
			return parsed
		}
	}
	return dflt
}

func millisecondsOr(m map[string]any, key string, dflt time.Duration) time.Duration {
	return time.Duration(intOr(m, key, int(dflt/time.Millisecond))) * time.Millisecond
}

func stringSliceOr(m map[string]any, key string, dflt []string) []string {
	v, ok := m[key]
	if !ok {
		return append([]string(nil), dflt...)
	}
	arr, ok := v.([]any)
	if !ok {
		if strArr, ok := v.([]string); ok {
			return append([]string(nil), strArr...)
		}
		return append([]string(nil), dflt...)
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), dflt...)
	}
	return out
}

func normalizePerStateConcurrency(m map[string]any) map[string]int {
	out := map[string]int{}
	for k, raw := range m {
		var n int
		switch v := raw.(type) {
		case int:
			n = v
		case int64:
			n = int(v)
		case float64:
			n = int(v)
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				continue
			}
			n = parsed
		}
		if n <= 0 {
			continue
		}
		out[strings.ToLower(strings.TrimSpace(k))] = n
	}
	return out
}

func expandPath(raw, workflowDir string) string {
	if strings.TrimSpace(raw) == "" {
		raw = filepath.Join(os.TempDir(), "symphony_workspaces")
	}
	if strings.HasPrefix(raw, "~/") || raw == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			raw = filepath.Join(home, strings.TrimPrefix(raw, "~/"))
		}
	}
	if !filepath.IsAbs(raw) {
		raw = filepath.Join(workflowDir, raw)
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return raw
	}
	return abs
}

func IsErr(err, target error) bool {
	return errors.Is(err, target)
}
