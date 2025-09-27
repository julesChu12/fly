package utils

import (
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	now1 := Now()
	time.Sleep(time.Millisecond)
	now2 := Now()

	if now1.After(now2) {
		t.Error("Now() should return increasing timestamps")
	}
}

func TestNowUTC(t *testing.T) {
	utcTime := NowUTC()
	if utcTime.Location() != time.UTC {
		t.Error("NowUTC() should return UTC time")
	}
}

func TestUnixNow(t *testing.T) {
	before := time.Now().Unix()
	unix := UnixNow()
	after := time.Now().Unix()

	if unix < before || unix > after {
		t.Errorf("UnixNow() = %d, should be between %d and %d", unix, before, after)
	}
}

func TestUnixMilliNow(t *testing.T) {
	before := time.Now().UnixMilli()
	unixMilli := UnixMilliNow()
	after := time.Now().UnixMilli()

	if unixMilli < before || unixMilli > after {
		t.Errorf("UnixMilliNow() = %d, should be between %d and %d", unixMilli, before, after)
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
	expected := "2023-12-25T15:30:45Z"

	if got := FormatTime(testTime); got != expected {
		t.Errorf("FormatTime() = %v, want %v", got, expected)
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid ISO 8601",
			input:     "2023-12-25T15:30:45Z",
			wantError: false,
		},
		{
			name:      "valid with timezone",
			input:     "2023-12-25T15:30:45+08:00",
			wantError: false,
		},
		{
			name:      "invalid format",
			input:     "2023-12-25 15:30:45",
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTime(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseTime() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		want      bool
	}{
		{
			name:      "future timestamp",
			timestamp: time.Now().Add(time.Hour).Unix(),
			want:      false,
		},
		{
			name:      "past timestamp",
			timestamp: time.Now().Add(-time.Hour).Unix(),
			want:      true,
		},
		{
			name:      "current timestamp",
			timestamp: time.Now().Unix(),
			want:      false, // might be false due to timing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExpired(tt.timestamp); got != tt.want {
				// Allow some tolerance for "current timestamp" test
				if tt.name != "current timestamp" {
					t.Errorf("IsExpired() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMinutes(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  time.Duration
	}{
		{"zero minutes", 0, 0},
		{"positive minutes", 5, 5 * time.Minute},
		{"negative minutes", -3, -3 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Minutes(tt.input); got != tt.want {
				t.Errorf("Minutes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHours(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  time.Duration
	}{
		{"zero hours", 0, 0},
		{"positive hours", 2, 2 * time.Hour},
		{"negative hours", -1, -1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hours(tt.input); got != tt.want {
				t.Errorf("Hours() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDays(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  time.Duration
	}{
		{"zero days", 0, 0},
		{"positive days", 3, 3 * 24 * time.Hour},
		{"negative days", -1, -1 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Days(tt.input); got != tt.want {
				t.Errorf("Days() = %v, want %v", got, tt.want)
			}
		})
	}
}
