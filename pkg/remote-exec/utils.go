package remote_exec

import (
	"strings"
	"time"
)

// WaitFor reads a buffer searching for a substring.
// True is returned if and when the substring is found.
// False is returned if the substring is not found within the timeout duration.
func WaitFor(running *bool, buffer *string, search string, timeout time.Duration) bool {
	start := time.Now()

	for time.Now().Before(start.Add(timeout)) {
		if strings.Contains(*buffer, search) {
			return true
		}
		if !*running {
			return false
		}
	}
	return false
}
