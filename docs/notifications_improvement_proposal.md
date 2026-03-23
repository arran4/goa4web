# Notification System Complete Overhaul Proposal

The current notification system tightly couples HTTP tasks to notification logic via interfaces (e.g., `SubscribersNotificationTemplateProvider`). This limits flexibility, makes code monolithic, and causes issues with runtime pattern subscriptions.

To create a more robust, scalable, and decoupled notification system, we propose a complete overhaul migrating away from interface-based triggers towards explicit asynchronous event triggers.

## 1. Explicit Async Event Triggers
Instead of embedding notification template methods directly within HTTP Task structs, tasks will simply publish strongly-typed events to an Event Bus or a Database Queue.

**Current:**
Tasks implement multiple interfaces (`AdminEmailTemplateProvider`, etc.) to return filenames based on context.

**Proposed:**
In the task handler, after successful database commit:
`cd.PublishEvent(eventbus.ThreadCreatedEvent{ThreadID: 1, ActorID: 123})`

A dedicated background worker or queue consumer will handle formatting and dispatching, separating the HTTP lifecycle from the notification delivery lifecycle.

## 2. Event Bus vs. Database Queue
*   **Database Queue (Recommended):** By persisting events to a new `notification_queue` table during the same transaction as the action (e.g., creating a post), we achieve high durability and guarantee notifications aren't lost if the server restarts. A background cron worker can then pull from this queue, evaluate subscriptions, and render templates.
*   **Event Bus:** Alternatively, we continue using the current in-memory event bus but change the payload to specific entity events rather than general `TaskEvent` structs.

## 3. Role-Based Template Defaults and Priorities
To allow flexible templates while preventing "missing template" crashes, we will implement role-based default templates.

Instead of hardcoded filenames, templates will be resolved via a hierarchy in the DB or a configuration map, checking the actor's or recipient's role.

**Resolution Priority:**
1.  **User Preference Override:** If a user specifies a specific template style (future feature).
2.  **Role-Specific Template:** e.g., `core/templates/notifications/admin_thread_created.gotxt` if the recipient is an Admin.
3.  **Action Default:** e.g., `core/templates/notifications/thread_created.gotxt`.
4.  **System Fallback (Optional/Configurable):** Generic notification if specific ones are absent (can be disabled based on preference).

## 4. Entity-Based Subscriptions
Replace URL pattern matching (e.g., `notify:/forum/thread/1`) with explicit `entity_type` and `entity_id` in the database.
*   **Why:** URL routing changes won't break subscriptions.
*   **How:** `INSERT INTO subscriptions (user_id, entity_type, entity_id) VALUES (123, 'forum_thread', 1)`.

## Next Steps
This overhaul requires replacing the current `internal/notifications` package logic and updating all handlers. The proposed system maximizes decoupling, reliability (via DB queues), and flexibility (via priority-based templates).
