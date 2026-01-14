//go:build integration

package reminder

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFirestoreReminderStore_EmulatorLifecycle(t *testing.T) {
	host := strings.TrimSpace(os.Getenv(envFirestoreEmulatorHost))
	if host == "" {
		t.Skipf("%s is not set", envFirestoreEmulatorHost)
	}
	if strings.TrimSpace(os.Getenv(envFirestoreProjectID)) == "" {
		t.Setenv(envFirestoreProjectID, "demo-test")
	}
	t.Setenv(envFirestoreServiceAccountJSONFile, "")

	conn, err := net.DialTimeout("tcp", host, 300*time.Millisecond)
	if err != nil {
		t.Skipf("Firestore emulator not reachable at %s", host)
	}
	_ = conn.Close()

	store, err := newFirestoreReminderStoreFromEnv(context.Background())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if store == nil {
		t.Fatalf("expected store, got nil (is emulator running?)")
	}
	t.Cleanup(func() { _ = store.Close() })

	now := time.Now()
	id := fmt.Sprintf("it-%d", now.UnixNano())

	info := ReminderInfoExec{
		title:       "integration",
		eventTime:   now.Add(10 * time.Minute),
		triggerTime: now.Add(-1 * time.Minute),
		UserName:    "tester",
		UserID:      "u1",
		ChannelID:   "c1",
		executed:    false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := store.UpsertDraft(ctx, id, info); err != nil {
		t.Fatalf("UpsertDraft: %v", err)
	}

	got, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.title != info.title || got.UserID != info.UserID || got.ChannelID != info.ChannelID {
		t.Fatalf("unexpected doc: %#v", got)
	}

	if err := store.Confirm(ctx, id); err != nil {
		t.Fatalf("Confirm: %v", err)
	}

	due, err := store.ListDueConfirmed(ctx, now, 10)
	if err != nil {
		t.Fatalf("ListDueConfirmed: %v", err)
	}
	if len(due) == 0 || due[0].ID != id {
		t.Fatalf("expected due reminder with id %s, got %#v", id, due)
	}

	if err := store.MarkExecuted(ctx, id, now); err != nil {
		t.Fatalf("MarkExecuted: %v", err)
	}

	due2, err := store.ListDueConfirmed(ctx, now, 10)
	if err != nil {
		t.Fatalf("ListDueConfirmed(2): %v", err)
	}
	if len(due2) != 0 {
		t.Fatalf("expected no due reminders after execution, got %#v", due2)
	}

	if err := store.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
