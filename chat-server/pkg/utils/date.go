package utils

import "time"

func Today() string {
	return time.Now().Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
