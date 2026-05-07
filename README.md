# symphony-go

A Go implementation scaffold for the [Symphony specification](https://github.com/openai/symphony/blob/main/SPEC.md).

## Current state

Implemented so far:

- `WORKFLOW.md` path resolution and loading
- YAML front matter parsing with typed workflow errors
- Strict prompt rendering with `issue` and `attempt` inputs
- Typed config resolution with defaults, `$VAR` resolution, and path normalization
- Dispatch preflight config validation
- Core issue model normalization helpers
- Workspace key sanitization and root-containment safety checks
- Workspace lifecycle hooks (`after_create`, `before_run`, `after_run`, `before_remove`)

Still missing scope is tracked in [`TODOS.md`](./TODOS.md).

## Quality gates

Run all baseline checks locally:

```bash
make check
```

Individual checks:

```bash
make fmt-check   # formatting
go vet ./...     # lint/code quality
go test ./...    # tests
```
