package reminder

import (
	"context"
	"log"
	"sync/atomic"
)

type storeHolder struct {
	s ReminderStore
}

var store atomic.Value // storeHolder

func init() {
	store.Store(storeHolder{})
}

func SetReminderStore(s ReminderStore) {
	store.Store(storeHolder{s: s})
}

func GetReminderStore() ReminderStore {
	return store.Load().(storeHolder).s
}

func InitReminderStoreFromEnv(ctx context.Context) {
	s, err := newFirestoreReminderStoreFromEnv(ctx)
	if err != nil {
		if isFirestoreConfigured() {
			log.Printf("reminder: Firestore is configured but failed to initialize: %v", err)
		} else {
			log.Printf("reminder: Firestore is not configured; using in-memory reminder storage")
		}
		return
	}
	if s == nil {
		return
	}
	SetReminderStore(s)
	log.Printf("reminder: Firestore storage enabled (collection=%s)", s.collection)
}
