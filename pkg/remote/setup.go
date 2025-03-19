package remote

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/ssh"
)

func InstallSnap(t *testing.T, host *ssh.Client) {
	// install linuxptp
	command := []string{"sudo", "snap", "install", "linuxptp"}
	stdout, stderr := Execute(t, host, command)
	fmt.Printf("===stdout===\n%s============\n", stdout)
	fmt.Printf("===stderr===\n%s============\n", stderr)
}
