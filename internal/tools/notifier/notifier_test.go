package notifier

import (
	"context"
	"testing"
)

// MockCommandExec is a function type for mocking command execution
type MockCommandExec func(ctx context.Context, name string, arg ...string) error

func TestNotify(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	mockExec := func(ctx context.Context, name string, arg ...string) error {
		capturedName = name
		capturedArgs = arg
		return nil
	}

	n := &Notifier{
		execFunc: mockExec,
	}

	err := n.Notify(context.Background(), "Bobik", "Hello World")
	if err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	if capturedName != "notify-send" {
		t.Errorf("expected notify-send, got %s", capturedName)
	}

	expectedArgs := []string{"Bobik", "Hello World"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(capturedArgs))
	} else {
		for i, v := range expectedArgs {
			if capturedArgs[i] != v {
				t.Errorf("arg %d: expected %s, got %s", i, v, capturedArgs[i])
			}
		}
	}
}

func TestNotifyError(t *testing.T) {
	mockExec := func(ctx context.Context, name string, arg ...string) error {
		return context.DeadlineExceeded
	}

	n := &Notifier{
		execFunc: mockExec,
	}

	err := n.Notify(context.Background(), "Title", "Message")
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}
