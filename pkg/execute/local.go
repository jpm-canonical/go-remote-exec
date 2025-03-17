package execute

import (
	"io"
	"log"
	"strings"
)

func (t Local) ExecuteBlocking(cmd string, args ...string) (stdout string, stderr string) {
	log.Printf("Should execute on local: %s %s", cmd, strings.Join(args, " "))

	return
}

func (t Local) ExecuteAsync(cmd string, args ...string) (stdin io.WriteCloser, stdout io.ReadCloser, stderr io.ReadCloser) {
	log.Printf("Should execute on local: %s %s", cmd, strings.Join(args, " "))

	return
}

func (t Local) Connect() error {
	// nop
	return nil
}

func (t Local) Disconnect() error {
	// nop
	return nil
}
