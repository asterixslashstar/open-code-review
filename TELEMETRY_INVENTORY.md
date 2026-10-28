# Telemetry & Network Outbound Data — Inventory

> **Generated:** 2026-10-26  
> **Branch:** `remove-telemetry`

This document lists every location in the codebase that can send data off the local machine, either through telemetry exports, update checks, or API calls. Each entry includes the file path, what data is sent, and how to disable it.

---

## 1. OpenTelemetry Exporter (Primary Telemetry)

### 1.1 Core Telemetry Package

| File | Purpose |
|------|---------|
| `internal/telemetry/config.go` | Configuration resolution — reads `OCR_ENABLE_TELEMETRY`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OCR_CONTENT_LOGGING` env vars and `~/.opencodereview/config.json` telemetry section |
| `internal/telemetry/provider.go` | Initializes global OTel TracerProvider & MeterProvider; collects **host info, OS info, process info** via `resource.WithHost()`, `resource.WithOS()`, `resource.WithProcess()`, `resource.WithFromEnv()` |
| `internal/telemetry/exporter.go` | Creates OTLP exporters (gRPC + HTTP) for traces and metrics; creates console exporters as fallback |
| `internal/telemetry/span.go` | Span creation helpers — wraps operations in trace spans |
| `internal/telemetry/metrics.go` | Metric definitions and recording functions |
| `internal/telemetry/events.go` | Structured event emission (short-lived spans) |
| `internal/telemetry/shutdown.go` | Flushes and shuts down providers on process exit |

### 1.2 What Data Is Collected & Sent

**Resource attributes** (attached to every span/metric):
- Host name (`resource.WithHost()`)
- OS type/version (`resource.WithOS()`)
- Process name/PID/start time (`resource.WithProcess()`)
- Environment variables matching OTEL patterns (`resource.WithFromEnv()`)

**Metrics exported** (periodic batch, default interval per OTel SDK):
| Metric Name | Type | Description | Data Sent |
|-------------|------|-------------|-----------|
| `ocr.review.duration_seconds` | Histogram | Total duration of a code review run | Duration in seconds |
| `ocr.files_reviewed_total` | Counter | Number of files reviewed | Integer count |
| `ocr.comments_generated_total` | Counter | Number of review comments generated | Integer count |
| `ocr.llm.requests_total` | Counter | Total LLM API requests made | Count + model name + status |
| `ocr.llm.tokens_used` | Counter | Tokens consumed by LLM requests | Integer count |
| `ocr.llm.request_duration_seconds` | Histogram | Duration of individual LLM API requests | Duration + model name |
| `ocr.tool.calls_total` | Counter | Total tool calls made | Count + tool name + success/failure |
| `ocr.tool.execution_duration_seconds` | Histogram | Duration of tool executions | Duration + tool name + success/failure |

**Trace spans exported** (batched):
| Span Name Pattern | Where Emitted | Attributes Sent |
|-------------------|---------------|-----------------|
| `review.run` | `cmd/opencodereview/shared.go` | Review duration |
| `diff.parse` | `internal/agent/agent.go` | Files changed count, lines inserted/deleted |
| `event.review.skipped` | `internal/agent/agent.go` | Reason, file count, too-large count |
| `event.no.files.changed` | `internal/agent/agent.go` | — |
| `event.review.started` | `internal/agent/agent.go` | File count, review count, repo dir |
| `event.scan.enumerate` | `internal/scan/agent.go` | Files enumerated count |
| `event.scan.no.files` | `internal/scan/agent.go` | — |
| `event.scan.started` | `internal/scan/agent.go` | Reviewable file count |
| `event.plan.skipped` | `internal/llmloop/loop.go` | Reason |
| `event.token.threshold.exceeded` | `internal/llmloop/loop.go` | Token counts |
| `event.subtask.error` | `internal/llmloop/loop.go` | Error message |
| `subtask.execute.group.<key>` | `internal/llmloop/loop.go` | Group key, file paths |
| `llm.request` | `internal/llmloop/loop.go`, `internal/agent/agent.go` | Model name, duration, tokens, status |
| `tool.execute.<name>` | `internal/llmloop/loop.go` | Tool name, duration_ms, status, error |

**Note:** Prompt and response content is **never** attached to telemetry spans (documented in `pages/src/content/docs/en/telemetry.md`). The `OCR_CONTENT_LOGGING` flag exists but is currently dead code.

### 1.3 Export Protocols & Endpoints

| Protocol | Traces Endpoint | Metrics Endpoint | File |
|----------|----------------|------------------|------|
| OTLP/gRPC (default) | `{endpoint}` (port 4317) | `{endpoint}` (port 4317) | `internal/telemetry/exporter.go` |
| OTLP/http/protobuf | `{endpoint}/v1/traces` | `{endpoint}/v1/metrics` | `internal/telemetry/exporter.go` |
| OTLP/http/json | `{endpoint}/v1/traces` | `{endpoint}/v1/metrics` | `internal/telemetry/exporter.go` |
| Console (stdout) | Pretty-printed JSON to stdout | Pretty-printed JSON to stdout | `internal/telemetry/exporter.go` |

### 1.4 How to Disable Telemetry

| Method | Details |
|--------|---------|
| **Default state** | Telemetry is **off by default**. Nothing is sent unless explicitly enabled. |
| **Environment variable** | Set `OCR_ENABLE_TELEMETRY=1` to enable; unset or any other value keeps it off |
| **Config file** | `~/.opencodereview/config.json` → `"telemetry": { "enabled": false }` (default) |
| **CLI** | `ocr config set telemetry.enabled false` |
| **Remove dependency** | Remove `go.opentelemetry.io/otel*` from `go.mod`; delete `internal/telemetry/` directory; remove imports from callers |

### 1.5 Callers That Use Telemetry

| File | Telemetry Functions Called |
|------|---------------------------|
| `cmd/opencodereview/main.go` | `telemetry.Init()`, `telemetry.ShutdownWithTimeout()` |
| `cmd/opencodereview/shared.go` | `RecordReviewDuration()`, `ContextWithTraceParentFromEnv()`, `TraceIDFromContext()` |
| `cmd/opencodereview/review_cmd.go` | `telemetry.Init()`, span creation, `EndSpan()` |
| `cmd/opencodereview/scan_cmd.go` | `telemetry.Init()`, span creation |
| `internal/agent/agent.go` | `StartSpan()`, `SetAttr()`, `Event()`, `RecordFilesReviewed()`, `RecordLLMResult()`, `StartToolSpan()`, `RecordToolResult()` |
| `internal/agent/grouping.go` | `Event()` (skip reason) |
| `internal/diff/relocation.go` | `StartLLMSpan()`, `RecordLLMResult()` |
| `internal/llmloop/loop.go` | `StartLLMSpan()`, `RecordLLMResult()`, `RecordLLMRequest()`, `RecordToolCall()`, `Event()` |
| `internal/scan/agent.go` | `StartSpan()`, `SetAttr()`, `Event()`, `RecordFilesReviewed()`, `RecordReviewDuration()` |

---

## 2. NPM Wrapper Update Checks

| File | What It Does | Data Sent | How to Disable |
|------|-------------|-----------|----------------|
| `bin/ocr.js` (line ~91) | Background update check via npm registry; writes `last-update-check`, `update.lock`, `update-available` to `~/.opencodereview/` | Checks npm registry for newer version of `@alibaba/open-code-review` | Set `OCR_NO_UPDATE=1` env var; set `OCR_UPDATE_INTERVAL` to tune interval |
| `scripts/update.js` (line ~19) | Same update-check logic (used by install script) | Checks npm registry | Same as above |

---

## 3. Install Scripts (GitHub API — Download Only)

These only **fetch** release assets; they do not send telemetry data back:

| File | What It Does |
|------|-------------|
| `install.sh` (line ~38) | Calls `https://api.github.com/repos/$REPO/releases/latest` to find the latest release asset |
| `install.ps1` (line ~36) | Same for Windows PowerShell |

