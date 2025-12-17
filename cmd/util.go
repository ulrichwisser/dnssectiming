package cmd

import (
	"fmt"
	"strings"
	"time"
)

func sec2str(seconds int64) string {
	var days = seconds / 86400
	seconds = seconds % 86400
	var hours = seconds / 3600
	seconds = seconds % 3600
	var minutes = seconds / 60
	seconds = seconds % 60

	var ret_str string = ""
	if days > 0 {
		ret_str += fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		ret_str += fmt.Sprintf(" %dh", hours)
	}
	if minutes > 0 {
		ret_str += fmt.Sprintf(" %dm", minutes)
	}
	if seconds > 0 {
		ret_str += fmt.Sprintf(" %ds", seconds)
	}

	return strings.TrimSpace(ret_str)
}

func normalizeDay(t time.Time) time.Time {
    y, m, d := t.Date()
    return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}