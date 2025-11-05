package reminder

import (
	"fmt"
	"strconv"
)

func validateEventTime(eventTime string) error {
	if len(eventTime) != 8 || !isAllDigits(eventTime) {
		return fmt.Errorf("invalid format: eventTime")
	} else {
		monthInput, _ := strconv.Atoi(eventTime[:2])
		dayInput, _ := strconv.Atoi(eventTime[2:4])
		hourInput, _ := strconv.Atoi(eventTime[4:6])
		minuteInput, _ := strconv.Atoi(eventTime[6:8])
		if monthInput < 1 || monthInput > 12 ||
			dayInput < 1 || dayInput > 31 ||
			hourInput < 0 || hourInput > 23 ||
			minuteInput < 0 || minuteInput > 59 {
			return fmt.Errorf("invalid value: eventTime")
		}
	}
	return nil
}