---

## 4. LLM Provider HTTP Connections (Not Telemetry — User-Configured)

These are **user-configured** endpoints for LLM API calls, not telemetry. They send user code/context to the user's chosen provider:

| File | Purpose |
|------|---------|
| `internal/llm/client.go` | HTTP client for OpenAI-compatible APIs |
| `internal/llm/responses_client.go` | HTTP client for OpenAI responses API |
| `internal/chatgptauth/oauth.go` | OAuth flow for ChatGPT auth |
| `internal/mcp/client.go` | MCP (Model Context Protocol) client HTTP connections |
| `acp/internal/intent/llm/client.go` | ACP intent LLM client HTTP calls |

---

## 5. Files to Delete / Modify to Remove All Telemetry

### Files to delete entirely:
```
internal/telemetry/config.go
internal/telemetry/config_test.go
internal/telemetry/events.go
internal/telemetry/events_test.go
internal/telemetry/exporter.go
internal/telemetry/exporter_test.go
internal/telemetry/metrics.go
internal/telemetry/metrics_test.go
internal/telemetry/provider.go
internal/telemetry/provider_test.go
internal/telemetry/shutdown.go
internal/telemetry/shutdown_test.go
internal/telemetry/span.go
internal/telemetry/span_test.go
internal/telemetry/traceid_test.go
```

