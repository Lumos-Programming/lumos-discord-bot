package reminder

import (
	"fmt"
	"time"
)

func invertEventTime(eventYear string, eventTime string) (time.Time, error) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	timeOfEvTime, err := time.ParseInLocation("200601021504", eventYear+eventTime, jst)
	if err != nil {
		return time.Time{}, err
	}
	if timeOfEvTime.Sub(time.Now()) < 0 {
		return time.Time{}, fmt.Errorf("invalid value: event is in the past")
	}
	return timeOfEvTime, nil
}
