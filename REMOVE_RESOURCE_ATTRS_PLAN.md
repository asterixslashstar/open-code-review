# Plan: Remove Resource Attributes References from TELEMETRY_INVENTORY.md

> **Branch:** `remove-telemetry`  
> **Created:** 2026-10-28

This plan removes references to **resource attributes** (host name, OS info, process info) from the TELEMETRY_INVENTORY.md document.

---

## Task 1: Remove Resource Attributes Section from 1.2

Remove the "Resource attributes" bullet list from section 1.2 "What Data Is Collected & Sent".

**Location:** `TELEMETRY_INVENTORY.md`, lines ~30-34  
**Before:**
```markdown
**Resource attributes** (attached to every metric):
- Host name (`resource.WithHost()`)
- OS type/version (`resource.WithOS()`)
- Process name/PID/start time (`resource.WithProcess()`)
- Environment variables matching OTEL patterns (`resource.WithFromEnv()`)
```

**After:** (remove these 5 lines entirely)

---

## Task 2: Remove Resource Attributes Row from Summary Table

Remove the "Resource attributes" row from the Summary table at the end of the document.

**Location:** `TELEMETRY_INVENTORY.md`, Summary table  
**Before:**
```markdown
| **Resource attributes** | `internal/telemetry/provider.go` | Host name, OS info, process info | **OFF** (only collected when telemetry is enabled) |
```

**After:** (remove this row entirely)

---

## Task 3: Update Section 1.1 Description

Update the provider.go description in section 1.1 to remove reference to resource.WithHost/WithOS/WithProcess calls.

**Location:** `TELEMETRY_INVENTORY.md`, section 1.1 table  
**Before:**
```markdown
| `internal/telemetry/provider.go` | Initializes global OTel MeterProvider; collects **host info, OS info, process info** via `resource.WithHost()`, `resource.WithOS()`, `resource.WithProcess()`, `resource.WithFromEnv()` |
```

**After:**
```markdown
| `internal/telemetry/provider.go` | Initializes global OTel MeterProvider |
```

---

## Task 4: Update Section 1.5 Callers Table

Update the callers table in section 1.5 to remove provider.go from the list (since it no longer collects resource attributes).

**Location:** `TELEMETRY_INVENTORY.md`, section 1.5  
**Before:**
```markdown
| `internal/telemetry/provider.go` | Host name, OS info, process info | **OFF** (only collected when telemetry is enabled) |
```

Wait — provider.go is not in the callers table. Let me check... Actually provider.go doesn't appear in section 1.5. No change needed here.

---

## Summary

| Task | File | Change |
|------|------|--------|
| 1 | TELEMETRY_INVENTORY.md | Remove resource attributes bullet list from section 1.2 |
| 2 | TELEMETRY_INVENTORY.md | Remove resource attributes row from Summary table |
| 3 | TELEMETRY_INVENTORY.md | Simplify provider.go description in section 1.1 |
