# Notification System Complete Overhaul Proposal

The current notification system tightly couples HTTP tasks to notification logic via interfaces (e.g., `SubscribersNotificationTemplateProvider`). This limits flexibility, makes code monolithic, and causes issues with runtime pattern subscriptions.

To create a more robust, scalable, and decoupled notification system, we propose a complete overhaul migrating away from interface-based triggers towards explicit asynchronous event triggers and a unified, file-based configuration system.

## 1. Unified Configuration via `txtar` Archives
**Current State:**
Notification logic is hardcoded across Go interfaces, and templates are split across multiple `.gotxt` and `.gohtml` files.

**Proposed State:**
Consolidate the entire configuration for a specific notification into a single `txtar` file. This file will be loaded into memory and contain:
*   **Event Listening Logic:** What specific event (e.g., `ThreadCreatedEvent`) triggers this notification.
*   **Filters & Rules:** Logic to determine who receives the notification (e.g., permissions, specific user states).
*   **Delivery Components:** The actual templates for different channels (`email.gohtml`, `email.gotxt`, `internal.gotxt`, `admin.gotxt`).

By consolidating into single files, administrators can easily "side-load" custom directories of `txtar` files rather than modifying massive preconfigured data structures in code.

## 2. Role-Based Toggles and Overrides
This unified `txtar` approach inherently supports robust toggles:
*   **User/Admin Toggles:** Entire notifications (defined by a single file) can be toggled on or off globally or per-user.
*   **Role-Based Defaults:** The default configuration is a template that is applied by role (which now acts as a notification toggle itself). Users inherit default toggles based on their role, but the underlying system remains file-driven.
*   **Priority:** If conflicting `txtar` configurations exist for the same event, a priority metadata field inside the archive resolves the conflict.

## 3. Explicit Async Event Triggers
Instead of embedding notification template methods directly within HTTP Task structs, tasks will simply publish strongly-typed events to an Event Bus or a Database Queue.

**Proposed Flow:**
In the task handler, after successful database commit:
`cd.PublishEvent(eventbus.ThreadCreatedEvent{ThreadID: 1, ActorID: 123})`

The system then evaluates all loaded `txtar` configurations against the `ThreadCreatedEvent`, executes their filters, and renders their component templates.

## 4. Event Bus vs. Database Queue
*   **Database Queue (Recommended):** By persisting events to a new `notification_queue` table during the same transaction as the action, we achieve high durability. A background worker processes this queue against the `txtar` rules.
*   **Event Bus:** Alternatively, we continue using the current in-memory event bus but change the payload to specific entity events.

## 5. Entity-Based Subscriptions
Replace URL pattern matching (e.g., `notify:/forum/thread/1`) with explicit `entity_type` and `entity_id` in the database.
*   **Why:** URL routing changes won't break subscriptions.
*   **How:** `INSERT INTO subscriptions (user_id, entity_type, entity_id) VALUES (123, 'forum_thread', 1)`.

## Next Steps
This overhaul requires replacing the current `internal/notifications` package logic, rebuilding the admin side of notifications to support file-based toggling, and updating all handlers. The proposed system maximizes decoupling, reliability (via DB queues), and ease of customization (via `txtar` side-loading).
