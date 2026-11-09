# Debugging Findings

## Bug 1 — Deduplication only worked inside one batch
Cause: The seen map was recreated for every batch of 200 events.
Fix: Use one event ID map across the entire input file.
Verification: Run against the supplied expected output and compare.

## Bug 2 — Unique opens were shared across campaigns
Cause: The map used only contact_id as its key.
Fix: Use campaign_id plus contact_id as a composite key.
Verification: A contact opening in two campaigns is counted once in each.

## Bug 3 — Daily buckets used local timezone
Cause: Timestamp.Local() depends on the machine timezone.
Fix: Use Timestamp.UTC() before formatting the date.
Verification: Compare daily buckets against the expected output.

## Bug 4 — Concurrent counter increments raced
Cause: Multiple workers incremented ordinary integer fields simultaneously.
Fix: Protect counter updates with a mutex.
Verification: Run with the Go race detector.
