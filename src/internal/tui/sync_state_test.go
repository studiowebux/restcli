package tui

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStreamState_ConcurrentAccess(t *testing.T) {
	state := &StreamState{}
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	state.Start(cancel)

	// Simulate concurrent access from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)

		// Reader goroutine
		go func() {
			defer wg.Done()
			_ = state.IsActive()
		}()

		// Writer goroutine
		go func(iteration int) {
			defer wg.Done()
			if iteration%2 == 0 {
				state.Cancel()
			} else {
				_, cancel2 := context.WithCancel(context.Background())
				defer cancel2()
				state.Start(cancel2)
			}
		}(i)
	}

	wg.Wait()
	// If test completes without panic or data race, success
}

func TestStreamState_Cancel(t *testing.T) {
	state := &StreamState{}
	ctx, cancel := context.WithCancel(context.Background())

	state.Start(cancel)

	if !state.IsActive() {
		t.Error("Expected stream to be active after Start()")
	}

	state.Cancel()

	if state.IsActive() {
		t.Error("Expected stream to be inactive after Cancel()")
	}

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not cancelled")
	}
}

func TestStreamState_CancelIdempotent(t *testing.T) {
	state := &StreamState{}

	// Cancel when not active should not panic
	state.Cancel()
	state.Cancel()

	_, cancel := context.WithCancel(context.Background())
	state.Start(cancel)

	// Multiple cancels should be safe
	state.Cancel()
	state.Cancel()
	state.Cancel()

	if state.IsActive() {
		t.Error("Expected stream to be inactive after multiple cancels")
	}
}

func TestStreamState_Stop(t *testing.T) {
	state := &StreamState{}
	ctx, cancel := context.WithCancel(context.Background())

	state.Start(cancel)

	if !state.IsActive() {
		t.Error("Expected stream to be active after Start()")
	}

	// Stop without cancelling context
	state.Stop()

	if state.IsActive() {
		t.Error("Expected stream to be inactive after Stop()")
	}

	// Context should NOT be cancelled (unlike Cancel())
	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled by Stop()")
	case <-time.After(50 * time.Millisecond):
		// Expected - context still active
	}
}

func TestRequestState_Cancel(t *testing.T) {
	state := &RequestState{}
	ctx, cancel := context.WithCancel(context.Background())

	state.SetCancel(cancel)
	state.Cancel()

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not cancelled")
	}
}

func TestRequestState_CancelIdempotent(t *testing.T) {
	state := &RequestState{}
	ctx, cancel := context.WithCancel(context.Background())

	state.SetCancel(cancel)

	// Multiple cancels should be safe
	state.Cancel()
	state.Cancel()
	state.Cancel()

	// Verify context was cancelled only once
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}

func TestRequestState_CancelWithoutSet(t *testing.T) {
	state := &RequestState{}

	// Cancel when no cancel func set should not panic
	state.Cancel()
	state.Cancel()
}

func TestRequestState_Clear(t *testing.T) {
	state := &RequestState{}
	ctx, cancel := context.WithCancel(context.Background())

	state.SetCancel(cancel)
	state.Clear()

	// Context should NOT be cancelled (unlike Cancel())
	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled by Clear()")
	case <-time.After(50 * time.Millisecond):
		// Expected - context still active
	}

	// Subsequent Cancel() should be safe (no-op)
	state.Cancel()
}

func TestRequestState_ConcurrentAccess(t *testing.T) {
	state := &RequestState{}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)

		// Writer goroutine
		go func() {
			defer wg.Done()
			_, cancel := context.WithCancel(context.Background())
			defer cancel()
			state.SetCancel(cancel)
		}()

		// Cancel goroutine
		go func() {
			defer wg.Done()
			state.Cancel()
		}()
	}

	wg.Wait()
	// If test completes without panic or data race, success
}

func TestWebSocketState_ConcurrentAccess(t *testing.T) {
	state := &WebSocketState{}
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	state.Start(cancel)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)

		// Reader goroutine
		go func() {
			defer wg.Done()
			_ = state.IsActive()
		}()

		// Writer goroutine
		go func(iteration int) {
			defer wg.Done()
			if iteration%2 == 0 {
				state.Cancel()
			} else {
				_, cancel2 := context.WithCancel(context.Background())
				defer cancel2()
				state.Start(cancel2)
			}
		}(i)
	}

	wg.Wait()
	// If test completes without panic or data race, success
}

