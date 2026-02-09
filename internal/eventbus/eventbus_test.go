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

	// Test subscribe all
	ch1 := bus.Subscribe()
	require.NotNil(t, ch1)
	assert.Equal(t, 1, cap(ch1))

	// Test subscribe specific type
	ch2 := bus.Subscribe(TaskMessageType)
	require.NotNil(t, ch2)
	assert.Equal(t, 1, cap(ch2))
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

	// Publish TaskEvent
	err := bus.Publish(taskMsg)
	require.NoError(t, err)

	// Verify chAll received taskMsg
	select {
	case msg := <-chAll:
		assert.Equal(t, taskMsg, msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("chAll did not receive taskMsg")
	}

	// Verify chTask received taskMsg
	select {
	case msg := <-chTask:
		assert.Equal(t, taskMsg, msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("chTask did not receive taskMsg")
	}

	// Verify chEmail did NOT receive taskMsg
	select {
	case msg := <-chEmail:
		t.Fatalf("chEmail received unexpected message: %v", msg)
	default:
		// OK
	}

	// Publish EmailQueueEvent
	err = bus.Publish(emailMsg)
	require.NoError(t, err)

	// Verify chAll received emailMsg
	select {
	case msg := <-chAll:
		assert.Equal(t, emailMsg, msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("chAll did not receive emailMsg")
	}

	// Verify chEmail received emailMsg
	select {
	case msg := <-chEmail:
		assert.Equal(t, emailMsg, msg)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("chEmail did not receive emailMsg")
	}

	// Verify chTask did NOT receive emailMsg
	select {
	case msg := <-chTask:
		t.Fatalf("chTask received unexpected message: %v", msg)
	default:
		// OK
	}
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
	case msg := <-ch:
		assert.Equal(t, msg1, msg)
	default:
		t.Fatal("Expected msg1 in channel")
	}

	// Verify we DO NOT receive the second message (it was dropped)
	select {
	case msg := <-ch:
		t.Fatalf("Received unexpected message (should have been dropped): %v", msg)
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
		<-ch
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
				case <-ch:
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
