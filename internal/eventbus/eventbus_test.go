package eventbus

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBus_Shutdown(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1 := bus.Subscribe(TaskMessageType)
	ch2 := bus.Subscribe(TaskMessageType)

	var wg sync.WaitGroup
	wg.Add(2)

	// Consumer 1
	go func() {
		defer wg.Done()
		for {
			select {
			case env, ok := <-ch1:
				if !ok {
					return
				}
				env.Ack()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Consumer 2 (Simulate slow processing)
	go func() {
		defer wg.Done()
		for {
			select {
			case env, ok := <-ch2:
				if !ok {
					return
				}
				time.Sleep(10 * time.Millisecond)
				env.Ack()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Publish messages
	for i := 0; i < 5; i++ {
		bus.Publish(TaskEvent{UserID: int32(i)})
	}

	// Shutdown
	done := make(chan error)
	go func() {
		done <- bus.Shutdown(context.Background())
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("shutdown timeout")
	}

	cancel()
	wg.Wait()
}

func TestBus_Ack(t *testing.T) {
	bus := NewBus()
	ch := bus.Subscribe(TaskMessageType)

	bus.Publish(TaskEvent{})

	select {
	case env := <-ch:
		// Ack multiple times should be safe
		env.Ack()
		env.Ack()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}

	if err := bus.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}

func TestBus_Backpressure(t *testing.T) {
	bus := NewBus()
	// Channel size is 1
	ch := bus.Subscribe(TaskMessageType)

	// Fill channel
	bus.Publish(TaskEvent{UserID: 1})

	// Try to publish more (should be dropped but WG handled)
	bus.Publish(TaskEvent{UserID: 2})
	bus.Publish(TaskEvent{UserID: 3})

	// Shutdown should succeed even if we only ack the one message we received
	go func() {
		time.Sleep(10 * time.Millisecond)
		select {
		case env := <-ch:
			env.Ack()
		default:
		}
	}()

	if err := bus.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}

func TestBus_ShutdownContext(t *testing.T) {
	bus := NewBus()
	ch := bus.Subscribe(TaskMessageType)

	bus.Publish(TaskEvent{})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Don't ack message, shutdown should timeout
	if err := bus.Shutdown(ctx); err == nil {
		t.Fatal("expected timeout error")
	}

	// Clean up for race detector (ack the pending message)
	go func() {
		select {
		case env := <-ch:
			env.Ack()
		default:
		}
	}()
}

// mockTask implements tasks.Task and tasks.Name
type mockTask string

func (t mockTask) Name() string {
	return string(t)
}

func (t mockTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func TestNewBus(t *testing.T) {
	bus := NewBus()
	require.NotNil(t, bus)
	assert.False(t, bus.closed)
}

func TestSubscribe(t *testing.T) {
	bus := NewBus()

	t.Run("SubscribeAll", func(t *testing.T) {
		ch1 := bus.Subscribe()
		require.NotNil(t, ch1)
		assert.Equal(t, 1, cap(ch1))
	})

	t.Run("SubscribeSpecific", func(t *testing.T) {
		ch2 := bus.Subscribe(TaskMessageType)
		require.NotNil(t, ch2)
		assert.Equal(t, 1, cap(ch2))
	})
}

func TestPublish(t *testing.T) {
	bus := NewBus()

	chAll := bus.Subscribe()
	chTask := bus.Subscribe(TaskMessageType)
	chEmail := bus.Subscribe(EmailQueueMessageType)

	taskMsg := TaskEvent{
		Task: mockTask("test-task"),
		Time: time.Now(),
	}

	emailMsg := EmailQueueEvent{
		Time: time.Now(),
	}

	t.Run("PublishTaskEvent", func(t *testing.T) {
		err := bus.Publish(taskMsg)
		require.NoError(t, err)

		// Verify chAll received taskMsg
		select {
		case env := <-chAll:
			assert.Equal(t, taskMsg, env.Msg)
			env.Ack()
		case <-time.After(100 * time.Millisecond):
			t.Fatal("chAll did not receive taskMsg")
		}

		// Verify chTask received taskMsg
		select {
		case env := <-chTask:
			assert.Equal(t, taskMsg, env.Msg)
			env.Ack()
		case <-time.After(100 * time.Millisecond):
			t.Fatal("chTask did not receive taskMsg")
		}

		// Verify chEmail did NOT receive taskMsg
		select {
		case env := <-chEmail:
			t.Fatalf("chEmail received unexpected message: %v", env.Msg)
			env.Ack()
		default:
			// OK
		}
	})

	t.Run("PublishEmailQueueEvent", func(t *testing.T) {
		err := bus.Publish(emailMsg)
		require.NoError(t, err)

		// Verify chAll received emailMsg
		select {
		case env := <-chAll:
			assert.Equal(t, emailMsg, env.Msg)
			env.Ack()
		case <-time.After(100 * time.Millisecond):
			t.Fatal("chAll did not receive emailMsg")
		}

		// Verify chEmail received emailMsg
		select {
		case env := <-chEmail:
			assert.Equal(t, emailMsg, env.Msg)
			env.Ack()
		case <-time.After(100 * time.Millisecond):
			t.Fatal("chEmail did not receive emailMsg")
		}

		// Verify chTask did NOT receive emailMsg
		select {
		case env := <-chTask:
			t.Fatalf("chTask received unexpected message: %v", env.Msg)
			env.Ack()
		default:
			// OK
		}
	})
}

func TestPublish_NonBlocking(t *testing.T) {
	bus := NewBus()
	ch := bus.Subscribe()

	msg1 := TaskEvent{Task: mockTask("1")}
	msg2 := TaskEvent{Task: mockTask("2")}

	// Fill the buffer (capacity 1)
	err := bus.Publish(msg1)
	require.NoError(t, err)

	// This should not block, even though channel is full.
	// The message will be dropped for this subscriber.
	done := make(chan struct{})
	go func() {
		err := bus.Publish(msg2)
		assert.NoError(t, err)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Publish blocked on full channel")
	}

	// Verify we received the first message
	select {
	case env := <-ch:
		assert.Equal(t, msg1, env.Msg)
		env.Ack()
	default:
		t.Fatal("Expected msg1 in channel")
	}

	// Verify we DO NOT receive the second message (it was dropped)
	select {
	case env := <-ch:
		t.Fatalf("Received unexpected message (should have been dropped): %v", env.Msg)
		env.Ack()
	default:
		// OK
	}
}

func TestShutdown(t *testing.T) {
	bus := NewBus()
	ch := bus.Subscribe()

	msg := TaskEvent{Task: mockTask("shutdown-test")}
	err := bus.Publish(msg)
	require.NoError(t, err)

	// Shutdown with context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Drain the channel in background so Shutdown can complete
	go func() {
		// Wait a bit to simulate processing time, but less than context timeout
		time.Sleep(50 * time.Millisecond)
		env := <-ch
		env.Ack()
		for env := range ch {
			env.Ack()
		}
	}()

	err = bus.Shutdown(ctx)
	require.NoError(t, err)
	assert.True(t, bus.closed)

	// Verify Publish returns error after shutdown
	err = bus.Publish(msg)
	assert.ErrorIs(t, err, ErrBusClosed)
}

func TestShutdown_Timeout(t *testing.T) {
	bus := NewBus()
	_ = bus.Subscribe() // Subscribe but never read

	msg := TaskEvent{Task: mockTask("timeout-test")}
	bus.Publish(msg) // Fills the buffer

	// Shutdown with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := bus.Shutdown(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestSyncPublish(t *testing.T) {
	bus := NewBus()
	var capturedMsg Message
	var wg sync.WaitGroup
	wg.Add(1)

	bus.SyncPublish = func(msg Message) {
		capturedMsg = msg
		wg.Done()
	}

	msg := TaskEvent{Task: mockTask("sync-test")}
	err := bus.Publish(msg)
	require.NoError(t, err)

	wg.Wait()
	assert.Equal(t, msg, capturedMsg)
}

func TestConcurrentAccess(t *testing.T) {
	bus := NewBus()
	const workers = 10
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	stop := make(chan struct{})

	// Subscribers
	for i := 0; i < workers; i++ {
		go func() {
			ch := bus.Subscribe()
			for {
				select {
				case env, ok := <-ch:
					if ok {
						env.Ack()
					}
				case <-stop:
					return
				}
			}
		}()
	}

	// Publishers
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				bus.Publish(TaskEvent{Task: mockTask("concurrent")})
			}
		}()
	}

	wg.Wait()
	close(stop)
}
