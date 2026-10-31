# Plan: Remove NPM Update Checks References from TELEMETRY_INVENTORY.md

> **Branch:** `remove-telemetry`  
> **Created:** 2026-10-28

This plan removes references to **NPM update checks** from the TELEMETRY_INVENTORY.md document.

---

## Task 1: Remove Section 2 Entirely

Remove the entire "## 2. NPM Wrapper Update Checks" section including its table and separator lines.

**Location:** `TELEMETRY_INVENTORY.md`, Section 2  
**Content to remove (lines ~70-80):**
```markdown
---

## 2. NPM Wrapper Update Checks

| File | What It Does | Data Sent | How to Disable |
|------|-------------|-----------|----------------|
| `bin/ocr.js` (line ~91) | Background update check via npm registry; writes `last-update-check`, `update.lock`, `update-available` to `~/.opencodereview/` | Checks npm registry for newer version of `@alibaba/open-code-review` | Set `OCR_NO_UPDATE=1` env var; set `OCR_UPDATE_INTERVAL` to tune interval |
| `scripts/update.js` (line ~19) | Same update-check logic (used by install script) | Checks npm registry | Same as above |

---
```

**After:** Section 3 becomes the next section (renumbered to 2).

---

## Task 2: Renumber Remaining Sections

After removing Section 2, renumber all subsequent sections:
- Old "## 3. Install Scripts" → "## 2. Install Scripts"
- Old "## 4. LLM Provider HTTP Connections" → "## 3. LLM Provider HTTP Connections"
- Old "## 5. Files to Delete / Modify..." → "## 4. Files to Delete / Modify..."

---

## Task 3: Remove NPM Update Checks Row from Summary Table

Remove the "NPM update checks" row from the Summary table at the end of the document.

**Location:** `TELEMETRY_INVENTORY.md`, Summary table  
**Before:**
```markdown
| **NPM update checks** | `bin/ocr.js`, `scripts/update.js` | Package version comparison against npm registry | ON (can be disabled with `OCR_NO_UPDATE=1`) |
```

**After:** (remove this row entirely)

---

## Summary

| Task | File | Change |
|------|------|--------|
| 1 | TELEMETRY_INVENTORY.md | Remove entire Section 2 (NPM Wrapper Update Checks) |
| 2 | TELEMETRY_INVENTORY.md | Renumber sections 3→2, 4→3, 5→4 |
| 3 | TELEMETRY_INVENTORY.md | Remove NPM update checks row from Summary table |
