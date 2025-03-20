package remote

import (
	"strings"
	"testing"
	"time"
)

// GetConfigDirectory returns the correct config directory depending on the installation type
func GetConfigDirectory(t *testing.T, installType InstallType) string {
	remotePath := ""
	if installType == Deb {
		remotePath = "/etc/linuxptp/"
	} else if installType == Snap {
		remotePath = "/var/snap/linuxptp/common/"
	}
	return remotePath
}

// WaitFor reads a buffer searching for a substring.
// True is returned if and when the substring is found.
// False is returned if the substring is not found within the timeout duration.
func WaitFor(buffer *string, search string, timeout time.Duration) bool {
	start := time.Now()

	for time.Now().Before(start.Add(timeout)) {
		if strings.Contains(*buffer, search) {
			return true
		}
	}
	return false
}
