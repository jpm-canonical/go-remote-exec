package remote

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
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

func CreatePtp4lSnapUds(t *testing.T, host *ssh.Client) {
	// Also make sure the directory exists - not sure if this is required on a fresh install or not
	//command := []string{"sudo", "mkdir", "-p", "/var/snap.linuxptp"}
	//Execute(t, host, command)
}

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
