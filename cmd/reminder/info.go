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
	errCode   []int
	errMsg    string
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

func (r ReminderInfo) validate() (time.Time, time.Time, error, []int) {
	log.Printf("Entering validation")
	var TimeOfEvTime time.Time
	var TimeOfTrTime time.Time
	var parseErr error
	errCode := make([]int, 3) //eventYear, eventTime, setTimeのどこにエラーがあるか
	jst, _ := time.LoadLocation("Asia/Tokyo")

	TimeOfEvTime, TimeOfTrTime, parseErr, errCode = parseEventtime(r, errCode)
	if parseErr != nil {
		log.Println("validationErr: parseErr")
		return time.Time{}, time.Time{}, parseErr, errCode
	} else {
		if !(TimeOfEvTime.After(time.Now().In(jst))) {
			log.Println("validationErr: valueErr")
			errCode[0] = 1
			log.Printf("errCode[0]=1 in validate()-1")
			errCode[1] = 1
			log.Printf("errCode[1]=1 in validate()-2")
			return time.Time{}, time.Time{}, fmt.Errorf("・イベントの日時は未来の日時を指定してください"), errCode
		} else {
			log.Println("validated")
			return TimeOfEvTime, TimeOfTrTime, nil, errCode
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
