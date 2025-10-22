package reminder

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type ReminderInfo struct {
	title     string
	eventYear string //YYYY
	eventTime string //MMDDHHMM
	setTime   string //e.g. '1h30m' (week: 'w', day: 'd', hour: 'h', minute: 'm')
	UserID    string
	ChannelID string
	Session   *discordgo.Session
}

type ReminderInfoExec struct {
	title       string
	eventTime   time.Time
	triggerTime time.Time
	UserID      string
	ChannelID   string
	Session     *discordgo.Session
	executed    bool
}

func (r ReminderInfo) calTime() (time.Time, time.Time) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	evTime, _ := time.ParseInLocation("200601021504", r.eventYear+r.eventTime, jst)
	seTime, _ := parseCustomDuration(r.setTime)
	trTime := evTime.Add(-1 * seTime)
	return evTime, trTime
}

func infoToExec(rmdinfo ReminderInfo) ReminderInfoExec {
	var rmdexec ReminderInfoExec
	rmdexec.title = rmdinfo.title
	rmdexec.eventTime, rmdexec.triggerTime = rmdinfo.calTime()
	rmdexec.UserID = rmdinfo.UserID
	rmdexec.ChannelID = rmdinfo.ChannelID
	rmdexec.Session = rmdinfo.Session
	rmdexec.executed = false
	return rmdexec
}

func (r ReminderInfo) validate() error {
	var errMsgs []string

	// Validate eventYear
	if r.eventYear != "" {
		y, err := strconv.Atoi(r.eventYear)
		if err != nil || len(r.eventYear) != 4 || y < time.Now().Year() {
			errMsgs = append(errMsgs, "開催年は4桁の数字で、今年以降にしてください")
		}
	}

	// Validate eventTime
	if len(r.eventTime) != 8 || !isAllDigits(r.eventTime) {
		errMsgs = append(errMsgs, "開催日時は8桁の数字 (MMDDHHmm) にしてください")
	} else {
		monthInput, _ := strconv.Atoi(r.eventTime[:2])
		dayInput, _ := strconv.Atoi(r.eventTime[2:4])
		hourInput, _ := strconv.Atoi(r.eventTime[4:6])
		minuteInput, _ := strconv.Atoi(r.eventTime[6:8])
		if monthInput < 1 || monthInput > 12 ||
			dayInput < 1 || dayInput > 31 ||
			hourInput < 0 || hourInput > 23 ||
			minuteInput < 0 || minuteInput > 59 {
			errMsgs = append(errMsgs, "開催日時の値が無効です")
		}
	}

	// Validate setTime
	if _, err := parseCustomDuration(r.setTime); err != nil {
		errMsgs = append(errMsgs, "リマインダーのタイミングは '1w2d3h4m' 形式にしてください")
	}

	if len(errMsgs) != 0 {
		err := strings.Join(errMsgs, ", ")
		return errors.New(err)
	}

	return nil
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
