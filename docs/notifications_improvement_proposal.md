# Notification System Complete Overhaul Proposal

The current notification system tightly couples HTTP tasks to notification logic via interfaces (e.g., `SubscribersNotificationTemplateProvider`). This limits flexibility, makes code monolithic, and causes issues with runtime pattern subscriptions.

We propose an overhaul migrating towards explicit asynchronous event triggers and a unified, file-based configuration system. However, this transition introduces significant complexities that must be critically evaluated.

## 1. Unified Configuration via `txtar` Archives
**Proposed State:**
Consolidate the entire configuration for a specific notification into a single `txtar` file containing event listening logic, filters/rules (e.g., permissions), and delivery component templates (email/internal/admin).

**Critical Analysis & Risks:**
*   **Loss of Type Safety:** Currently, Go interfaces enforce that notification providers return necessary data. Moving logic and rules into text files removes compile-time checks. A malformed `txtar` file or a typo in a rule string won't be caught until runtime.
*   **Parsing Overhead & Complexity:** Designing a bespoke DSL (Domain Specific Language) inside a `txtar` file to evaluate "filters and rules" (like checking database grants) is a massive undertaking. It risks reinventing a slow, fragile interpreter within the notification worker.
*   **Debugging Nightmare:** Tracing why a notification *didn't* send becomes much harder when the logic is buried in a parsed text file rather than step-through Go code.

## 2. Role-Based Toggles and Overrides
**Proposed State:**
User/Admin toggles defined by the `txtar` file, with role-based default templates acting as the primary notification toggle, resolved via a priority metadata field.

**Critical Analysis & Risks:**
*   **Admin UI Complete Rebuild:** As noted, this requires tearing down the current data-structure-driven Admin UI. Rebuilding a UI to dynamically read, parse, and represent state from a side-loaded directory of text archives is highly complex and prone to synchronization issues (e.g., concurrent map reads/writes if the filesystem changes).
*   **Priority Clashes:** "Priority fields" in decoupled files often lead to hidden bugs. If two separate `txtar` files claim priority `100` for the same event, the system's behavior becomes non-deterministic without strict validation logic.

## 3. Explicit Async Event Triggers
**Proposed State:**
Tasks publish strongly-typed events to an Event Bus or a Database Queue (e.g., `cd.PublishEvent(eventbus.ThreadCreatedEvent{ThreadID: 1, ActorID: 123})`).

**Critical Analysis & Risks:**
*   **Data Stale-ness:** Asynchronous resolution means the notification worker evaluates rules *after* the event occurs. If a user's permissions change between the event firing and the worker processing it, they might receive an invalid notification (or miss a valid one). The payload must encapsulate all necessary state at the time of the event, bloating the event payload.

## 4. Event Bus vs. Database Queue
*   **Database Queue:**
    *   *Pros:* High durability, no lost notifications on server restart.
    *   *Cons:* Introduces write contention on a single `notification_queue` table for every action on the site. Polling the queue introduces latency and constant database load compared to memory channels.
*   **Event Bus:**
    *   *Pros:* Zero DB latency, extremely fast in-memory dispatch.
    *   *Cons:* Non-durable. If the server crashes, all queued notifications are permanently lost.

## 5. Entity-Based Subscriptions
Replace URL pattern matching (e.g., `notify:/forum/thread/1`) with explicit `entity_type` and `entity_id` in the database.
*   *Critique:* This is a universally positive change. The current string-matching approach is highly brittle and couples subscriptions directly to URL routing design.

## Conclusion
While migrating to a side-loaded `txtar` system provides immense flexibility for power users, the development cost to build a safe parser, evaluator, and dynamic Admin UI is extremely high. The team must weigh if the flexibility of "no code modifications for new notifications" justifies the risk of runtime evaluation failures and the loss of Go's strict typing.
