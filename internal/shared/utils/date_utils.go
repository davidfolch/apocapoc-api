package utils

import "time"

func ShouldAppearToday(
	frequency string,
	specificDays []int,
	specificDates []int,
	targetDate time.Time,
) bool {
	switch frequency {
	case "DAILY":
		return true
	case "WEEKLY":
		weekday := int(targetDate.Weekday())
		return contains(specificDays, weekday)
	case "MONTHLY":
		day := targetDate.Day()
		return contains(specificDates, day)
	}
	return false
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
