package execute

import (
	"io"
	"log"
	"strings"

	"golang.org/x/crypto/ssh"
)

func (t *Remote) ExecuteBlocking(cmd string, args ...string) (stdoutResponse string, stderrResponse string) {
	command := cmd + " " + strings.Join(args, " ")

	session, err := t.SSHClient.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to create stderr pipe: %v", err)
	}

	if err := session.Start(command); err != nil {
		log.Fatalf("Failed to start session with command '%s': %v", command, err)
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		log.Fatalf("Failed to read stdout: %v", err)
	}
	stdoutResponse = string(output)

	output, err = io.ReadAll(stderr)
	if err != nil {
		log.Fatalf("Failed to read stderr: %v", err)
	}
	stdoutResponse = string(output)

	if err := session.Wait(); err != nil {
		log.Fatalf("Command '%s' failed: %v", command, err)
	}

	return
}

func (t *Remote) ExecuteAsync(cmd string, args ...string) (stdin io.WriteCloser, stdout io.ReadCloser, stderr io.ReadCloser) {
	log.Printf("Should execute on remote: %s %s", cmd, strings.Join(args, " "))

	return
}

func (t *Remote) Connect(host string, username string, password string) error {
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

func (t *Remote) Disconnect() error {
	return t.SSHClient.Close()
}
