package reminder

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
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

type DiscordSender interface {
	ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

type storeHolder struct {
	s ReminderStore
}

type senderHolder struct {
	s DiscordSender
}

var store atomic.Value  // storeHolder
var sender atomic.Value // senderHolder

func init() {
	store.Store(storeHolder{})
	sender.Store(senderHolder{})
}

func SetReminderStore(s ReminderStore) {
	store.Store(storeHolder{s: s})
}

func GetReminderStore() ReminderStore {
	return store.Load().(storeHolder).s
}

func SetDiscordSender(s DiscordSender) {
	sender.Store(senderHolder{s: s})
}

func getDiscordSender() DiscordSender {
	return sender.Load().(senderHolder).s
}

func InitReminderStoreFromEnv(ctx context.Context) {
	if !isFirestoreConfigured() {
		log.Printf("reminder: Firestore is not configured; using in-memory reminder storage")
		return
	}

	s, err := newFirestoreReminderStoreFromEnv(ctx)
	if err != nil {
		log.Printf("reminder: Firestore is configured but failed to initialize: %v", err)
		return
	}
	if s == nil {
		return
	}
	SetReminderStore(s)
	log.Printf("reminder: Firestore storage enabled (collection=%s)", s.collection)
}
