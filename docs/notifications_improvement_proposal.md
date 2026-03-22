# Notification System Improvement Proposal

The current notification system in Goa4Web is robust but suffers from several design choices that make it fragile to code changes (indirect routing) and computationally expensive at runtime. This proposal outlines a path to simplify subscriptions, improve performance, and decouple notification logic from HTTP tasks.

## 1. Shift to Entity-Based Subscriptions
**Current State:**
Subscriptions are tracked via URL pattern matching strings (e.g., `notify:/forum/thread/1`).
* **Problem:** If the routing structure or URL scheme changes, the `collectSubscribers` matching logic breaks, stranding old subscriptions.

**Proposed State:**
Change the `subscriptions` database table to use explicit `entity_type` and `entity_id` columns.
* Instead of inserting `pattern: notify:/forum/thread/1`, insert `entity_type: forum_thread, entity_id: 1`.
* This creates a hard, database-level association that is immune to HTTP routing changes and allows for straightforward referential integrity.

## 2. Decouple Task Definitions from Notification Execution
**Current State:**
HTTP tasks (e.g., `ForgotPasswordTask`) directly implement notification interfaces (`SubscribersNotificationTemplateProvider`, etc.) and dictate which templates are used. The event bus relies on the Task interface to execute logic during the worker thread.
* **Problem:** Tasks become monolithic, mixing HTTP validation, database execution, and notification formatting.

**Proposed State:**
Introduce explicit `NotificationEvent` structures that are published to the event bus containing only metadata (actor ID, entity ID, action).
* Move template resolution into the Notification worker based on the event type, rather than calling methods on the original HTTP Task.
* Example: `eventbus.Publish(ThreadCreatedEvent{ThreadID: 1, UserID: 1})` -> The notifier resolves the subscribers for `entity_type: forum_thread, entity_id: 1` and uses the standard thread creation template.

## 3. Simplify Permission Evaluation (SystemCheckGrant) in Workers
**Current State:**
`notifySubscribers` dynamically evaluates `SystemCheckGrant` in a loop for *every* individual subscriber to ensure they still have access to the object before sending the notification.
* **Problem:** This leads to N+1 query problems and significantly degrades batch performance when notifying hundreds or thousands of users.

**Proposed State:**
* **Option A (Eager Eviction):** When a user's role or grant is revoked, explicitly delete their associated subscriptions in the database.
* **Option B (Optimized Bulk Query):** If runtime evaluation is strictly required, update `SystemCheckGrant` to allow bulk ID checking (e.g., `SELECT idusers FROM ... WHERE id IN (...) AND <grant_logic>`), fetching all authorized users in a single query rather than iterating.

## 4. Default Templates (Implemented)
As a short-term resilience fix, default fallback templates (`default.gotxt`, `defaultEmail.txtar`) have been added. If a specific component forgets to register a template or it is deleted, the system falls back to a generic message rather than crashing the worker or throwing errors.
