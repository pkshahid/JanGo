package signals

import (
	"fmt"
	"reflect"
	"sync"
)

// ReceiverFunc is the signature for signal receivers.
type ReceiverFunc func(sender any, kwargs map[string]any)

// Response holds the result of a single receiver.
type Response struct {
	Receiver reflect.Value // To identify which receiver
	Result   any           // Any returned value (usually nil in simple signals)
	Err      error         // Any error encountered or panic recovered
}

type receiverEntry struct {
	fn     ReceiverFunc
	sender any     // The specific sender to filter on, or nil for all
	ptr    uintptr // For identifying the function pointer to allow Disconnect
	async  bool
}

// Signal represents a specific event that receivers can listen to.
type Signal struct {
	mu        sync.RWMutex
	receivers []receiverEntry
	Name      string
}

// NewSignal creates a new signal instance.
func NewSignal(name string) *Signal {
	return &Signal{
		Name:      name,
		receivers: make([]receiverEntry, 0),
	}
}

// Connect attaches a receiver function to the signal.
// If sender is non-nil, the receiver will only fire when Send is called with that exact sender type.
// If async is true, the receiver will be dispatched in a separate goroutine during Send.
func (s *Signal) Connect(receiver ReceiverFunc, sender any, async bool) error {
	if receiver == nil {
		return fmt.Errorf("receiver cannot be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get pointer to the function for uniqueness checking/disconnecting
	ptr := reflect.ValueOf(receiver).Pointer()

	// Check if already connected with this sender
	for _, entry := range s.receivers {
		if entry.ptr == ptr && typesMatch(entry.sender, sender) {
			return nil // Already connected
		}
	}

	s.receivers = append(s.receivers, receiverEntry{
		fn:     receiver,
		sender: sender,
		ptr:    ptr,
		async:  async,
	})
	return nil
}

// Disconnect removes a receiver from the signal.
func (s *Signal) Disconnect(receiver ReceiverFunc, sender any) error {
	if receiver == nil {
		return fmt.Errorf("receiver cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ptr := reflect.ValueOf(receiver).Pointer()
	filtered := make([]receiverEntry, 0, len(s.receivers))
	found := false

	for _, entry := range s.receivers {
		if entry.ptr == ptr && typesMatch(entry.sender, sender) {
			found = true
			continue // skip this one
		}
		filtered = append(filtered, entry)
	}

	if !found {
		return fmt.Errorf("receiver not found")
	}

	s.receivers = filtered
	return nil
}

// Send dispatches the signal to all connected receivers synchronously (unless marked async).
func (s *Signal) Send(sender any, kwargs map[string]any) []Response {
	s.mu.RLock()
	// Copy receivers to avoid holding lock during execution
	targets := make([]receiverEntry, 0)
	for _, r := range s.receivers {
		if r.sender == nil || typesMatch(r.sender, sender) {
			targets = append(targets, r)
		}
	}
	s.mu.RUnlock()

	var responses []Response
	var wg sync.WaitGroup
	var respMu sync.Mutex

	for _, r := range targets {
		receiver := r
		if receiver.async {
			wg.Add(1)
			go func() {
				defer wg.Done()
				receiver.fn(sender, kwargs)
				// Async responses are typically ignored or hard to correlate, but we log nil
				respMu.Lock()
				responses = append(responses, Response{Receiver: reflect.ValueOf(receiver.fn)})
				respMu.Unlock()
			}()
		} else {
			receiver.fn(sender, kwargs)
			responses = append(responses, Response{Receiver: reflect.ValueOf(receiver.fn)})
		}
	}

	wg.Wait()
	return responses
}

// SendRobust is like Send, but it catches panics from receivers and returns them as errors.
func (s *Signal) SendRobust(sender any, kwargs map[string]any) []Response {
	s.mu.RLock()
	targets := make([]receiverEntry, 0)
	for _, r := range s.receivers {
		if r.sender == nil || typesMatch(r.sender, sender) {
			targets = append(targets, r)
		}
	}
	s.mu.RUnlock()

	var responses []Response
	var wg sync.WaitGroup
	var respMu sync.Mutex

	for _, r := range targets {
		receiver := r
		if receiver.async {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := executeRobust(receiver.fn, sender, kwargs)
				respMu.Lock()
				responses = append(responses, Response{Receiver: reflect.ValueOf(receiver.fn), Err: err})
				respMu.Unlock()
			}()
		} else {
			err := executeRobust(receiver.fn, sender, kwargs)
			responses = append(responses, Response{Receiver: reflect.ValueOf(receiver.fn), Err: err})
		}
	}

	wg.Wait()
	return responses
}

func executeRobust(fn ReceiverFunc, sender any, kwargs map[string]any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()
	fn(sender, kwargs)
	return nil
}

// typesMatch compares two sender types.
// A nil sender matches anything.
// If both are non-nil, we check if their reflection Types are identical.
// In Django, sender is typically a model class (type). Here we pass instances or zero-values.
func typesMatch(registeredSender, actualSender any) bool {
	if registeredSender == nil {
		return true
	}
	if actualSender == nil {
		return false // Registered expects a specific type, but actual is nil
	}

	// We want to match based on type, ignoring whether it's a pointer or value in some cases,
	// but to be strict, we match exact type.
	t1 := reflect.TypeOf(registeredSender)
	t2 := reflect.TypeOf(actualSender)

	// If one is a pointer and the other is not, we might want to match their underlying types
	if t1.Kind() == reflect.Ptr {
		t1 = t1.Elem()
	}
	if t2.Kind() == reflect.Ptr {
		t2 = t2.Elem()
	}

	return t1 == t2
}