### Files to remove telemetry imports/calls from:
```
cmd/opencodereview/main.go
cmd/opencodereview/shared.go
cmd/opencodereview/review_cmd.go
cmd/opencodereview/scan_cmd.go
cmd/opencodereview/config_cmd.go          (TelemetryConfig struct + config keys)
internal/agent/agent.go
internal/agent/grouping.go
internal/diff/relocation.go
internal/llmloop/loop.go
internal/scan/agent.go
```

### go.mod entries to remove:
```
go.opentelemetry.io/otel v1.45.0
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.45.0
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.45.0
(go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc)
(go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp)
(go.opentelemetry.io/otel/exporters/stdout/stdoutmetric)
(go.opentelemetry.io/otel/exporters/stdout/stdouttrace)
(go.opentelemetry.io/otel/sdk v1.45.0)
(go.opentelemetry.io/otel/sdk/metric v1.45.0)
(go.opentelemetry.io/otel/sdk v1.45.0)
(and all related transitive dependencies)
```

### Config file keys to remove:
- `cmd/opencodereview/config_cmd.go`: `TelemetryConfig` struct (line ~380), config keys `"telemetry.enabled"`, `"telemetry.exporter"`, `"telemetry.otlp_endpoint"`, `"telemetry.content_logging"` (lines ~624-627), case handlers in config dispatch

### Documentation files referencing telemetry:
```
pages/src/content/docs/en/telemetry.md
pages/src/content/docs/ja/telemetry.md
pages/src/content/docs/ko/telemetry.md
pages/src/content/docs/ru/telemetry.md
pages/src/content/docs/zh/telemetry.md
pages/src/content/docs/index.ts              (imports telemetry docs)
pages/src/content/docs/en/architecture.md    (telemetry section)
pages/src/content/docs/ja/architecture.md
pages/src/content/docs/ko/architecture.md
pages/src/content/docs/ru/architecture.md
pages/src/content/docs/zh/architecture.md
pages/src/content/docs/en/faq.md             (telemetry FAQ entries)
pages/src/content/docs/en/viewer.md          (viewer + telemetry reference)
pages/src/content/docs/ja/viewer.md
pages/src/content/docs/ko/viewer.md
pages/src/content/docs/ru/viewer.md
pages/src/content/docs/zh/viewer.md
pages/src/i18n/*.ts                          ('docs.configKeyTelemetry' keys)
pages/src/pages/DocsPage.tsx                 (telemetry sidebar entry)
docs/i18n/README.*.md                       (telemetry bullet in feature list)
CONTRIBUTING.md                              (telemetry directory mention)
```

---

## Summary

| Category | Files | Data Sent | Default State |
|----------|-------|-----------|---------------|
| **OpenTelemetry traces** | `internal/telemetry/**` + 10 callers | Spans with operation names, durations, file counts, model names, tool names, error messages | **OFF** (requires `OCR_ENABLE_TELEMETRY=1`) |
| **OpenTelemetry metrics** | `internal/telemetry/metrics.go` + callers | Review duration, files reviewed, comments generated, LLM request counts/tokens/duration, tool call counts/duration | **OFF** (requires `OCR_ENABLE_TELEMETRY=1`) |
| **Resource attributes** | `internal/telemetry/provider.go` | Host name, OS info, process info | **OFF** (only collected when telemetry is enabled) |
| **NPM update checks** | `bin/ocr.js`, `scripts/update.js` | Package version comparison against npm registry | ON (can be disabled with `OCR_NO_UPDATE=1`) |
| **Install scripts** | `install.sh`, `install.ps1` | GitHub release metadata (GET only) | ON (one-time, install-time only) |
