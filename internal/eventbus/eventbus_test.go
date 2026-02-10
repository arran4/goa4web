package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"
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
