package symphony

import (
	"strings"
	"time"
)

type BlockerRef struct {
	ID         *string `json:"id,omitempty"`
	Identifier *string `json:"identifier,omitempty"`
	State      *string `json:"state,omitempty"`
}

type Issue struct {
	ID          string       `json:"id"`
	Identifier  string       `json:"identifier"`
	Title       string       `json:"title"`
	Description *string      `json:"description,omitempty"`
	Priority    *int         `json:"priority,omitempty"`
	State       string       `json:"state"`
	BranchName  *string      `json:"branch_name,omitempty"`
	URL         *string      `json:"url,omitempty"`
	Labels      []string     `json:"labels,omitempty"`
	BlockedBy   []BlockerRef `json:"blocked_by,omitempty"`
	CreatedAt   *time.Time   `json:"created_at,omitempty"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
}

func (i Issue) NormalizedLabels() []string {
	out := make([]string, len(i.Labels))
	for idx, l := range i.Labels {
		out[idx] = strings.ToLower(l)
	}
	return out
}

func (i Issue) TemplateMap() map[string]any {
	return map[string]any{
		"id":          i.ID,
		"identifier":  i.Identifier,
		"title":       i.Title,
		"description": i.Description,
		"priority":    i.Priority,
		"state":       i.State,
		"branch_name": i.BranchName,
		"url":         i.URL,
		"labels":      i.NormalizedLabels(),
		"blocked_by":  i.BlockedBy,
		"created_at":  i.CreatedAt,
		"updated_at":  i.UpdatedAt,
	}
}
