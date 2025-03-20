package remote_exec

import (
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func Connect(t *testing.T, tag string, hostName string, username string, password string) *ssh.Client {
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
		t.Fatalf("%s | %v", tag, err)
	}

	t.Cleanup(func() {
		t.Logf("%s | Disconnecting %s", tag, hostName)
		sshClient.Close()
	})

	return sshClient
}
