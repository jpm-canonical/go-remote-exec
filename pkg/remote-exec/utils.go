package remote_exec

import (
	"net"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
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

func GetIpV4Address(t *testing.T, tag string, client *ssh.Client, interfaceName string) net.IP {
	command := []string{
		"ip -f inet addr show", interfaceName, " | awk '/inet / {print $2}'",
	}
	stdout, stderr := Execute(t, tag, client, command)
	if stderr != "" {
		t.Error(stderr)
	}
	ip, _, err := net.ParseCIDR(strings.TrimSpace(stdout))
	if err != nil {
		t.Error(err)
	}
	return ip
}
