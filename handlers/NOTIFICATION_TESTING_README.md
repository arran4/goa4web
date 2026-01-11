# Notification Template Testing

This directory contains end-to-end tests for notification templates triggered by various tasks in the application.

## What These Tests Do

These tests verify that:
1. **Notification templates exist** - Email (HTML, text, subject) and internal templates are properly defined
2. **Templates render correctly** - No template errors when given realistic event data
3. **Event data is correct** - The data passed to templates matches what the templates expect
4. **Provider interfaces work** - Tasks implement the correct notification provider interfaces

## Example: Forum Reply Notifications

See `handlers/forum/forum_reply_notifications_test.go` for a complete example that verifies:

```go
// When a user replies to a forum thread
func TestForumReplyNotifications(t *testing.T) {
    // Create realistic event data as it would appear in production
    replyEvent := handlers.CreateTestEvent(
        &ReplyTask{},
        "/forum/topic/5/thread/42",
        123, // user ID
        map[string]any{
            "Username":   "testuser",
            "CommentText": "This is a test reply",
            "CommentURL": "https://example.com/forum/topic/5/thread/42#c999",
            // ... more realistic data
        },
    )
    
    // Test all notification providers for this task
    handlers.TestNotificationTemplates(t, []handlers.NotificationTemplateTest{
        {
            Name:  "ForumReply_SubscriberNotifications",
            Task:  &ReplyTask{},
            Event: replyEvent,
        },
    })
}
```

This test verifies:
- ✅ Subscriber email templates (HTML, text, subject)
- ✅ Admin notification templates
- ✅ Internal notification templates
- ✅ Auto-subscription path generation
- ✅ All templates render without errors

## Writing Tests for Other Tasks

For any task that implements notification interfaces:

```go
func TestMyTaskNotifications(t *testing.T) {
    task := &MyTask{}
    
    // Create event with realistic data your task produces
    evt := handlers.CreateTestEvent(
        task,
        "/path/to/resource",
        userId,
        map[string]any{
            "Username": "user123",
            // Add fields your templates expect
        },
    )
    
    // Test all notification interfaces
    handlers.TestNotificationTemplates(t, []handlers.NotificationTemplateTest{
        {Name: "MyTask", Task: task, Event: evt},
    })
}
```

## Helper Functions

Core testing infrastructure in `handlers/notification_test_helpers.go`:

- `CreateTestEvent()` - Creates realistic event data
- `TestNotificationTemplates()` - Tests all notification interfaces for tasks
- `RequireEmailTemplates()` - Verifies email templates exist and render
- `RequireNotificationTemplate()` - Verifies internal notification templates

## Running Tests

```bash
# Test specific task notifications
go test -v ./handlers/forum -run TestForumReplyNotifications

# Test all notification tests
go test -v ./handlers/... -run "Notification"

# Run all tests
go test ./...
```

## What Gets Verified

For each task, depending on which interfaces it implements:

### SubscribersNotificationTemplateProvider
- Email templates for users subscribed to the thread/topic
- Internal notification templates
- Templates render with event data

### AdminEmailTemplateProvider  
- Admin notification emails (HTML, text, subject)
- Admin internal notifications
- Templates render correctly

### TargetUsersNotificationProvider
- Email templates for specific target users
- Internal notifications for target users

### DirectEmailNotificationTemplateProvider
- Direct email templates (bypassing user preferences)

### AutoSubscribeProvider
- Correct subscription path generation
- Action name matches expected value

## Implementation

Tests use the actual task implementations (not mocks) with realistic event data that matches production usage. This ensures templates work correctly in the real application.
