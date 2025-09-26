package utils

import (
	"time"
)

// Now returns current time
func Now() time.Time {
	return time.Now()
}

// NowUTC returns current UTC time
func NowUTC() time.Time {
	return time.Now().UTC()
}

// UnixNow returns current Unix timestamp
func UnixNow() int64 {
	return time.Now().Unix()
}

// UnixMilliNow returns current Unix timestamp in milliseconds
func UnixMilliNow() int64 {
	return time.Now().UnixMilli()
}

// FormatTime formats time to ISO 8601 string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTime parses ISO 8601 time string
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// IsExpired checks if a timestamp has expired
func IsExpired(timestamp int64) bool {
	return time.Now().Unix() > timestamp
}

// Duration helpers

// Minutes converts minutes to Duration
func Minutes(m int) time.Duration {
	return time.Duration(m) * time.Minute
}

// Hours converts hours to Duration
func Hours(h int) time.Duration {
	return time.Duration(h) * time.Hour
}

// Days converts days to Duration
func Days(d int) time.Duration {
	return time.Duration(d) * 24 * time.Hour
}
