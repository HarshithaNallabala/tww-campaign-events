# Campaign Events Service — Engineering Notes

## Part 1: Problem Interpretation and Decisions

### 1. Interpretation

Relay receives webhook events from external delivery providers for email, SMS, and push campaigns. Providers may retry events or send them late and out of order.

The service must:
- Accept provider events through `POST /events`.
- Avoid counting the same event more than once.
- Aggregate sent, delivered, opened, and clicked events per campaign.
- Provide campaign statistics through `GET /campaigns/{campaign_id}/stats`.
- Remain correct when concurrent requests arrive.

### 2. Assumptions

- `event_id` uniquely identifies an event across providers and campaigns.
- Two events with the same `event_id` are considered duplicates.
- Event types are case-sensitive and restricted to `sent`, `delivered`, `opened`, and `clicked`.
- Events require a non-empty event ID, campaign ID, contact ID, valid event type, and RFC3339 timestamp.
- Timestamps represent event occurrence time, not ingestion time.
- The current service counts accepted events by type. `opened` is a total event count, while unique opens are not part of the service's current stats response.
- Duplicate events are ignored and reported separately.
- Invalid events are rejected individually; valid events in the same batch are still accepted.
- The initial implementation uses in-memory storage because the assignment permits it and the expected traffic is small.
- Statistics are calculated from accepted events stored in memory.

### 3. Ambiguities

- Is `event_id` globally unique, or unique only within a provider?
- Should invalid events be rejected individually or should the entire batch fail?
- Should duplicate events return success, conflict, or a separate response?
- Does "opened" mean total open events or unique contacts who opened?
- Should unique opens be campaign-specific?
- How should events from unknown or unsupported event types be handled?
- Should events with missing contact IDs be accepted for event types where contact identity is unavailable?
- Should timestamps far in the future or very old timestamps be accepted?
- Should campaign statistics include events received late?
- Is the dashboard expected to show eventual consistency or strictly up-to-date counts?

### 4. Questions for the PM

1. Is `event_id` globally unique across all providers?
2. What should the API response and HTTP status be for mixed valid/invalid batches?
3. Should duplicate events be treated as successful idempotent requests?
4. What exact definition should we use for opened and unique opens?
5. Are campaign statistics based on event occurrence time or ingestion time?
6. Do we need to preserve raw provider payloads for auditing?
7. What is the expected retention period for events?
8. Should providers authenticate or sign webhook requests?
9. What latency and availability targets should the dashboard API meet?
10. Should events with unknown types be stored for future compatibility?

### 5. Priorities

1. Correct event validation and ingestion.
2. Idempotency through event ID deduplication.
3. Thread-safe storage and aggregation.
4. Campaign stats API.
5. Tests for valid events, duplicates, invalid events, and statistics.
6. Debugging the supplied program and verifying expected output.
7. Scale and operational considerations.
8. Optional features only after required work is complete.

---

## Part 4: Scaling to 100 Million Events per Day

100 million events/day is approximately 1,157 events/second on average, with potentially much higher peak traffic.

### What breaks first?

- The in-memory event map grows without limit and consumes increasing RAM.
- Restarting the process loses all events and counters.
- A single mutex serializes ingestion and stats access.
- Calculating statistics by scanning every event becomes increasingly expensive.
- A single instance cannot provide availability or horizontal scaling.
- There is no durable recovery or replay mechanism.

### What I would change, in order

1. Measure peak traffic, request latency, memory usage, and aggregation costs.
2. Move event storage to durable storage such as PostgreSQL, partitioned by time or campaign.
3. Add a unique constraint on `event_id` (or provider + event ID if uniqueness is provider-scoped).
4. Maintain campaign aggregates incrementally rather than scanning all events.
5. Add indexes for campaign and event-time queries.
6. Introduce batching and connection pooling.
7. Partition or shard data when a single database becomes insufficient.
8. Add horizontal API scaling and load balancing.
9. Add observability, alerting, and recovery procedures.

### Would the API contract change?

Not necessarily. The same endpoints and response shapes can remain stable while the storage and processing architecture changes behind them.

I would preserve the current API contract unless product requirements demand pagination, asynchronous ingestion responses, or richer reporting.

### Would storage change?

Yes. In-memory storage would be replaced with durable database storage. Event records would have a unique key and indexes. Aggregates could be stored separately for fast stats queries.

### Would I introduce a queue?

Yes, if ingestion and aggregation need to be decoupled.

A queue such as Kafka could absorb bursts and allow independent consumers to aggregate events.

New problems introduced:
- Event processing becomes eventually consistent.
- Consumers may process messages more than once.
- Queue lag can make dashboard counts stale.
- Ordering across partitions is not guaranteed.
- Failed messages need retries and dead-letter handling.
- Replay can accidentally double-count aggregates without idempotent processing.

