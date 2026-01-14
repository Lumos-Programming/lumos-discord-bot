package reminder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

type fakeSender struct {
	calls    int
	failWith error
}

func (f *fakeSender) ChannelMessageSend(channelID string, content string, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	f.calls++
	if f.failWith != nil {
		return nil, f.failWith
	}
	return &discordgo.Message{}, nil
}

type fakeReminderStore struct {
	due []DueReminder

	upsertDraftIDs []string
	confirmIDs     []string
	deleteIDs      []string
	markedIDs      []string
	lastErrors     map[string]string
}

func (f *fakeReminderStore) UpsertDraft(ctx context.Context, id string, info ReminderInfoExec) error {
	f.upsertDraftIDs = append(f.upsertDraftIDs, id)
	return nil
}

func (f *fakeReminderStore) Confirm(ctx context.Context, id string) error {
	f.confirmIDs = append(f.confirmIDs, id)
	return nil
}

func (f *fakeReminderStore) Delete(ctx context.Context, id string) error {
	f.deleteIDs = append(f.deleteIDs, id)
	return nil
}

func (f *fakeReminderStore) Get(ctx context.Context, id string) (ReminderInfoExec, error) {
	for _, due := range f.due {
		if due.ID == id {
			return due.Info, nil
		}
	}
	return ReminderInfoExec{}, errors.New("not found")
}

func (f *fakeReminderStore) ListDueConfirmed(ctx context.Context, now time.Time, limit int) ([]DueReminder, error) {
	return f.due, nil
}

func (f *fakeReminderStore) MarkExecuted(ctx context.Context, id string, when time.Time) error {
	f.markedIDs = append(f.markedIDs, id)
	return nil
}

func (f *fakeReminderStore) SetLastError(ctx context.Context, id string, msg string) error {
	if f.lastErrors == nil {
		f.lastErrors = map[string]string{}
	}
	f.lastErrors[id] = msg
	return nil
}

func (f *fakeReminderStore) Close() error { return nil }

func TestReminderRepository_CheckAndExecute_StoreMarksExecutedOnSuccess(t *testing.T) {
	store := &fakeReminderStore{
		due: []DueReminder{
			{ID: "r1", Info: ReminderInfoExec{title: "t1", ChannelID: "c1", UserID: "u1"}},
		},
	}
	sender := &fakeSender{}
	SetDiscordSender(sender)
	t.Cleanup(func() { SetDiscordSender(nil) })
	SetReminderStore(store)
	t.Cleanup(func() { SetReminderStore(nil) })
	repo := &ReminderRepository{}

	repo.CheckAndExecute()

	if sender.calls != 1 {
		t.Fatalf("expected 1 send call, got %d", sender.calls)
	}
	if len(store.markedIDs) != 1 || store.markedIDs[0] != "r1" {
		t.Fatalf("expected MarkExecuted for r1, got: %#v", store.markedIDs)
	}
	if len(store.lastErrors) != 0 {
		t.Fatalf("expected no errors, got: %#v", store.lastErrors)
	}
}

func TestReminderRepository_CheckAndExecute_StoreSetsErrorOnSendFailure(t *testing.T) {
	store := &fakeReminderStore{
		due: []DueReminder{
			{ID: "r2", Info: ReminderInfoExec{title: "t2", ChannelID: "c2", UserID: "u2"}},
		},
	}
	sender := &fakeSender{failWith: errors.New("send failed")}
	SetDiscordSender(sender)
	t.Cleanup(func() { SetDiscordSender(nil) })
	SetReminderStore(store)
	t.Cleanup(func() { SetReminderStore(nil) })
	repo := &ReminderRepository{}

	repo.CheckAndExecute()

	if sender.calls != 1 {
		t.Fatalf("expected 1 send call, got %d", sender.calls)
	}
	if len(store.markedIDs) != 0 {
		t.Fatalf("expected no MarkExecuted, got: %#v", store.markedIDs)
	}
	if store.lastErrors["r2"] == "" {
		t.Fatalf("expected last error set for r2")
	}
}

func TestReminderRepository_StoreInfo_UsesStoreAndDeletesDraft(t *testing.T) {
	store := &fakeReminderStore{}
	SetReminderStore(store)
	t.Cleanup(func() { SetReminderStore(nil) })
	repo := &ReminderRepository{}

	repo.reminders.Store("id1", ReminderInfoExec{title: "x"})
	if err := repo.StoreInfo("id1", ReminderInfoExec{title: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.confirmIDs) != 1 || store.confirmIDs[0] != "id1" {
		t.Fatalf("expected Confirm called with id1, got: %#v", store.confirmIDs)
	}
	if _, ok := repo.reminders.Load("id1"); ok {
		t.Fatalf("expected draft deleted from reminders map")
	}
}

func TestReminderRepository_DeleteDraft_UsesStore(t *testing.T) {
	store := &fakeReminderStore{}
	SetReminderStore(store)
	t.Cleanup(func() { SetReminderStore(nil) })
	repo := &ReminderRepository{}

	repo.reminders.Store("id2", ReminderInfoExec{title: "x"})
	if err := repo.DeleteDraft("id2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.deleteIDs) != 1 || store.deleteIDs[0] != "id2" {
		t.Fatalf("expected Delete called with id2, got: %#v", store.deleteIDs)
	}
	if _, ok := repo.reminders.Load("id2"); ok {
		t.Fatalf("expected draft deleted from reminders map")
	}
}

func TestReminderRepository_HoldInfo_UsesStoreUpsert(t *testing.T) {
	store := &fakeReminderStore{}
	SetReminderStore(store)
	t.Cleanup(func() { SetReminderStore(nil) })
	repo := &ReminderRepository{}

	if err := repo.HoldInfo("id3", ReminderInfoExec{title: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.upsertDraftIDs) != 1 || store.upsertDraftIDs[0] != "id3" {
		t.Fatalf("expected UpsertDraft called with id3, got: %#v", store.upsertDraftIDs)
	}
}
