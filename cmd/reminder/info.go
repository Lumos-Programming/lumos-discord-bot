package reminder

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type ReminderInfo struct {
	title     string
	eventYear string
	eventTime string
	setTime   string
}

func (i ReminderInfo) validate() error {
	var errMsgs []string

	// Validate title (non-empty)
	if i.title == "" {
		errMsgs = append(errMsgs, "イベント名は必須です") // 必須入力の確認はDiscordがやってくれるので必要ないと思う
	}

	// Validate eventYear
	if i.eventYear != "" {
		y, err := strconv.Atoi(i.eventYear)
		if err != nil || len(i.eventYear) != 4 || y < time.Now().Year() {
			errMsgs = append(errMsgs, "開催年は4桁の数字で、今年以降にしてください")
		}
	}

	// Validate eventTime
	if len(i.eventTime) != 8 || !isAllDigits(i.eventTime) {
		errMsgs = append(errMsgs, "開催日時は8桁の数字 (MMDDHHmm) にしてください")
	} else {
		monthInput, _ := strconv.Atoi(i.eventTime[:2])
		dayInput, _ := strconv.Atoi(i.eventTime[2:4])
		hourInput, _ := strconv.Atoi(i.eventTime[4:6])
		minuteInput, _ := strconv.Atoi(i.eventTime[6:8])
		if monthInput < 1 || monthInput > 12 ||
			dayInput < 1 || dayInput > 31 ||
			hourInput < 0 || hourInput > 23 ||
			minuteInput < 0 || minuteInput > 59 {
			errMsgs = append(errMsgs, "開催日時の値が無効です")
		}
	}

	// Validate setTime
	if _, err := parseCustomDuration(i.setTime); err != nil {
		errMsgs = append(errMsgs, "リマインダーのタイミングは '1w2d3h4m' 形式にしてください")
	}

	if len(errMsgs) != 0 {
		err := strings.Join(errMsgs, ", ")
		return errors.New(err)
	}

	return nil
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