### Where can duplicates sneak in?

- Provider retries.
- API retries after timeouts.
- Queue redelivery.
- Consumer crashes after updating aggregates but before acknowledging a message.
- Replay or backfill jobs.
- Concurrent processing by multiple consumers.

Deduplication must therefore be durable and atomic, not only an in-memory check.

### How would I keep counters trustworthy?

- Use durable unique event IDs.
- Make event insertion and aggregate updates transactional where possible.
- Use idempotent consumers.
- Reconcile aggregates periodically against raw events.
- Support replay and correction workflows.
- Track rejected, duplicate, delayed, and malformed events separately.

### What would I monitor?

- Ingestion throughput and latency.
- HTTP error rates.
- Duplicate and invalid event rates.
- Queue depth and consumer lag.
- Database latency, connection usage, and storage growth.
- Memory and CPU usage.
- Aggregation lag.
- Campaign stats query latency.
- Reconciliation discrepancies.

### What would I deliberately not solve yet?

I would not immediately introduce a complex distributed architecture, multiple databases, or elaborate sharding without measurements. I would start with durable storage, correct idempotency, incremental aggregation, and observability, then scale based on actual bottlenecks.

---

## Part 5: The Angry Marketer

### Observed dashboard

At 10:00:
- Sent: 1,000,000
- Delivered: 970,000
- Opened: 250,000
- Clicked: 20,000

At 10:30:
- Sent: 1,000,000
- Delivered: 975,000
- Opened: 248,000
- Clicked: 20,000

Delivered increased by 5,000, while opened decreased by 2,000.

### Plausible explanations

- Late delivery events arrived after the first snapshot.
- Late open events were not yet received or were processed differently.
- The dashboard uses different time windows or filters for delivered and opened.
- The opened metric represents unique contacts, and deduplication or corrections changed the count.
- Events were reclassified or invalidated after provider reconciliation.
- A data correction or backfill changed historical aggregates.
- The dashboard reads from different replicas or snapshots with different freshness.
- A cache was refreshed inconsistently.
- Events were deleted, corrected, or excluded due to invalid data.
- A bug in aggregation, deduplication, or dashboard rendering caused the decrease.
- Open tracking is affected by privacy features, image blocking, or email-client behavior.
- Provider reporting semantics differ from the dashboard's metric definition.

### What I would check first

1. Confirm both dashboard snapshots use the same campaign, filters, time window, and metric definitions.
2. Check whether opened is a total event count or unique-contact count.
3. Inspect ingestion timestamps and event timestamps for late arrivals.
4. Compare raw events and aggregate records for the affected campaign.
5. Check duplicate, rejected, correction, and backfill logs.
6. Check cache and database replica freshness.
7. Recalculate the metric independently from raw events and compare it with the dashboard.
8. If the discrepancy remains, investigate the aggregation code and dashboard query.

### How to decide bug vs expected behavior

If the metric definition allows corrections, deduplication changes, or unique-contact calculations, a decrease can be legitimate.

If the same metric definition and source data produce a different result without an explained correction, it is likely a bug.

I would reproduce the two snapshots from raw events before concluding.

### Is delivered increasing suspicious?

No. Delivered increasing is expected when delivery events arrive asynchronously or late. It is consistent with the provider retry and out-of-order behavior described in the brief.

The suspicious part is opened decreasing, which requires checking metric semantics, corrections, data freshness, and aggregation consistency.

---

## What I completed / what I intentionally skipped

### Completed
- Implemented `POST /events`.
- Implemented `GET /campaigns/{campaign_id}/stats`.
- Added event validation and RFC3339 timestamp validation.
- Added per-event decoding for mixed valid/invalid batches.
- Added duplicate event ID handling.
- Protected shared in-memory storage with a mutex.
- Added backend tests for ingestion, campaign stats, duplicates, and invalid event types.
- Fixed all four bugs in the debugging program.
- Verified debugging output against `expected_output.txt`.
- Verified deterministic output across repeated runs.
- Created `BUGS.md`.

### Intentionally skipped
- Optional campaign events listing endpoint.
- Frontend/dashboard.
- Persistent database storage.
- Authentication and provider signature verification.
- Deployment and Docker.
- Race detector execution because CGO is unavailable in the current environment.

## If I had another day

- Add integration tests using the complete supplied seed dataset.
- Improve request validation, error responses, and HTTP request limits.
- Add tests for malformed individual JSON elements and concurrent ingestion.
- Introduce SQLite persistence and verify recovery after restart.
- Add structured logging and metrics.
- Add pagination for campaign events.
- Improve API documentation and run instructions.
- Revisit the design with measured traffic and performance benchmarks.