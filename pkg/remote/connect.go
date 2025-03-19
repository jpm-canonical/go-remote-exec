package remote

import (
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func Connect(t *testing.T, hostName string, username string, password string) *ssh.Client {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", hostName+":22", config)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		t.Logf("Disconnecting %s", hostName)
		sshClient.Close()
	})

	return sshClient
}
