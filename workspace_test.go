package symphony

import (
"context"
"path/filepath"
"testing"
)

func TestSanitizeWorkspaceKey(t *testing.T) {
got := SanitizeWorkspaceKey("ABC/123 test")
if got != "ABC_123_test" {
t.Fatalf("unexpected sanitized key: %q", got)
}
}

func TestWorkspaceManager_AfterCreateRunsOnlyOnFirstCreate(t *testing.T) {
tmp := t.TempDir()
count := 0
wm := NewWorkspaceManager(tmp, HooksConfig{AfterCreate: "echo hi", Timeout: 1000000000})
wm.RunHook = func(ctx context.Context, hookName, script, cwd string) error {
count++
if hookName != "after_create" {
t.Fatalf("unexpected hook name: %s", hookName)
}
if cwd != filepath.Join(tmp, "ABC-1") {
t.Fatalf("unexpected cwd: %s", cwd)
}
return nil
}

ws1, err := wm.EnsureWorkspace("ABC-1")
if err != nil {
t.Fatalf("EnsureWorkspace first failed: %v", err)
}
if !ws1.CreatedNow {
t.Fatalf("expected CreatedNow=true on first call")
}
ws2, err := wm.EnsureWorkspace("ABC-1")
if err != nil {
t.Fatalf("EnsureWorkspace second failed: %v", err)
}
if ws2.CreatedNow {
t.Fatalf("expected CreatedNow=false on second call")
}
if count != 1 {
t.Fatalf("expected hook once, got %d", count)
}
}
