# ADR proposal: durable work without persisting the whole event bus

**Status:** Draft for review. **Issue:** [#3197](https://github.com/arran4/goa4web/issues/3197). **Companion:** [#3196](https://github.com/arran4/goa4web/issues/3196) (cooperative cancellation). **Date:** 2026-11-04.

This is a proposed architecture, **not the current implementation**. It records the user's decisions separately from engineering recommendations. No schema migration, new CLI command, persistent queue, or changed worker behavior is introduced by this ADR.

## 1. Decision summary

Use a **database-authoritative, compact, typed work queue** for activities that must survive a restart. Preserve the existing event bus as in-process *fan-out* for notifications and optional wake-up signals. Do **not** serialize every TaskEvent or change the semantics of eventbus.Envelope.Ack. Existing durable systems (pending email, remote image cache and DLQ) should be integrated, not replaced in bulk.

A worker may run inside the web server or as a separate **same-host** command with access to the same database, application configuration and existing cache/blob storage. Polling is the fallback that guarantees progress; no cross-process event bus is required.

Prefer *restarting* an interrupted unit of work, not checkpointing a partial download or index operation. Keep work payloads small and versioned, and reference existing records rather than copying source HTML, media, request state, or Go task implementations.

**First operational slice:** external-link metadata enqueue/claim/fetch/ack/retry, embedded and standalone worker, and command-line inspection/control. Image-cache reconciliation, administrator UI, alerts, and search/cache invalidation follow separately.

### Decisions explicitly confirmed

- Same-host separate worker initially; multi-host deployment is deferred.
- Coalesce duplicate active work; schedule a fresh pass only if the resource is stale/changed or a deliberate forced refresh warrants it.
- Auto-allow retriable exhausted fetches after a cooldown (24 hours is the starting proposal).
- An admin-enqueued unused URL creates/reuses an external_links row.
- First Ctrl+C starts graceful stop; **second Ctrl+C terminates promptly** and unfinished claims recover after expiry.
- Aggregate unread admin failures by **work kind + target + error class**.
- DB job + metadata worker + CLI comes before image and admin-UI work.
- Per-type retry policy; for metadata/images use **one initial attempt plus up to three retries**, proposed intervals **10/20/30 seconds**. Do not accidentally stack this onto the old 2/4/6-second in-memory retry loop.
- Keep terminal history per kind for up to 24 hours or at most 1,000 rows (configurable; first limit reached), never prune active/retry-pending work.

## 2. Current code constraints (verified)

- internal/eventbus/eventbus.go: Message is a typed *publish/subscribe* interface. TaskEvent includes a tasks.Task object and map[string]any; publishing fans out to subscribers. Envelope.Ack only adjusts in-memory delivery accounting; it is not a durable job ack.
- workers/externallinkworker/worker.go: consumes TaskEvent, parses the post body for URLs, resolves external_links, fetches metadata, optionally downloads the image, updates database. The worker's outer context is checked between events, not throughout its retry.
- internal/externallink/fetch.go (merged #3184): retry helper takes string URL, fetch closure and pause closure; no context parameter; 2s/4s/6s pauses under the existing three-attempt policy.
- internal/opengraph/fetch.go: defaults to an SSRF-aware HTTP client but creates its request without a context; DNS safety preflight and redirects must remain secure during a context refactor.
- workers/backgroundtaskworker/worker.go: generic background execution defaults to concurrency 8 and buffer 100 and calls context.WithoutCancel(ctx). Do not assume all workers are serial or that Ctrl+C already interrupts work.
- workers/searchworker/worker.go: uses context.WithoutCancel(ctx) and indexes from event payloads, so a dropped event can leave derived search state stale.
- core/common/download_image.go: already has durable pending remote-image cache records with retry_count/next_attempt_at and scheduled recovery, *plus* an immediate goroutine using a background context. Avoid adding a parallel conflicting image queue.
- workers/emailqueue/email_queue.go: already processes durable pending_emails, with acknowledgement semantics that are separate from eventbus.
- cmd/goa4web/serve.go: already uses signal.NotifyContext for SIGINT/SIGTERM. Its first signal cancels the server context; a deliberate second-signal escalation and worker propagation are still needed. cmd/goa4web/main.go's root command otherwise initializes context.Background().
- cmd/goa4web/email_queue_cmd.go and subcommands provide a local CLI structure that the new jobs commands can follow. Adoption of go-subcommand is separate issue #3194.
- SQL and migrations: source queries belong in internal/db/*.sql, SQLite equivalents under internal/dbsqlite_queries, and generated code is regenerated rather than edited manually. Never reuse already merged migration numbers. Match both SQLite and MariaDB/MySQL.

## 3. Why this option

| Option | Startup/recovery | Operational complexity | Failure risks | Result |
| --- | --- | --- | --- | --- |
| **A. Compact DB work rows + optional wake-up** | Read due work on startup; atomic claims; retries persist | One schema and tiny polling loop | Transaction/lease correctness, DB contention | **Chosen** |
| B. DB metadata with per-job spool files | Possible, with reconciliation | DB + filesystem cleanup and backup coordination | Orphan files and records, shared-path requirements | Only if payload size justifies it |
| C. Persist/replay all eventbus messages | Requires schema for each event/consumer | High: fan-out offsets, schema upgrades, replay policies | Duplicate notifications, nonserializable Task objects, accidental re-execution | Reject initially |
| D. Dedicated Redis/NATS/other queue | Good after configuration | External service and deployment/runtime dependencies | Additional outages, authentication, monitoring | Defer |

Payload size is expected to be very small: metadata jobs should usually carry an external-link **ID**, not the URL body, downloaded image, whole post, or credential. Keep a hard application-level payload size limit (e.g. 4 KiB), and store binary images through the existing upload/cache provider. A future job requiring a large payload must explicitly justify external blob storage.

## 4. Ownership and identities

Distinguish:
1. **Event identity:** identifies an application occurrence, if one exists. It may have multiple consumers and should not be assumed persistent.
2. **Work ID:** immutable opaque identifier for one durable queue record; useful in logs, CLI, and admin UI.
3. **Active deduplication identity:** defined by *kind* and canonical target (plus relevant requested version/scope), for example external-link metadata and external_links.id. Multiple request origins can refer to the same active work.
4. **Resource revision:** controls whether a resource changed while an existing job ran; it is not the same as work ID.
5. **Lease token:** unique to each claim attempt, used to reject an expired worker's late commit.

Do not serialize eventbus.TaskEvent itself. A small typed work-intent descriptor, passed explicitly to a durable enqueue operation, can be associated with an event ID for tracing but must be committed *before* any optional wake-up signal.

The first supported kind should be **external-link.metadata**. Keep type registration small and explicit (e.g. a switch/registry of named handlers with versioned decoders); do not add a generic workflow-definition language.

## 5. Proposed storage model

Use a migration-defined **jobs** table shared by the two supported SQL dialects. The following is a *logical* schema; exact SQL syntax, types, and locking operations are to be verified in the implementation PR:

~~~text
jobs
  id                    opaque immutable ID, primary key
  kind                  short stable name (external-link.metadata)
  target_kind           optional resource classification
  target_id             stable target reference; external_links.id for first slice
  active_dedup_key      nullable fixed-length SHA-256 of kind + canonical target
  payload_version       integer
  payload_json          bounded small JSON; preferably just validated IDs/flags
  status                queued | running | retry_wait | completed | failed | cancelled
  requested_revision    integer (coalescing / forced refresh)
  running_revision      integer (revision snapshotted at claim)
  attempts              integer, incremented when actual work starts
  next_run_at           UTC timestamp, indexed
  lease_token           nullable opaque fencing token
  lease_owner           nullable process-identifying diagnostic
  lease_until           nullable UTC timestamp
  last_error_class      nullable short typed identifier
  last_error_summary    nullable bounded/sanitized text
  created_at, updated_at, finished_at
  requested_by          nullable actor reference; do not use as authorization
~~~

**Critical unique constraint:** unique(kind, active_dedup_key), with active_dedup_key set only on a live/claimable job, and NULL for terminal history. Both supported database engines permit multiple NULL values in a unique index. Use an unambiguous canonical type+target byte sequence before hashing. Distinguish synthetic force-refresh requests from ordinary duplicate access via requested_revision, not by creating unlimited simultaneous rows.

**Indexes:** one for due candidates (kind, status, next_run_at), one for expiry recovery (status, lease_until), one for retention (kind, finished_at), and unique active deduplication. Reevaluate index count with real query plans and low-volume workload; avoid speculative indexes.

**Cooldown state:** terminal history is pruned, so *do not* rely on the existence of a failed job row as the only 24-hour throttling source. Add a minimal authoritative external-link fetch eligibility timestamp / classification to external_links, or equivalent durable per-resource state, during the first metadata migration. This keeps cooldown intact if history is purged at 1,000 records. Hard-invalid URLs and prohibited SSRF targets remain nonretryable even after cooldown unless corrected/explicitly revalidated; never weaken URL safety to meet retry requirements.

Do **not** add a second generic pending image job record while core/common/download_image.go already owns image retry state. Its eventual integration must elect one authoritative work record per image.

## 6. State machine and exactly what gets acknowledged

~~~text
            durable enqueue
                   |
                   v
                queued ------+       re-access after cooldown
                   |         |          or admin retry
             atomic claim    |               |
                   v         |               |
                running      |               |
               /   |   \     |               |
         success transient terminal         |
            |      |        |                |
            v      v        v                |
        completed retry_wait failed ---------+
                       |
                  next_run_at
                       |
                     queued

  first Ctrl+C: stop claims, cancel running task -> queued/retry_wait
  crash or 2nd Ctrl+C: running lease expires -> eligible to be claimed again
  admin cancel: queued/retry_wait -> cancelled (terminal; no auto-resume)
~~~

**One queued job may be claimed by exactly one eligible worker at a time.** Use short database transactions with compare-and-swap/locking, not SELECT followed by an unguarded UPDATE. MySQL/MariaDB and SQLite may need different claim SQL; the interface and state transitions remain identical. Prefer database UTC time for lease comparisons. Initially one worker per *kind* is sufficient; a same-host embedded and separate process must still be protected against accidental competing claims.

Each attempt is claimed with a fresh, random lease token. While an attempt runs, it may renew its lease if bounded execution requires it. Completion writes must be conditioned on matching current lease token/revision; an expired worker must not publish stale metadata or ack work claimed by a successor. The first slice can keep an HTTP timeout shorter than its bounded claim lease; no generic heartbeat is needed unless real tasks exceed it.

**Strongly preferred for the metadata first slice:** in one short database transaction, verify ownership/fencing, update external_links metadata, mark the job completed (or requeue if a newer revision requires another pass), and commit together. Network fetching happens *outside* the transaction. If work crashes between fetch and commit, retry safely from the beginning. For filesystem/blob work, idempotent object writes and eventual reconciliation replace cross-store transactions.

Do not use eventbus.Envelope.Ack as the job-completion operation. A failed optional bus publish after DB commit must not roll back durable completion. Live UI fragments should re-fetch authoritative resource state.

## 7. Enqueue and deduplication

**Producer ordering:**
1. Validate authorization and normalize the target URL. Apply SSRF rules at execution and on re-execution as well.
2. For an administrator-supplied URL, create/reuse the canonical external_links row using existing ResolveRecord/EnsureExternalLink semantics (do not invent a new orphaned content model).
3. Commit a work intent / claimable job **in the same transaction as the authoritative mutation**, or derive the job by scanning a committed dirty/resource marker. Never rely solely on an in-memory event published after a commit.
4. Optional local wake-up signals the worker to poll. A standalone worker without shared memory polls its database at a modest configurable cadence (proposal: about 2 seconds when active/idle, with bounded queries and adaptive sleep).
5. At process start, enumerate due jobs and expired claims before waiting for future notifications.

Existing post-body events can remain for notification consumers, but the externally visible mutation path should arrange the durable enqueue. If it cannot do so in the same transaction, document the reconciliation scanner that closes the commit→enqueue crash gap before claiming restart-safety.

**Dedup semantics by kind:**
- Metadata: multiple accesses to a current in-flight URL coalesce. A forced refetch or authoritative change advances requested_revision; after a successful run, compare running_revision to requested_revision and queue **at most one** follow-up. A plain view while processing does not automatically schedule another.
- Image: same canonical image/cache key coalesces; an interrupted transfer restarts from byte zero. Existing pending cache record is the source of truth.
- Search index: mark indexed resource dirty or bump generation in the mutation transaction, index from durable resource contents (not an expired event body's text), and only mark clean if processed generation still equals requested generation. New writes during indexing produce one more pass.
- Regenerable caches: invalidate version/cache key and rebuild on demand; avoid creating rows merely to mark derived data dirty.
- Existing email/digest queue: keep its durable semantics, dedup and provider acknowledgement independent of the new queue.

## 8. Retry, terminal errors, and bounded history

**Initial metadata policy proposal:** 1 initial attempt + 3 retries at 10s, 20s, 30s. Each failed attempt commits next_run_at. On startup, use the persisted due time; do not reset backoff. Do not start one goroutine/timer for each job. Add per-kind configuration and upper bounds; jitter and per-host rate limits may be considered later but are not necessary for low volume.

The existing #3184 FetchMetadata performs three in-memory attempts and sleeps 2/4/6 seconds. **When migrating it to a durable attempt scheduler, invoke one network attempt per claim or explicitly disable its internal retry path.** Do not accidentally get 4 × 3 nested fetch attempts or uninterruptible final sleeps. #3196 should separately make any remaining pauses cancellation-aware without conflating cancellation with a retryable failure.

Proposed error classes:
- Transient: connect failures, timeouts, HTTP 429 (respect bounded Retry-After), typical HTTP 5xx, temporary storage failure.
- Normally terminal until corrected: invalid/malformed URL, disallowed destination/SSRF, unsupported scheme, payload/version decode failure, permanent authorization error.
- Context cancellation from graceful shutdown: **not a failed attempt**; requeue or leave safely claimable; avoid consuming retry budget.
- Repeated transient exhaustion: terminal failed history, logs/aggregated admin diagnostics, and per-resource cooldown before automatic reaccess. Manual admin retry can explicitly override the transient cooldown but must never bypass SSRF checks.

Keep **terminal records** up to 24h and at most 1,000 per kind (prune the oldest terminal rows first if the per-kind count is exceeded), independently of queued/running/retry_wait rows. These are operational audit-history limits, not a queue capacity cap. If active work becomes too large, apply bounded intake/backpressure and report it; never silently drop durable jobs. Keep cooldown and unread failure aggregation independent of the prunable history.

## 9. Process model and signals

Two deployment modes, with no distributed infrastructure:

- Embedded (default): goa4web serve starts the due-job poller/worker along with the server.
- External: goa4web jobs worker (optionally --kind=external-link.metadata) runs using the same DB/config/cache storage, without binding an HTTP port. A configuration switch disables embedded workers to avoid unnecessary double processing; database claims must still remain safe if both are accidentally enabled.

A single process per kind is the default; **generic backgroundtaskworker already has a separate concurrency=8 policy**, so this is a policy for the new durable work types, not a claim about current workers. Later, make concurrency configurable per kind if warranted.

**Signal behavior:**
1. First SIGINT/SIGTERM stops accepting work/claiming new jobs, signals cancellation through real task contexts, and gives a bounded cleanup window (proposed 30s, configurable). Use a fresh bounded cleanup context to release a claim; do not use the already-cancelled task context for DB cleanup.
2. For interrupted work, release/requeue without charging a failed attempt when possible. If cleanup fails or the process is killed, lease expiry makes the work eligible again.
3. Second Ctrl+C terminates promptly. cmd/goa4web/serve.go currently calls signal.NotifyContext; implementation must restore default handling or explicitly escalate on a second signal, because a notification context alone does not establish two-stage semantics.
4. SIGKILL/power loss requires no callback or special file write. Recovery is from committed job records and expired leases.

**Shutdown must not wait indefinitely** for context.WithoutCancel tasks. #3196 should audit backgroundtaskworker, searchworker, external-link workers, Open Graph fetches, image cache and app server shutdown; do not change all unrelated semantics without separate tests.

## 10. CLI and administration

Build CLI against the shared job store, using existing cmd/goa4web/email_queue_* patterns and configuration precedence. No dependency on the future go-subcommand adoption (#3194).

Proposed CLI shape (names subject to CLI review):

~~~shell
goa4web jobs list --status=failed --kind=external-link.metadata
goa4web jobs show --id <job-id>
goa4web jobs add external-link --url https://example.org/article
goa4web jobs retry --id <job-id>
goa4web jobs cancel --id <job-id>
goa4web jobs purge --kind=external-link.metadata --terminal-only
goa4web jobs worker --kind=external-link.metadata
~~~

The CLI should support structured JSON output for inspection in addition to human-readable output. Mutating operations must check state and be idempotent or return a clear conflict. Cancel of an active running job is *not* required for the initial slice; pending/retry-wait cancellation and terminal retry are. Do not promise a suspended state that does not yet exist.

The later admin web UI uses the same service, not direct ad-hoc SQL. Ordinary users should see passive placeholders/diagnostics and optional re-fetch controls, never a global queue dashboard.

**Unread admin failure aggregation** belongs to a later PR: group by (kind, target, error class), count occurrences, track last occurrence, and update an existing unread admin notification instead of creating duplicates. On acknowledgement/read, subsequent failures may start a fresh grouping period. Any notification or log rendering must enforce authorization; private source URLs should not leak to unrelated administrators. A summary row should survive terminal-history cleanup long enough to preserve unread counts. Integrate with activity-key notification dedup if its own PR lands; don't invent a competing notification store prematurely.

## 11. Failure-mode analysis

| Failure | Required observable result |
| --- | --- |
| DB commit succeeds; wake-up is lost | Poller/startup scan sees job; eventual processing |
| Source mutation commits; process dies before enqueue | Same-transaction work row or durable dirty scan recovers; no silent gap |
| Duplicate enqueue during running | One active row; optional revision bump; no second concurrent fetch |
| First Ctrl+C during request | Task context cancels; DB work retained; no ordinary failure count |
| SIGKILL / second Ctrl+C | Expired lease reclaimed; full job restarts |
| Worker A lease expires and worker B claims | A's stale token cannot commit metadata or ack B's job |
| Metadata DB update succeeds but ack fails | Prefer a *single transaction* for result + ack; otherwise idempotent replay |
| Persistent failure | Counted attempts, terminal result, 24h reaccess cooldown, admin diagnostic |
| History exceeds 1,000 | Oldest terminal history purged by kind, active work/cooldown preserved |
| DB unavailable at startup | Worker fails visibly/retries startup safely; do not claim jobs without DB |
| Image file written but DB state fails | Deterministic object ID and reconciliation; no lost authoritative pending work |
| Web process and external worker both started | Transactional claim prevents simultaneous job ownership |
| A job type's payload version is unknown | Do not execute; record bounded, inspectable error without dropping payload |
| Malicious/internal URL is enqueued | Validate at enqueue and every fetch; refuse SSRF even on admin retry |

## 12. Tests required before describing implementation as reliable

**Driver parity:** SQLite and MariaDB/MySQL upgrade tests (latest migration number checked against main at implementation time), schema seed test, uniqueness/active-key NULL handling, claims, collisions, generated sqlc adapters, rollback/Down where safe.

**Crash/restart:** enqueue + stop before wake-up, kill after claim, kill during HTTP read, restart during retry_wait, claim expiry and fencing, DB commit then lost wake-up, duplicate enqueue and force-refresh while running, process killed after successful external-link update, restart and automatic completion.

**Concurrency:** accidentally run embedded and external worker at once; verify only one fenced claim; no growing goroutine/timer count with thousands of delayed jobs; per-kind serialization and bounded query batches.

**Cancellation:** first Ctrl+C interrupt during HTTP and during retry; second Ctrl+C forced exit; stale lease reclaim; context cancellation not treated as ordinary failure; existing generic worker shutdown regression.

**Behavior/security:** URL canonicalization, SSRF DNS/redirect protections, admin enqueue for unused URL, cooldown through history cleanup, terminal error classification, DB outage, stale metadata, nonroot configured base URL where relevant, user/private data not logged or exposed.

**Derived work follow-up:** image cache pending-row integration, search revision change during indexing, cache invalidation, email queue unaffected. Admin failure grouping and UI are later issue-owned tests.

Prefer deterministic stubbed clocks, synthetic DB failures, and explicit process-kill integration tests for the OS behavior. Do not substitute a mock-only successful unit test for restart safety.

## 13. Implementation sequencing

1. **#3196 (separate PR):** stop leaking cancellation; make Open Graph requests and retry waits cancellable; ensure second Ctrl+C and bounded shutdown semantics. Avoid changing existing durable business behavior.
2. **#3197 (this Draft ADR PR):** review this model, verify schema/claim dialect details and record final acceptance or requested changes. No migration yet.
3. **New implementation issue/PR after ADR approval:** add typed external-link work enqueue + compact DB schema + claims/retries + start recovery + embedded/standalone worker + CLI. Migrate the current external-link worker carefully without deleting other TaskEvent consumers; preserve #3181 lifecycle service integration. Add operational tests above. Do not claim success before lost-event/transaction gap is closed.
4. **Image follow-up:** consolidate image-cache pending work with the chosen durable design, one authoritative queue; preserve existing placeholder and cache storage semantics.
5. **Admin follow-up:** web UI, permission-aware controls, debounced unread failure group counts and terminal-history cleanup.
6. **Search/cache follow-up:** resource dirty/revision markers and invalidation, avoiding redundant per-update work rows.
7. **Optional later:** per-kind concurrency configuration, multi-host distribution, specialized blob payload storage, and remote worker authentication. No blanket system rewrite.

Do not merge independently conflicting migrations; use the next unallocated migration number at implementation time (related notification PRs currently occupy adjacent migration work). Include code-generated DB output and test both supported dialects.

## 14. Items for technical validation during ADR review

These are **not unanswered user product questions**; the implementation agent should investigate and report them:

- Verify MySQL/MariaDB and SQLite atomic claiming implementation and whether a lease heartbeat is necessary for actual metadata timeouts. Consider portability of SELECT FOR UPDATE SKIP LOCKED versus conditional UPDATE/RETURNING.
- Identify the exact HTTP mutation transaction points that must persist work intent; avoid a crash window that only a bus event could have closed.
- Confirm which existing external-link metadata trigger paths currently cause duplicate fetches and how to disable the old worker safely.
- Verify the external_links persistent cooldown fields and their consistency with manual forced refresh and authorization.
- Confirm how the existing pending image-cache query marks retries and whether it can later become a handler for the shared claim system without two queues.
- Validate graceful shutdown ordering between server, event bus, background worker, new poller and DB Close.
- Assess whether one small job package (e.g. internal/jobqueue) suffices; keep domain-specific metadata handling in internal/externallink, not in a generic application service locator.

**Review outcome sought:** approve the small architectural boundary and the first implementation slice, not detailed speculative APIs or a replacement event bus.
