package remote_exec

import (
	"fmt"
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

func GetIPv4Address(t *testing.T, tag string, client *ssh.Client, interfaceName string) (net.IP, error) {
	command := []string{
		"ip -f inet addr show", interfaceName, " | awk '/inet / {print $2}'",
	}
	stdout, stderr := Execute(t, tag, client, command)
	if stderr != "" {
		return nil, fmt.Errorf("can't find IPv4: %s", stderr)
	}
	ip, _, err := net.ParseCIDR(strings.TrimSpace(stdout))
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func GetIPv6GlobalAddress(t *testing.T, tag string, client *ssh.Client, interfaceName string) (net.IP, error) {
	// ip -6 addr ls -deprecated primary dev enp0s20f0u1u4 scope global | awk '/inet6/{print $2}'
	// See https://serverfault.com/a/1167232
	command := []string{
		"ip -6 addr ls -deprecated primary dev", interfaceName, "scope global | awk '/inet6/{print $2}'",
	}
	stdout, stderr := Execute(t, tag, client, command)
	if stderr != "" {
		return nil, fmt.Errorf("can't find global IPv6: %s", stderr)
	}
	ip, _, err := net.ParseCIDR(strings.TrimSpace(stdout))
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func GetIPv6LocalAddress(t *testing.T, tag string, client *ssh.Client, interfaceName string) (net.IP, error) {
	// ip -6 addr ls -deprecated primary dev enp0s20f0u1u4 scope link | awk '/inet6/{print $2}'
	command := []string{
		"ip -6 addr ls -deprecated primary dev", interfaceName, "scope link | awk '/inet6/{print $2}'",
	}
	stdout, stderr := Execute(t, tag, client, command)
	if stderr != "" {
		return nil, fmt.Errorf("can't find local IPv6: %s", stderr)
	}
	ip, _, err := net.ParseCIDR(strings.TrimSpace(stdout))
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func GetMacAddress(t *testing.T, tag string, client *ssh.Client, interfaceName string) (string, error) {
	command := fmt.Sprintf("cat /sys/class/net/%s/address", interfaceName)
	stdout, stderr := Execute(t, tag, client, []string{command})
	if stderr != "" {
		return "", fmt.Errorf("can't find MAC address: %s", stderr)
	}
	return strings.TrimSpace(stdout), nil
}
