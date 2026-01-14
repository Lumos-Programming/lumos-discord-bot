package reminder

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	envFirestoreProjectID              = "FIRESTORE_PROJECT_ID"
	envFirestoreServiceAccountJSONFile = "FIRESTORE_SERVICE_ACCOUNT_JSON_FILE"
	envFirestoreEmulatorHost           = "FIRESTORE_EMULATOR_HOST"
)

const (
	reminderStateDraft     = "draft"
	reminderStateConfirmed = "confirmed"
)

type firestoreReminderStore struct {
	client     *firestore.Client
	collection string
	col        *firestore.CollectionRef
}

type reminderDoc struct {
	Title         string    `firestore:"title"`
	EventTime     time.Time `firestore:"eventTime"`
	TriggerTime   time.Time `firestore:"triggerTime"`
	UserName      string    `firestore:"userName"`
	UserID        string    `firestore:"userId"`
	ChannelID     string    `firestore:"channelId"`
	State         string    `firestore:"state"`
	Executed      bool      `firestore:"executed"`
	CreatedAt     time.Time `firestore:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt"`
	ExecutedAt    time.Time `firestore:"executedAt"`
	LastError     string    `firestore:"lastError"`
	LastAttemptAt time.Time `firestore:"lastAttemptAt"`
}

func isFirestoreConfigured() bool {
	if strings.TrimSpace(os.Getenv(envFirestoreEmulatorHost)) != "" {
		return true
	}
	return strings.TrimSpace(os.Getenv(envFirestoreServiceAccountJSONFile)) != ""
}

func newFirestoreReminderStoreFromEnv(ctx context.Context) (*firestoreReminderStore, error) {
	if !isFirestoreConfigured() {
		return nil, nil
	}

	projectID := strings.TrimSpace(os.Getenv(envFirestoreProjectID))
	emulatorHost := strings.TrimSpace(os.Getenv(envFirestoreEmulatorHost))
	credsFile := strings.TrimSpace(os.Getenv(envFirestoreServiceAccountJSONFile))

	var client *firestore.Client
	var err error
	if emulatorHost != "" {
		if projectID == "" {
			projectID = "demo-test"
		}
		client, err = firestore.NewClient(ctx, projectID, option.WithoutAuthentication())
		if err != nil {
			return nil, err
		}
	} else {
		if projectID == "" {
			return nil, fmt.Errorf("%s is required", envFirestoreProjectID)
		}
		if credsFile == "" {
			return nil, fmt.Errorf("Firestore credentials are required; set %s", envFirestoreServiceAccountJSONFile)
		}
		client, err = firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credsFile))
		if err != nil {
			return nil, err
		}
	}

	collection := "reminders"

	return &firestoreReminderStore{
		client:     client,
		collection: collection,
		col:        client.Collection(collection),
	}, nil
}

func (s *firestoreReminderStore) Close() error {
	return s.client.Close()
}

func (s *firestoreReminderStore) UpsertDraft(ctx context.Context, id string, info ReminderInfoExec) error {
	now := time.Now()
	_, err := s.col.Doc(id).Set(ctx, map[string]interface{}{
		"title":       info.title,
		"eventTime":   info.eventTime,
		"triggerTime": info.triggerTime,
		"userName":    info.UserName,
		"userId":      info.UserID,
		"channelId":   info.ChannelID,
		"state":       reminderStateDraft,
		"executed":    false,
		"lastError":   "",
		"createdAt":   now,
		"updatedAt":   now,
	}, firestore.MergeAll)
	return err
}

func (s *firestoreReminderStore) Confirm(ctx context.Context, id string) error {
	now := time.Now()
	_, err := s.col.Doc(id).Set(ctx, map[string]interface{}{
		"state":     reminderStateConfirmed,
		"executed":  false,
		"lastError": "",
		"updatedAt": now,
	}, firestore.MergeAll)
	return err
}

func (s *firestoreReminderStore) Delete(ctx context.Context, id string) error {
	_, err := s.col.Doc(id).Delete(ctx)
	return err
}

func (s *firestoreReminderStore) Get(ctx context.Context, id string) (ReminderInfoExec, error) {
	snap, err := s.col.Doc(id).Get(ctx)
	if err != nil {
		return ReminderInfoExec{}, err
	}
	var doc reminderDoc
	if err := snap.DataTo(&doc); err != nil {
		return ReminderInfoExec{}, err
	}
	return reminderInfoExecFromDoc(doc), nil
}

func (s *firestoreReminderStore) ListDueConfirmed(ctx context.Context, now time.Time, limit int) ([]DueReminder, error) {
	iter := s.col.
		Where("state", "==", reminderStateConfirmed).
		Where("executed", "==", false).
		Where("triggerTime", "<=", now).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()

	var out []DueReminder
	for {
		snap, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		var doc reminderDoc
		if err := snap.DataTo(&doc); err != nil {
			return nil, err
		}
		out = append(out, DueReminder{
			ID:   snap.Ref.ID,
			Info: reminderInfoExecFromDoc(doc),
		})
	}
	return out, nil
}

func (s *firestoreReminderStore) MarkExecuted(ctx context.Context, id string, when time.Time) error {
	_, err := s.col.Doc(id).Set(ctx, map[string]interface{}{
		"executed":      true,
		"executedAt":    when,
		"updatedAt":     when,
		"lastError":     "",
		"lastAttemptAt": when,
	}, firestore.MergeAll)
	return err
}

func (s *firestoreReminderStore) SetLastError(ctx context.Context, id string, msg string) error {
	now := time.Now()
	_, err := s.col.Doc(id).Set(ctx, map[string]interface{}{
		"lastError":     msg,
		"lastAttemptAt": now,
		"updatedAt":     now,
	}, firestore.MergeAll)
	return err
}

func reminderInfoExecFromDoc(doc reminderDoc) ReminderInfoExec {
	return ReminderInfoExec{
		title:       doc.Title,
		eventTime:   doc.EventTime,
		triggerTime: doc.TriggerTime,
		UserName:    doc.UserName,
		UserID:      doc.UserID,
		ChannelID:   doc.ChannelID,
		executed:    doc.Executed,
	}
}

var _ ReminderStore = (*firestoreReminderStore)(nil)
