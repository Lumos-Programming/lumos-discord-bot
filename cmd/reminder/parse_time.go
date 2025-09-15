package reminder

import (
	"fmt"
	"time"
)

func parseCustomDuration(s string) (time.Duration, error) {
	var d time.Duration
	i := 0
	for i < len(s) {
		num := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			num = num*10 + int(s[i]-'0')
			i++
		}
		if i == len(s) {
			return 0, fmt.Errorf("invalid format")
		}
		unit := s[i]
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
			return 0, fmt.Errorf("invalid unit: %c", unit)
		}
	}
	return d, nil
}
