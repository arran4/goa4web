# Notification System Overhaul Proposal

## 1. Overview
The current notification system is fragile and difficult to test due to its reliance on interface-based triggers scattered throughout the codebase. This proposal outlines a comprehensive rewrite to make the system more reliable, understandable, testable, and highly configurable for users based on their roles and subscription tiers.

## 2. Core Architecture
We will move away from implicit interface-based triggers and instead rely on **explicit asynchronous event triggers**.

- **Event Bus Integration**: All actionable system events (e.g., `thread_replied`, `news_published`, `mention`) will publish explicit events to the internal event bus.
- **Unified Configuration via `txtar`**: We will use `txtar` archives to bundle the entire definition of a notification type into a single, cohesive unit. Each archive will contain:
  - **Event Listening Logic & Filters**: Configuration detailing which bus events trigger this notification and under what conditions.
  - **Component Templates**: The actual templates for different delivery methods (e.g., `internal.gohtml` for in-app, `email.gotxt` and `email.gohtml` for emails).
  - **Metadata**: Defining default subscriptions, tier requirements, and role-based access.
- **Memory-Loaded Configuration**: At startup, these `txtar` configurations will be loaded into memory, providing a centralized and highly testable registry of all available notifications.

## 3. Subscription Management & Access Control
To make it easier for users to manage their preferences and to support advanced billing/tiering features:

- **Role-Based Defaults**: Each notification configuration will define default opt-in/opt-out states based on user roles (e.g., admins might default to receiving system alerts, while standard users default to thread replies).
- **Tier-Specific Notifications**: Configurations will specify minimum required tiers (e.g., "Premium" for instant SMS or specific premium digests). The system will strictly enforce these when users attempt to subscribe or when delivering notifications.
- **User Configuration Endpoint**: A centralized user interface and API will be generated dynamically from the memory-loaded notification registry, allowing users to easily toggle their preferences for the notifications they have access to.

## 4. Reliability and Testability
- **Testability**: By moving the event logic and templates into unified `txtar` files, we can easily write unit tests that mock the event bus, trigger an event, and verify the correct templates are rendered and dispatched to the queue.
- **Reliability**: Decoupling the notification generation from the core request cycle (using the async event bus) ensures that failures in notification generation or delivery do not impact the user's primary actions (e.g., posting a comment).

## 5. Implementation Steps
1. **Define the unified `txtar` format** for notification configurations.
2. **Implement the memory-loaded registry** that parses these `txtar` files on startup.
3. **Refactor existing notifications** (e.g., thread replies, linker comments) into the new `txtar` format.
4. **Update the event bus listeners** to trigger the new registry logic instead of the legacy interface-based handlers.
5. **Implement role-based defaults and tier checks** in the subscription API and delivery pipeline.
6. **Migrate existing user preferences** to the new schema and update the frontend UI.

## 6. Administration and Tooling
To manage the notification configurations effectively without requiring codebase redeployments for text changes:

- **Web Administration UI**:
  - The admin panel will have a dedicated "Notification Templates" section.
  - It will display all loaded `txtar` configurations from the memory registry.
  - Administrators with the `manage_notifications` grant will be able to edit the metadata (e.g., changing required tiers or default roles) and update the email/internal templates directly in the browser.
  - Changes made via the UI will be saved back into the database or configuration directory, overriding the compiled defaults.

- **CLI Tooling**:
  - `goa4web admin notifications list` - Displays all registered notifications and their trigger patterns.
  - `goa4web admin notifications render <event_pattern> --data <path_to_json>` - Allows an admin to test template rendering locally by feeding mock event data into the `txtar` configuration to verify the generated email or internal HTML output.
  - `goa4web admin notifications export` - Dumps the current configurations into raw `.txtar` files for backup or version control.

## 7. Layered Permissions and Tiers Model
The permission system for notifications should be layered to ensure both flexibility and security:

- **Subscription Control**:
  - `Permission to Opt-In`: Dictates if a user role/tier is allowed to subscribe to a notification.
  - `Permission to Opt-Out`: Dictates if a user role/tier is allowed to unsubscribe. Certain system-critical alerts (e.g., account security warnings) might deny opt-out.
- **Resolution Strategy**:
  1. Evaluate if the notification is available at the user's base level (e.g., "everyone").
  2. Apply role-based and tier-based overrides (e.g., "premium tier adds SMS").
  3. Apply the user's personal preferences (opt-in/opt-out) against the remaining allowable set.
- **Preference Persistence**: A user can register their interest (opt-in) or disinterest (opt-out) for any notification in the system. However, the system will actively filter delivery based on their *current* effective permissions (roles/tiers) at the time the event fires.
