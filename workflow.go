package symphony

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type WorkflowDefinition struct {
	Config         map[string]any
	PromptTemplate string
}

func ResolveWorkflowPath(explicitPath, cwd string) string {
	if strings.TrimSpace(explicitPath) != "" {
		return explicitPath
	}
	return filepath.Join(cwd, "WORKFLOW.md")
}

func LoadWorkflow(path string) (WorkflowDefinition, error) {
	// #nosec G304 -- workflow path is explicit runtime input and expected to be repository-owned.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return WorkflowDefinition{}, ErrMissingWorkflowFile
		}
		return WorkflowDefinition{}, fmt.Errorf("%w: %v", ErrWorkflowParse, err)
	}

	cfg, body, err := splitWorkflow(data)
	if err != nil {
		return WorkflowDefinition{}, err
	}

	return WorkflowDefinition{Config: cfg, PromptTemplate: strings.TrimSpace(body)}, nil
}

func splitWorkflow(data []byte) (map[string]any, string, error) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return map[string]any{}, text, nil
	}

	rest := strings.TrimPrefix(text, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		if strings.TrimSpace(rest) == "---" {
			return map[string]any{}, "", nil
		}
		return nil, "", fmt.Errorf("%w: missing closing front matter", ErrWorkflowParse)
	}

	fm := rest[:idx]
	body := rest[idx+len("\n---\n"):]
	if strings.TrimSpace(fm) == "" {
		return map[string]any{}, body, nil
	}

	var parsed any
	if err := yaml.Unmarshal([]byte(fm), &parsed); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrWorkflowParse, err)
	}
	asMap, ok := parsed.(map[string]any)
	if !ok {
		return nil, "", ErrWorkflowFrontMatterMap
	}
	return asMap, body, nil
}

func RenderPrompt(templateText string, issue Issue, attempt *int) (string, error) {
	tmpl, err := template.New("workflow").Option("missingkey=error").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}

	input := map[string]any{"issue": issue.TemplateMap(), "attempt": attempt}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, input); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateRender, err)
	}
	return strings.TrimSpace(out.String()), nil
}
