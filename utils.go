package main

import (
	"fmt"
	"time"
)


func formatDuration(nanoseconds int64) string {
	if nanoseconds == 0 {
		return "0s"
	}

	duration := time.Duration(nanoseconds)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}