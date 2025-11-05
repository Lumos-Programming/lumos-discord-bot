package reminder

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strings"
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
	var errMsgs []string
	var TimeOfEvTime time.Time
	var DurationOfStTime time.Duration
	var TimeOfTrTime time.Time
	var parseErr error

	// Validate eventYear
	if err := validateEventYear(r.eventYear); err != nil {
		errMsgs = append(errMsgs, "開催年は4桁の数字で、今年以降にしてください")
	}

	// Validate eventTime
	if err := validateEventTime(r.eventTime); err != nil {
		errMsgs = append(errMsgs, "開催日時は8桁の数字 (MMDDHHmm) で，有効な月・日・時・分の値にしてください")
	}

	// Validate and caliculate setTime
	DurationOfStTime, parseErr = parseCustomDuration(r.setTime)
	if parseErr != nil {
		errMsgs = append(errMsgs, "リマインダーのタイミングは '1w2d3h4m' 形式にしてください")
	}

	if len(errMsgs) != 0 {
		err := strings.Join(errMsgs, ", ")
		return time.Time{}, time.Time{}, errors.New(err)
	}

	//caliculate EventTime TriggerTime
	TimeOfEvTime, _ = invertEventTime(r.eventYear, r.eventTime)
	TimeOfTrTime = TimeOfEvTime.Add(-1 * DurationOfStTime)

	return TimeOfEvTime, TimeOfTrTime, nil
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
