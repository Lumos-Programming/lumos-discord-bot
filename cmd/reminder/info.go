package reminder

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"time"
)

type ReminderInfo struct {
	title     string
	eventYear string //YYYY
	eventTime string //MMDDHHMM
	setTime   string //1w2d3h4m
}

type ReminderInfoExec struct {
	title       string
	eventTime   time.Time
	triggerTime time.Time
	UserName    string
	UserID      string
	ChannelID   string
	Session     *discordgo.Session
	executed    bool
}

func (r ReminderInfo) validate() (time.Time, time.Time, error) {
	log.Printf("Entering validation")
	var TimeOfEvTime time.Time
	var TimeOfTrTime time.Time
	var parseErr error
	jst, _ := time.LoadLocation("Asia/Tokyo")

	TimeOfEvTime, TimeOfTrTime, parseErr = parseEventtime(r)
	if parseErr != nil {
		log.Println("validationErr: parseErr")
		return time.Time{}, time.Time{}, parseErr
	} else {
		if !(TimeOfEvTime.After(time.Now().In(jst))) {
			log.Println("validationErr: valueErr")
			return time.Time{}, time.Time{}, fmt.Errorf("・イベントの日時は未来の日時を指定してください")
		} else {
			log.Println("validated")
			return TimeOfEvTime, TimeOfTrTime, nil
		}
	}
}

func (r ReminderInfo) generateCustomID() string {
	dateStr := r.eventYear + r.eventTime // YYYYMMDDHHMM
	randNum := rand.Intn(10000)
	randStr := fmt.Sprintf("%04d", randNum)
	customID := fmt.Sprintf("reminder-%s-%s", dateStr, randStr)
	return customID
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
