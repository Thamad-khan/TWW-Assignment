# Engineering Notes

## Part 1 — Understanding the problem

The service receives webhook events from delivery providers and calculates campaign-level totals. Providers may retry events, and events may arrive late or out of order. The service must avoid counting the same provider event twice.

## Assumptions
- event_id is globally unique across providers.
- The first valid event with an event_id wins.
- A duplicate is ignored rather than treated as a new event.
- opened counts events; unique_opens counts distinct contacts per campaign.
- All timestamps must be RFC3339.
- Invalid items should not invalidate valid items in the same batch.
- The initial implementation uses in-memory storage.

## Ambiguities and questions for the PM
- Is event_id globally unique or unique only within a provider?
- How long must deduplication records be retained?
- Should conflicting payloads with the same event_id be alerted?
- Should opened count every open or only the first open per message?
- Are campaign statistics expected to be eventually consistent?
- Is there a retention or deletion policy?
- What authentication mechanism do providers use?
- Should batches be atomic or partially accepted?

## Priorities
1. Correct idempotency.
2. Correct event validation.
3. Thread-safe statistics.
4. API error handling.
5. Tests and documentation.
6. Optional pagination and durable storage later.

## Part 4 — Scale memo: 100 million events/day

At this volume, in-memory storage and a single process would fail first. Restarting loses data, memory grows without bound, and one mutex can become a contention point.

I would first measure throughput and latency, then move events to durable storage with indexed event IDs and campaign IDs. I would partition data by time or campaign and use database transactions to make deduplication and counter updates atomic.

A queue could absorb bursts and separate ingestion from aggregation. It introduces retries, consumer lag, ordering issues, poison messages and at-least-once delivery. Consumers must remain idempotent.

Duplicates could occur at the provider, HTTP retry, queue redelivery and consumer retry layers. I would keep event IDs unique in durable storage and reconcile aggregates against raw events.

I would monitor ingestion rate, rejected events, duplicate rate, queue lag, processing latency, database latency, error rate and counter reconciliation differences.

I would not introduce complex multi-region consistency or elaborate streaming infrastructure until measurements justify them.

The API contract can remain stable while the storage changes. The in-memory implementation must be replaced before production scale.

## Part 5 — The angry marketer

Possible explanations for opened decreasing:
- A rolling time window or changed reporting window.
- Late-arriving events or backfilled data.
- Duplicate cleanup or deduplication corrections.
- A change in unique-open definitions.
- Data correction or deletion.
- Different filters, timezone boundaries or campaign attribution.
- A cache refresh or stale dashboard snapshot.
- A bug in aggregation or query logic.
- A change in provider data or event classification.

Delivered increasing is not inherently suspicious. Delivery events can arrive late, so the count can rise after the original send.

I would first compare the exact query, time window, timezone, filters and aggregation version used for both snapshots. Then inspect raw event IDs and timestamps, replay the affected events, and reconcile counts against stored records. If the same immutable event set and same query produce different totals, I would investigate a bug.

## What I completed / what I intentionally skipped
Completed: API implementation, in-memory idempotency, validation, campaign statistics and assignment documentation.
Intentionally skipped: durable database storage, authentication, pagination and deployment. These were outside the essential time-boxed scope.

## If I had another day
I would add integration tests, SQLite persistence, pagination, structured logging, metrics and provider authentication.
