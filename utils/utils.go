package utils

import (
	"strconv"
	"time"
)

// FormatDate formats a time.Time into a human-readable date
func FormatDate(t time.Time) string {
	return t.Format("Jan 02, 2006")
}

// FormatDateTime formats a time.Time into a human-readable date and time
func FormatDateTime(t time.Time) string {
	return t.Format("Jan 02, 2006 15:04")
}

// FormatDateTimePointer safely formats a *time.Time into a human-readable date and time
func FormatDateTimePointer(t *time.Time) string {
	if t == nil {
		return ""
	}
	return FormatDateTime(*t)
}

// FormatDatePointer safely formats a *time.Time into a human-readable date
func FormatDatePointer(t *time.Time) string {
	if t == nil {
		return ""
	}
	return FormatDate(*t)
}

// GetRemainingDays calculates the number of days between now and a future date
func GetRemainingDays(futureDate *time.Time) int {
	if futureDate == nil {
		return 0
	}
	
	duration := time.Until(*futureDate)
	return int(duration.Hours() / 24)
}

// GetOverdueDays calculates the number of days since a past date
func GetOverdueDays(pastDate *time.Time) int {
	if pastDate == nil {
		return 0
	}
	
	duration := time.Since(*pastDate)
	return int(duration.Hours() / 24)
}

// IntToString converts an int to a string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// GetPageRange returns a range of page numbers for pagination
func GetPageRange(currentPage, totalPages, rangeSize int) []int {
	if totalPages <= 1 {
		return []int{1}
	}

	start := currentPage - rangeSize/2
	if start < 1 {
		start = 1
	}

	end := start + rangeSize - 1
	if end > totalPages {
		end = totalPages
		start = end - rangeSize + 1
		if start < 1 {
			start = 1
		}
	}

	var pages []int
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	return pages
}
