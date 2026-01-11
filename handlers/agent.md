# Handler Guidelines

Tasks in the handlers package must include compile time checks to ensure their concrete types satisfy the `tasks.Task` interface. Declare these using:

```go
var _ tasks.Task = (*myTaskType)(nil)
```

This pattern prevents accidental API drift if method signatures change.


