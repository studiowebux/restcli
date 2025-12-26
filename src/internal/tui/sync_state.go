package tui

import (
	"context"
	"sync"
)

// StreamState manages streaming request state with thread safety
type StreamState struct {
	mu     sync.Mutex
	active bool
	cancel context.CancelFunc
}

// IsActive returns whether streaming is active
func (s *StreamState) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// Start marks stream as active and stores cancel function
func (s *StreamState) Start(cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	s.cancel = cancel
}

// Cancel stops the stream if active
func (s *StreamState) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active && s.cancel != nil {
		s.cancel()
		s.active = false
		s.cancel = nil
	}
}

// Stop marks stream as inactive without calling cancel
func (s *StreamState) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
	s.cancel = nil
}

// RequestState manages regular request cancellation with thread safety
type RequestState struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

// SetCancel stores the cancel function
func (r *RequestState) SetCancel(cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancel = cancel
}

// Cancel cancels the request if active
func (r *RequestState) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
}

// Clear removes the cancel function without calling it
func (r *RequestState) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancel = nil
}

// WebSocketState manages WebSocket connection state with thread safety
type WebSocketState struct {
	mu              sync.Mutex
	active          bool
	cancel          context.CancelFunc
	droppedMessages int64 // Count of messages dropped due to full channel
}

// IsActive returns whether WebSocket is connected
func (w *WebSocketState) IsActive() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.active
}

// Start marks WebSocket as active and resets dropped messages counter
func (w *WebSocketState) Start(cancel context.CancelFunc) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.active = true
	w.cancel = cancel
	w.droppedMessages = 0 // Reset counter on new connection
}

// Cancel disconnects the WebSocket
func (w *WebSocketState) Cancel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	w.active = false
}

// Stop marks WebSocket as inactive without calling cancel
func (w *WebSocketState) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.active = false
	w.cancel = nil
}

// IncrementDropped increments the dropped messages counter
func (w *WebSocketState) IncrementDropped() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.droppedMessages++
}

// GetDroppedMessages returns the count of dropped messages
func (w *WebSocketState) GetDroppedMessages() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.droppedMessages
}

// ResetDropped resets the dropped messages counter
func (w *WebSocketState) ResetDropped() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.droppedMessages = 0
}
