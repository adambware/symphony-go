package symphony

import "errors"

var (
	ErrMissingWorkflowFile     = errors.New("missing_workflow_file")
	ErrWorkflowParse           = errors.New("workflow_parse_error")
	ErrWorkflowFrontMatterMap  = errors.New("workflow_front_matter_not_a_map")
	ErrTemplateParse           = errors.New("template_parse_error")
	ErrTemplateRender          = errors.New("template_render_error")
	ErrUnsupportedTrackerKind  = errors.New("unsupported_tracker_kind")
	ErrMissingTrackerAPIKey    = errors.New("missing_tracker_api_key")
	ErrMissingTrackerProject   = errors.New("missing_tracker_project_slug")
	ErrMissingCodexCommand     = errors.New("missing_codex_command")
	ErrInvalidWorkspacePath    = errors.New("invalid_workspace_path")
	ErrInvalidHookTimeout      = errors.New("invalid_hook_timeout")
	ErrInvalidAgentMaxTurns    = errors.New("invalid_agent_max_turns")
	ErrInvalidPollingInterval  = errors.New("invalid_polling_interval")
	ErrInvalidConcurrentAgents = errors.New("invalid_max_concurrent_agents")
	ErrInvalidRetryBackoff     = errors.New("invalid_max_retry_backoff_ms")
)
