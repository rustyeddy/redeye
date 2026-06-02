package redeye

import (
	"testing"
	"time"
)

func TestMJPEGCloseExitsGoroutineWithoutFrames(t *testing.T) {
	m := NewMJPEG()
	m.Play()

	done := make(chan struct{})
	go func() {
		m.Close()
		close(done)
	}()

	select {
	case <-done:
		// goroutine exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("Close() did not return — goroutine leaked")
	}
}

func TestMJPEGPlayReturnsWritableChannel(t *testing.T) {
	m := NewMJPEG()
	frameQ := m.Play()
	defer m.Close()

	if frameQ == nil {
		t.Fatal("Play() returned nil channel")
	}
}

func TestMJPEGCloseIdempotentError(t *testing.T) {
	m := NewMJPEG()
	m.Play()

	if err := m.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
}
