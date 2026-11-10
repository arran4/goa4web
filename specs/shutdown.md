# Process shutdown

`goa4web serve` handles `SIGINT` and `SIGTERM` in two stages. The first signal
restores the operating system's default signal handlers and starts cooperative
shutdown. A second signal therefore terminates the process promptly even if a
request or worker is blocked.

Cooperative shutdown has a five-second bound. The server stops accepting HTTP
work and waits for active handlers, closes the in-memory event bus so accepted
envelopes can drain, then cancels and waits for managed worker goroutines. When
the deadline expires, active HTTP connections are closed, cancellable worker
operations are interrupted, and database/cache resources are closed without an
unbounded worker wait.

External-link metadata requests carry worker cancellation through DNS lookup,
redirect validation, HTTP requests, response parsing, and retry waits. Immediate
remote-image cache fetches run as managed background work. Their network and
cache operations are cancellable; short failure bookkeeping uses a separate
bounded context so the persistent pending record remains eligible for retry.

The event bus is volatile. Graceful shutdown gives accepted in-memory work a
bounded chance to finish, but a deadline, second signal, or crash can still lose
unfinished bus work. No resume or durable-job guarantee is provided.
