package reminder

import (
	"fmt"
	"strconv"
	"time"
)

func validateEventYear(eventYear string) error {
	if eventYear != "" {
		y, err := strconv.Atoi(eventYear)
		if err != nil || len(eventYear) != 4 {
			return fmt.Errorf("invalid format: eventYear")
		} else if y < time.Now().Year() {
			return fmt.Errorf("invalid value: eventYear")
		}
	}
	return nil
}
