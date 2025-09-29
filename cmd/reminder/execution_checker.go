package reminder

import (
	"log"
	"time"
)

func (r *ReminderInfoExec) ShouldRun() bool {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	return !r.executed && r.triggerTime.Before(time.Now().In(jst))
}

func (r *ReminderInfoExec) Run() {
	log.Printf("It's Time!!!  %s  %s", r.title, r.eventTime)
	r.executed = true
}

// RemindChecker リマインダーチェックのバックグラウンドプロセス
func (r *ReminderRepository) RemindChecker() {
	log.Printf("Started reminder executing checker")
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Printf("10 seconds left")
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
		}
	}()
}
