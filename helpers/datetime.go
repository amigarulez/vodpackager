package helpers

import (
	"fmt"
	"time"
)

// TODO replace this quick and dirty solution with a library
func ToISO(d time.Duration) string {
	// Normalise to a positive value; keep the sign for negative durations.
	neg := d < 0
	if neg {
		d = -d
	}

	// Break the duration down.
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	// Build the string piece‑by‑piece.
	out := ""
	if neg {
		out = "-"
	}
	out += "P"

	// Time part always starts with ‘T’.
	out += "T"
	if hours > 0 {
		out += fmt.Sprintf("%dH", hours)
	}
	if mins > 0 {
		out += fmt.Sprintf("%dM", mins)
	}
	if secs > 0 {
		out += fmt.Sprintf("%dS", secs)
	}

	// Edge case: completely zero → PT0S
	if out == "PT" {
		out = "PT0S"
	}
	return out
}
