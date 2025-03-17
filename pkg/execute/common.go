package execute

import (
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

type Remote struct {
	SSHClient *ssh.Client

	Timeout     time.Duration
	Environment map[string]string
}

type Local struct {
	Timeout     int
	Environment map[string]string
}

type Target interface {
	//Connect() error // custom between remote and local, so don't add to interface
	Disconnect() error
	ExecuteBlocking(string, ...string) (stdout string, stderr string)
	ExecuteAsync(string, ...string) (stdin io.WriteCloser, stdout io.ReadCloser, stderr io.ReadCloser)
}
