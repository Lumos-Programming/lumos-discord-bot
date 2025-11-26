package reminder

import (
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
	r.SendRemindMessage()
	r.executed = true
}

func (r *ReminderInfoExec) SendRemindMessage() {
	message := fmt.Sprintf("<@%s> Reminder: %s, %s", r.UserID, r.title, r.eventTime)
	_, err := r.Session.ChannelMessageSend(r.ChannelID, message)
	if err != nil {
		log.Printf("Error sending reminder message to channel %s: %v", r.ChannelID, err)
		return
	}
	log.Printf("Successfully sent reminder message to %s", r.ChannelID)
}

func (r *ReminderRepository) CheckAndExecute() {
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

// RemindChecker リマインダーチェックのバックグラウンドプロセス
func (r *ReminderRepository) RemindChecker() {
	log.Printf("Started reminder executing checker")
	go func() {
		ticker := time.NewTicker(executionInterval * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Printf("%d seconds left", executionInterval)
				r.CheckAndExecute()
			}
		}
	}()
}
