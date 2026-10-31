# Plan: Remove OpenTelemetry Metrics References from TELEMETRY_INVENTORY.md

> **Branch:** `remove-telemetry`  
> **Created:** 2026-10-28

This plan removes all references to **OpenTelemetry metrics** from the TELEMETRY_INVENTORY.md document — specifically the "Metrics exported" table in Section 1.2 and the corresponding Summary table row. Other telemetry content (traces, resource attributes) is preserved.

---

## Task 1: Remove "Metrics exported" Table from Section 1.2

Remove only the **Metrics exported** subsection from Section 1.2 ("What Data Is Collected & Sent"), keeping the Resource attributes and Trace spans sections intact.

**Location:** `TELEMETRY_INVENTORY.md`, Section 1.2  
**Content to remove (lines ~34-50):**
```markdown
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

```

**After:** Section 1.2 will have Resource attributes and Trace spans sections, but no Metrics exported table.

---

## Task 2: Remove "OpenTelemetry metrics" Row from Summary Table

Remove the "OpenTelemetry metrics" row from the Summary table at the end of the document.

**Location:** `TELEMETRY_INVENTORY.md`, Summary table (line ~180)  
**Before:**
```markdown
| **OpenTelemetry traces** | ... |
| **OpenTelemetry metrics** | `internal/telemetry/metrics.go` + callers | Review duration, files reviewed, comments generated, LLM request counts/tokens/duration, tool call counts/duration | **OFF** (requires `OCR_ENABLE_TELEMETRY=1`) |
| **Resource attributes** | ... |
```

**After:**
```markdown
| **OpenTelemetry traces** | ... |
| **Resource attributes** | ... |
```

---

## Summary

| Task | File | Change |
|------|------|--------|
| 1 | TELEMETRY_INVENTORY.md | Remove "Metrics exported" table from Section 1.2 (keep Resource attributes + Trace spans) |
| 2 | TELEMETRY_INVENTORY.md | Remove OpenTelemetry metrics row from Summary table |
