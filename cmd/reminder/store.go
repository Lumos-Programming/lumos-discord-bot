package reminder

import (
	"context"
	"time"
)

type DueReminder struct {
	ID   string
	Info ReminderInfoExec
}

type ReminderStore interface {
	UpsertDraft(ctx context.Context, id string, info ReminderInfoExec) error
	Confirm(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (ReminderInfoExec, error)

	ListDueConfirmed(ctx context.Context, now time.Time, limit int) ([]DueReminder, error)
	MarkExecuted(ctx context.Context, id string, when time.Time) error
	SetLastError(ctx context.Context, id string, msg string) error

	Close() error
}
