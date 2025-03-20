package linuxptp_testing

import (
	"testing"

	remote "go-remote-exec/pkg/remote-exec"
	"golang.org/x/crypto/ssh"
)

// GetConfigDirectory returns the correct config directory depending on the installation type
func GetConfigDirectory(t *testing.T, installType remote.InstallType) string {
	remotePath := ""
	if installType == remote.Deb {
		remotePath = "/etc/linuxptp/"
	} else if installType == remote.Snap {
		remotePath = "/var/snap/linuxptp/common/"
	}
	return remotePath
}

func CreatePtp4lSnapUds(t *testing.T, host *ssh.Client) {
	// Also make sure the directory exists - not sure if this is required on a fresh install or not
	//command := []string{"sudo", "mkdir", "-p", "/var/snap.linuxptp"}
	//Execute(t, host, command)
}