func TestWebSocketState_Cancel(t *testing.T) {
	state := &WebSocketState{}
	ctx, cancel := context.WithCancel(context.Background())

	state.Start(cancel)

	if !state.IsActive() {
		t.Error("Expected WebSocket to be active after Start()")
	}

	state.Cancel()

	if state.IsActive() {
		t.Error("Expected WebSocket to be inactive after Cancel()")
	}

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not cancelled")
	}
}

func TestWebSocketState_CancelIdempotent(t *testing.T) {
	state := &WebSocketState{}

	// Cancel when not active should not panic
	state.Cancel()
	state.Cancel()

	_, cancel := context.WithCancel(context.Background())
	state.Start(cancel)

	// Multiple cancels should be safe
	state.Cancel()
	state.Cancel()
	state.Cancel()

	if state.IsActive() {
		t.Error("Expected WebSocket to be inactive after multiple cancels")
	}
}

func TestWebSocketState_Stop(t *testing.T) {
	state := &WebSocketState{}
	ctx, cancel := context.WithCancel(context.Background())

	state.Start(cancel)

	if !state.IsActive() {
		t.Error("Expected WebSocket to be active after Start()")
	}

	// Stop without cancelling context
	state.Stop()

	if state.IsActive() {
		t.Error("Expected WebSocket to be inactive after Stop()")
	}

	// Context should NOT be cancelled (unlike Cancel())
	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled by Stop()")
	case <-time.After(50 * time.Millisecond):
		// Expected - context still active
	}
}

func TestWebSocketState_DroppedMessages(t *testing.T) {
	state := &WebSocketState{}

	// Initial count should be 0
	if count := state.GetDroppedMessages(); count != 0 {
		t.Errorf("Expected initial dropped count to be 0, got %d", count)
	}

	// Increment counter
	state.IncrementDropped()
	state.IncrementDropped()
	state.IncrementDropped()

	if count := state.GetDroppedMessages(); count != 3 {
		t.Errorf("Expected dropped count to be 3, got %d", count)
	}

	// Reset counter
	state.ResetDropped()

	if count := state.GetDroppedMessages(); count != 0 {
		t.Errorf("Expected dropped count to be 0 after reset, got %d", count)
	}
}

func TestWebSocketState_DroppedMessagesResetOnStart(t *testing.T) {
	state := &WebSocketState{}

	// Increment counter
	state.IncrementDropped()
	state.IncrementDropped()

	if count := state.GetDroppedMessages(); count != 2 {
		t.Errorf("Expected dropped count to be 2, got %d", count)
	}

	// Start should reset counter
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	state.Start(cancel)

	if count := state.GetDroppedMessages(); count != 0 {
		t.Errorf("Expected dropped count to be 0 after Start(), got %d", count)
	}
}

func TestWebSocketState_ConcurrentDroppedAccess(t *testing.T) {
	state := &WebSocketState{}
	done := make(chan bool)
	increments := 100

	// Multiple goroutines incrementing
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < increments; j++ {
				state.IncrementDropped()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify count
	expected := int64(10 * increments)
	if count := state.GetDroppedMessages(); count != expected {
		t.Errorf("Expected dropped count to be %d, got %d", expected, count)
	}

	// Concurrent reads while writing
	go func() {
		for i := 0; i < 100; i++ {
			state.IncrementDropped()
		}
	}()

	for i := 0; i < 100; i++ {
		_ = state.GetDroppedMessages()
	}
}

// Benchmark concurrent access to verify performance
func BenchmarkStreamState_ConcurrentReads(b *testing.B) {
	state := &StreamState{}
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	state.Start(cancel)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = state.IsActive()
		}
	})
}

func BenchmarkStreamState_ConcurrentWrites(b *testing.B) {
	state := &StreamState{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, cancel := context.WithCancel(context.Background())
			state.Start(cancel)
			state.Cancel()
		}
	})
}
