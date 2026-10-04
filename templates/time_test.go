package templates_test

import (
	"testing"

	"muttley/templates"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		milliseconds int64
		want         string
	}{
		{0, "00:00.000"},
		{45000, "00:45.000"},
		{45123, "00:45.123"},
		{59999, "00:59.999"},
		{125000, "02:05.000"},
	}
	for _, test := range tests {
		if got := templates.FormatTime(test.milliseconds); got != test.want {
			t.Fatalf("FormatTime(%d) = %s, want %s", test.milliseconds, got, test.want)
		}
	}
}
