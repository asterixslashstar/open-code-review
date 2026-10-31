# Plan: Remove OpenTelemetry Traces from open-code-review

> **Branch:** `remove-telemetry`  
> **Created:** 2026-10-28

This plan removes **OpenTelemetry traces** (spans, trace exporters, trace-related config) from the codebase while preserving metrics and other telemetry infrastructure.

---

## Phase 1: Delete Trace-Specific Files

### Task 1.1 — Delete `internal/telemetry/span.go`
- This file provides `StartSpan()`, `EndSpan()`, `SetAttr()` helpers used by callers.
- **Impact:** All callers that use span creation will need those calls removed.

### Task 1.2 — Delete `internal/telemetry/events.go`
- This file provides `Event()` for structured event emission (short-lived spans).
- **Impact:** Callers using `Event()` will need those calls removed.

### Task 1.3 — Delete trace exporter code from `internal/telemetry/exporter.go`
- Remove gRPC/HTTP trace exporter creation (`otlptracegrpc`, `otlptracehttp`).
- Remove console trace exporter (`stdout/stdouttrace`).
- Keep metrics exporter code intact.

### Task 1.4 — Delete trace test files
- `internal/telemetry/span_test.go`
- `internal/telemetry/events_test.go`
- `internal/telemetry/exporter_test.go` (trim to keep metrics tests)

---

## Phase 2: Remove Trace Imports & Calls from Callers

### Task 2.1 — `cmd/opencodereview/main.go`
- Remove `telemetry.Init()` call if it only initializes traces.
- Keep `telemetry.ShutdownWithTimeout()` if it also flushes metrics.

### Task 2.2 — `cmd/opencodereview/shared.go`
- Remove `ContextWithTraceParentFromEnv()` function/usage.
- Remove `TraceIDFromContext()` function/usage.
- Remove `RecordReviewDuration()` if it's trace-only (verify).

### Task 2.3 — `cmd/opencodereview/review_cmd.go`
- Remove `telemetry.Init()` call.
- Remove span creation (`StartSpan`, `EndSpan`).

### Task 2.4 — `cmd/opencodereview/scan_cmd.go`
- Remove `telemetry.Init()` call.
- Remove span creation.

### Task 2.5 — `internal/agent/agent.go`
- Remove `StartSpan()`, `SetAttr()`, `Event()` calls.
- Keep `RecordFilesReviewed()`, `RecordLLMResult()`, `StartToolSpan()`, `RecordToolResult()` only if they are metrics-only (verify each).

### Task 2.6 — `internal/agent/grouping.go`
- Remove `Event()` call (skip reason).

### Task 2.7 — `internal/diff/relocation.go`
- Remove `StartLLMSpan()`, `RecordLLMResult()` calls.

### Task 2.8 — `internal/llmloop/loop.go`
- Remove `StartLLMSpan()`, `RecordLLMResult()`, `RecordLLMRequest()`, `RecordToolCall()`, `Event()` calls.

### Task 2.9 — `internal/scan/agent.go`
- Remove `StartSpan()`, `SetAttr()`, `Event()` calls.
- Keep `RecordFilesReviewed()`, `RecordReviewDuration()` only if metrics-only.

---

## Phase 3: Clean Up Telemetry Package API

### Task 3.1 — Trim `internal/telemetry/provider.go`
- Remove `TracerProvider` initialization.
- Keep `MeterProvider` initialization for metrics.

### Task 3.2 — Trim `internal/telemetry/config.go`
- Remove trace-related config keys (`telemetry.exporter`, `telemetry.otlp_endpoint` if trace-only).
- Keep metrics config keys.

### Task 3.3 — Trim `internal/telemetry/metrics.go`
- Verify all functions are metrics-only; remove any that were trace-mixed.

### Task 3.4 — Update `internal/telemetry/shutdown.go`
- Remove trace provider shutdown logic.
- Keep metrics provider shutdown.

---

## Phase 4: Update go.mod Dependencies

### Task 4.1 — Remove trace exporter modules from `go.mod` / `go.sum`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
- `go.opentelemetry.io/otel/exporters/stdout/stdouttrace`
- Run `go mod tidy` to clean up unused transitive deps.

---

## Phase 5: Update Documentation

### Task 5.1 — Delete telemetry doc pages
- `pages/src/content/docs/en/telemetry.md`
- `pages/src/content/docs/ja/telemetry.md`
- `pages/src/content/docs/ko/telemetry.md`
- `pages/src/content/docs/ru/telemetry.md`
- `pages/src/content/docs/zh/telemetry.md`

### Task 5.2 — Remove telemetry references from architecture docs
- Remove telemetry sections from all locale versions of `architecture.md`.

### Task 5.3 — Remove telemetry FAQ entries from `faq.md`

### Task 5.4 — Update viewer docs to remove trace references

### Task 5.5 — Update i18n files (`pages/src/i18n/*.ts`)
- Remove `docs.configKeyTelemetry` keys.

### Task 5.6 — Update DocsPage.tsx sidebar
- Remove telemetry sidebar entry.

### Task 5.7 — Update READMEs and CONTRIBUTING.md
- Remove telemetry bullets from localized READMEs.
- Remove telemetry directory mention from CONTRIBUTING.md.

---

## Phase 6: Verify & Finalize

### Task 6.1 — Run `make check` to verify formatting, vet, and English-only compliance.

### Task 6.2 — Run `make test` to ensure no compilation errors or test failures.

### Task 6.3 — Run `go mod tidy` and verify `go build` succeeds.

### Task 6.4 — Review git diff for any missed references.
