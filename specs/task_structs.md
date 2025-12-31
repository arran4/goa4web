# Task Struct Guidelines

This document outlines conventions for defining new tasks. Tasks represent
an action that may trigger notifications or be referenced by permission
checks. All handlers use the same pattern when introducing new task types.

## Constants

Every package declares `tasks.TaskString` constants for the task names. Each
constant must include a short comment describing the action, as in
`handlers/writings/tasks.go`:

```go
const (
        // TaskSubmitWriting submits a new writing.
        TaskSubmitWriting tasks.TaskString = "Submit writing"
)
```

## Struct Definition

Tasks are implemented as structs embedding `tasks.TaskString` with a comment
explaining their purpose. A package level variable holds the instance used for
routing and registration. Compile-time assertions are used to guarantee the
struct implements `tasks.Task` and any additional interfaces such as
notification providers or search indexers.

```go
// SubmitWritingTask encapsulates creating a new writing.
type SubmitWritingTask struct{ tasks.TaskString }

var submitWritingTask = &SubmitWritingTask{TaskString: TaskSubmitWriting}

var _ tasks.Task = (*SubmitWritingTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*SubmitWritingTask)(nil)
```

Interfaces from `internal/notifications` or other packages may also be
implemented as needed. Declare a `var _ InterfaceName = (*TaskStruct)(nil)`
line for every interface so that the compiler verifies your task provides the
expected methods.

## Registration

Packages expose a `RegisterTasks` function returning a slice of
`tasks.NamedTask` to register all available tasks. The writings package is an
example:

```go
func RegisterTasks() []tasks.NamedTask {
        return []tasks.NamedTask{
                submitWritingTask,
                replyTask,
                editReplyTask,
                // ...
        }
}
```

`tasks.Action` should wrap handlers so the current task is recorded on the
request event and automatically registered. (Note: `tasks.Action` is currently a placeholder concept in some contexts or implemented via `tasks.Task` interface method `Action`).

The `tasks.Task` interface is defined as:

```go
type Task interface {
	Action(w http.ResponseWriter, r *http.Request) any
}
```

`TaskString` implements `Task` with a no-op action, so it must be embedded or `Action` must be implemented.

## Template and Interface Compliance

Many tasks send notifications. When implementing one of the notification
provider interfaces from `internal/notifications`, ensure there are matching
email and internal notification templates. Tests should load the compiled
template sets and verify that the names returned by the task exist. See
`internal/notifications/reply_templates_test.go` for an example of checking
reply templates.

Compile-time interface assertions guarantee that tasks implement the provider
interfaces while the tests confirm the referenced templates are present.
