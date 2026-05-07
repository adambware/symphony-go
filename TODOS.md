# Symphony spec TODOs

This file tracks major portions of the Symphony spec that are not yet implemented in this repository.

## Core runtime / orchestration

- [ ] Polling orchestrator with in-memory runtime state (`running`, `claimed`, `retry_attempts`, token totals, rate limits)
- [ ] Dispatch loop with candidate sorting, slot accounting, and eligibility filtering
- [ ] Retry queue with continuation retry + exponential backoff behavior
- [ ] Active-run reconciliation (stall detection + tracker state refresh)
- [ ] Startup terminal workspace cleanup
- [ ] Dynamic `WORKFLOW.md` reload/watch with last-known-good fallback

## Tracker integration (Linear)

- [ ] Linear client for candidate fetch, issue-state refresh, and terminal-state fetch
- [ ] Linear GraphQL pagination and normalized issue mapping
- [ ] Tracker error mapping and operational handling semantics from the spec

## Agent runner / Codex integration

- [ ] Agent attempt lifecycle (workspace prep, prompt build, launch, streaming, finish)
- [ ] Codex app-server subprocess startup (`bash -lc <codex.command>`) in per-issue workspace
- [ ] Session/thread/turn identity extraction and session metadata tracking
- [ ] Turn timeout/read-timeout/stall-timeout handling and normalized error mapping
- [ ] Continuation turns on same thread up to `agent.max_turns`
- [ ] Structured event forwarding from app-server updates to orchestrator
- [ ] Documented approval/sandbox/user-input-required policy handling

## Optional extensions

- [ ] Optional status surface/runtime snapshot API
- [ ] Optional HTTP extension (`/`, `/api/v1/state`, `/api/v1/<issue_identifier>`, `/api/v1/refresh`)
- [ ] Optional `linear_graphql` client-side tool extension
- [ ] Optional SSH worker extension

## CLI / host lifecycle

- [ ] CLI startup wiring with positional workflow path support and startup failure behavior
- [ ] Service lifecycle management (startup validation, scheduling, graceful shutdown)

## Observability and logging

- [ ] Structured runtime logs with issue/session context across orchestrator + worker lifecycle
- [ ] Aggregate token/rate-limit accounting with live runtime totals
- [ ] Operator-visible validation/dispatch/reconciliation failures

## Additional validation coverage

- [ ] Expand tests to cover orchestration state transitions and retries
- [ ] Expand tests to cover tracker normalization/pagination/error contracts
- [ ] Expand tests to cover agent-runner protocol behavior and timeout/error paths
