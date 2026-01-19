package notifier

import (
	"context"
	"os/exec"
)

type commandExec func(ctx context.Context, name string, arg ...string) error

// Notifier handles system notifications using notify-send.
type Notifier struct {
	execFunc commandExec
}

// New creates a new Notifier.
func New() *Notifier {
	return &Notifier{
		execFunc: func(ctx context.Context, name string, arg ...string) error {
			return exec.CommandContext(ctx, name, arg...).Run()
		},
	}
}

// Notify sends a system notification.
func (n *Notifier) Notify(ctx context.Context, title, message string) error {
	return n.execFunc(ctx, "notify-send", title, message)
}
