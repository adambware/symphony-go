package symphony

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var invalidWorkspaceChars = regexp.MustCompile(`[^A-Za-z0-9._-]`)

type Workspace struct {
	Path         string
	WorkspaceKey string
	CreatedNow   bool
}

type HookRunner func(ctx context.Context, hookName, script, cwd string) error

type WorkspaceManager struct {
	Root    string
	Hooks   HooksConfig
	RunHook HookRunner
}

func SanitizeWorkspaceKey(issueIdentifier string) string {
	return invalidWorkspaceChars.ReplaceAllString(issueIdentifier, "_")
}

func NewWorkspaceManager(root string, hooks HooksConfig) *WorkspaceManager {
	return &WorkspaceManager{Root: root, Hooks: hooks, RunHook: defaultHookRunner}
}

func (wm *WorkspaceManager) EnsureWorkspace(issueIdentifier string) (Workspace, error) {
	key := SanitizeWorkspaceKey(issueIdentifier)
	root, err := filepath.Abs(wm.Root)
	if err != nil {
		return Workspace{}, fmt.Errorf("%w: %v", ErrInvalidWorkspacePath, err)
	}
	path := filepath.Join(root, key)
	if !isPathWithinRoot(root, path) {
		return Workspace{}, ErrInvalidWorkspacePath
	}

	createdNow := false
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return Workspace{}, fmt.Errorf("%w: workspace path exists but is not a directory", ErrInvalidWorkspacePath)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0o750); err != nil {
			return Workspace{}, err
		}
		createdNow = true
	} else {
		return Workspace{}, err
	}

	ws := Workspace{Path: path, WorkspaceKey: key, CreatedNow: createdNow}
	if createdNow && strings.TrimSpace(wm.Hooks.AfterCreate) != "" {
		if err := wm.runHook("after_create", wm.Hooks.AfterCreate, ws.Path, wm.Hooks.Timeout); err != nil {
			return Workspace{}, err
		}
	}
	return ws, nil
}

func (wm *WorkspaceManager) BeforeRun(workspacePath string) error {
	if strings.TrimSpace(wm.Hooks.BeforeRun) == "" {
		return nil
	}
	return wm.runHook("before_run", wm.Hooks.BeforeRun, workspacePath, wm.Hooks.Timeout)
}

func (wm *WorkspaceManager) AfterRun(workspacePath string) {
	if strings.TrimSpace(wm.Hooks.AfterRun) == "" {
		return
	}
	_ = wm.runHook("after_run", wm.Hooks.AfterRun, workspacePath, wm.Hooks.Timeout)
}

func (wm *WorkspaceManager) RemoveWorkspace(issueIdentifier string) error {
	key := SanitizeWorkspaceKey(issueIdentifier)
	path := filepath.Join(wm.Root, key)
	if strings.TrimSpace(wm.Hooks.BeforeRemove) != "" {
		_ = wm.runHook("before_remove", wm.Hooks.BeforeRemove, path, wm.Hooks.Timeout)
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func isPathWithinRoot(root, candidate string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	candAbs, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, candAbs)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "")
}

func (wm *WorkspaceManager) runHook(hookName, script, cwd string, timeout time.Duration) error {
	runner := wm.RunHook
	if runner == nil {
		runner = defaultHookRunner
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return runner(ctx, hookName, script, cwd)
}

func defaultHookRunner(ctx context.Context, _ string, script, cwd string) error {
	// #nosec G204 -- hooks are trusted, repository-owned workflow scripts by specification.
	cmd := exec.CommandContext(ctx, "bash", "-lc", script)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
