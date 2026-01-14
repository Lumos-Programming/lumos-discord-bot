package reminder

import (
	"context"
	"fmt"
	"log"
	"time"
)

const (
	executionInterval = 10
)

func (r *ReminderInfoExec) ShouldRun() bool {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	return !r.executed && r.triggerTime.Before(time.Now().In(jst))
}

func (r *ReminderInfoExec) Run() {
	log.Printf("Executed reminder:  %s  %s", r.title, r.eventTime)
	_ = r.SendRemindMessage()
	r.executed = true
}

func (r *ReminderInfoExec) SendRemindMessage() error {
	s := getDiscordSender()
	if s == nil {
		return fmt.Errorf("discord sender is not set")
	}
	message := fmt.Sprintf("<@%s> Reminder: %s, %s", r.UserID, r.title, r.eventTime)
	_, err := s.ChannelMessageSend(r.ChannelID, message)
	if err != nil {
		log.Printf("Error sending reminder message to channel %s: %v", r.ChannelID, err)
		return err
	}
	log.Printf("Successfully sent reminder message to %s", r.ChannelID)
	return nil
}

func (r *ReminderRepository) CheckAndExecute() {
	if GetReminderStore() != nil {
		r.checkAndExecuteStore(context.Background())
		return
	}
	r.reminderStatus.Range(func(key, value interface{}) bool {
		reminder, ok := value.(ReminderInfoExec)
		if !ok {
			log.Printf("Invalid reminder type for key %v: got %T", key, value)
			return true
		}
		if reminder.ShouldRun() {
			(&reminder).Run()
			r.reminderStatus.Store(key, reminder)
		}
		return true
	})
}

func (r *ReminderRepository) checkAndExecuteStore(ctx context.Context) {
	if getDiscordSender() == nil {
		return
	}
	s := GetReminderStore()
	if s == nil {
		return
	}
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst)

	due, err := s.ListDueConfirmed(ctx, now, 50)
	if err != nil {
		log.Printf("reminder: store query failed: %v", err)
		return
	}

	for _, dueItem := range due {
		info := dueItem.Info
		if err := info.SendRemindMessage(); err != nil {
			log.Printf("reminder: send failed (id=%s): %v", dueItem.ID, err)
			_ = s.SetLastError(ctx, dueItem.ID, err.Error())
			continue
		}
		if err := s.MarkExecuted(ctx, dueItem.ID, time.Now()); err != nil {
			log.Printf("reminder: store MarkExecuted failed (id=%s): %v", dueItem.ID, err)
		}
	}
}

// RemindChecker リマインダーチェックのバックグラウンドプロセス
func (r *ReminderRepository) RemindChecker() {
	log.Printf("Started reminder executing checker")
	go func() {
		ticker := time.NewTicker(executionInterval * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			log.Printf("%d seconds left", executionInterval)
			r.CheckAndExecute()
		}
	}()
}
