package symphony

import (
"errors"
"os"
"path/filepath"
"testing"
)

func TestLoadWorkflow_WithFrontMatter(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "WORKFLOW.md")
content := "---\ntracker:\n  kind: linear\n---\n\nHello {{index .issue \"identifier\"}}"
if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
t.Fatal(err)
}

wf, err := LoadWorkflow(path)
if err != nil {
t.Fatalf("LoadWorkflow failed: %v", err)
}
if wf.Config["tracker"] == nil {
t.Fatalf("expected parsed tracker config")
}
if wf.PromptTemplate != "Hello {{index .issue \"identifier\"}}" {
t.Fatalf("unexpected template: %q", wf.PromptTemplate)
}
}

func TestLoadWorkflow_NonMapFrontMatter(t *testing.T) {
dir := t.TempDir()
path := filepath.Join(dir, "WORKFLOW.md")
if err := os.WriteFile(path, []byte("---\n- one\n- two\n---\nbody"), 0o644); err != nil {
t.Fatal(err)
}

_, err := LoadWorkflow(path)
if !errors.Is(err, ErrWorkflowFrontMatterMap) {
t.Fatalf("expected ErrWorkflowFrontMatterMap, got %v", err)
}
}

func TestRenderPrompt_StrictUnknownVariable(t *testing.T) {
_, err := RenderPrompt("{{.issue.missing_key}}", Issue{Identifier: "ABC-1"}, nil)
if !errors.Is(err, ErrTemplateRender) {
t.Fatalf("expected ErrTemplateRender, got %v", err)
}
}
