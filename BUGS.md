# Debugging Assignment — BUGS.md

## Bug 1: Duplicate Event IDs Across Batches

**Problem:** The `seen` map was created inside `processBatch()`. Consequently, duplicate event IDs appearing in different batches were counted multiple times.

**Fix:** Introduced a persistent `processedIDs` map that tracks event IDs across the entire input file.

**Result:** Each event ID is processed at most once.

## Bug 2: Unique Opens Were Not Campaign-Specific

**Problem:** The `openedBy` map used only `ContactID` as its key. A contact opening multiple campaigns was incorrectly counted as a unique open only once globally.

**Fix:** Changed the key to `CampaignID + ":" + ContactID`.

**Result:** Unique opens are tracked independently for each campaign.

## Bug 3: Daily Delivered Buckets Used Local Time

**Problem:** `Timestamp.Local()` depended on the machine's timezone, potentially assigning events to incorrect calendar dates.

**Fix:** Replaced `Timestamp.Local()` with `Timestamp.UTC()`.

**Result:** Daily delivered buckets consistently use UTC dates.

## Bug 4: Concurrent Counter Updates Caused a Data Race

**Problem:** Multiple workers concurrently incremented shared campaign counters without synchronization.

**Fix:** Added `statsMu` to protect counter updates in `apply()`.

**Result:** Counter updates are synchronized between workers.

## Verification

- Program output matches `expected_output.txt`.
- Repeated executions produced identical output.
- `go test ./...` completed successfully.
- Race detector could not be executed because CGO is not enabled in the current Windows Go environment.
