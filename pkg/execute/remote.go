package execute

import (
	"io"
	"log"
	"strings"

	"golang.org/x/crypto/ssh"
)

func (t Remote) ExecuteBlocking(cmd string, args ...string) (stdout string, stderr string) {
	log.Printf("Should execute on remote: %s %s", cmd, strings.Join(args, " "))

	return
}

func (t Remote) ExecuteAsync(cmd string, args ...string) (stdin io.WriteCloser, stdout io.ReadCloser, stderr io.ReadCloser) {
	log.Printf("Should execute on remote: %s %s", cmd, strings.Join(args, " "))

	return
}

func (t Remote) Connect(host string, username string, password string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout:         t.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return err
	}

	t.SSHClient = sshClient

	return nil
}

func (t Remote) Disconnect() error {
	return t.SSHClient.Close()
}
