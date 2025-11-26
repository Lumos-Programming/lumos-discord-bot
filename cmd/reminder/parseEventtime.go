package reminder

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

func parseEventtime(r ReminderInfo) (time.Time, time.Time, error) {
	var errMsgs []string
	var timeOfEvTime time.Time
	var timeOfTrTime time.Time
	jst, _ := time.LoadLocation("Asia/Tokyo")

	//format check: eventYear
	if r.eventYear != "" {
		if len(r.eventYear) != 4 || !isAllDigits(r.eventYear) {
			errMsgs = append(errMsgs, "・開催年は4桁の数字にしてください")
		}
	} else {
		r.eventYear = strconv.Itoa(time.Now().In(jst).Year())
	}

	//format check: eventTime
	if len(r.eventTime) != 8 || !isAllDigits(r.eventTime) {
		errMsgs = append(errMsgs, "・開催日時は8桁の数字 (MMDDHHmm)にしてください")
	} else {
		monthInput, _ := strconv.Atoi(r.eventTime[:2])
		dayInput, _ := strconv.Atoi(r.eventTime[2:4])
		hourInput, _ := strconv.Atoi(r.eventTime[4:6])
		minuteInput, _ := strconv.Atoi(r.eventTime[6:8])
		if monthInput < 1 || monthInput > 12 ||
			dayInput < 1 || dayInput > 31 ||
			hourInput < 0 || hourInput > 23 ||
			minuteInput < 0 || minuteInput > 59 {
			errMsgs = append(errMsgs, "・開催日時は有効な月・日・時・分の値にしてください")
		}
	}

	//format check: setTime
	var d time.Duration
	d = 0
	i := 0
	for i < len(r.setTime) {
		num := 0
		lastWasDigit := false
		for i < len(r.setTime) && r.setTime[i] >= '0' && r.setTime[i] <= '9' {
			num = num*10 + int(r.setTime[i]-'0')
			lastWasDigit = true
			i++
		}
		if i == len(r.setTime) {
			errMsgs = append(errMsgs, "・リマインダーのタイミングの単位を指定してください")
			break
		}
		if !lastWasDigit {
			errMsgs = append(errMsgs, "・リマインダーのタイミングの単位の前には数字を入力してください")
			break
		}
		unit := r.setTime[i]
		i++
		switch unit {
		case 'w':
			d += time.Duration(num) * aWeek
		case 'd':
			d += time.Duration(num) * aDay
		case 'h':
			d += time.Duration(num) * anHour
		case 'm':
			d += time.Duration(num) * aMinute
		default:
			errMsgs = append(errMsgs, "・リマインダーのタイミングの単位はw,d,h,mのいずれかにしてください")
		}
	}

	//parse eventTime
	var err error
	timeOfEvTime, err = time.ParseInLocation("200601021504", r.eventYear+r.eventTime, jst)
	if err != nil {
		log.Printf("failed to parse eventTime:%v", err)
		return time.Time{}, time.Time{}, errors.New(strings.Join(errMsgs, "\n"))
	}

	//caliculate triggerTime
	timeOfTrTime = timeOfEvTime.Add(-1 * d)

	if len(errMsgs) != 0 {
		return time.Time{}, time.Time{}, errors.New(strings.Join(errMsgs, "\n"))
	}
	return timeOfEvTime, timeOfTrTime, nil
}
