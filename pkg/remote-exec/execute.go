package remote_exec

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func Execute(t *testing.T, tag string, host *ssh.Client, commandArgs []string) (string, string) {

	command := strings.Join(commandArgs, " ")

	t.Logf("%s [exec-block] | %s\n", tag, command)

	// Set sudo to read the password from stdin
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		command = fmt.Sprintf(`sudo -S %s`, command)
	}

	if host == nil {
		t.Fatalf("%s | SSH client not initialized. Please connect to remote device first", tag)
	}

	session, err := host.NewSession()
	if err != nil {
		t.Fatalf("%s | failed to create session: %v", tag, err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		t.Fatalf("%s | failed to create stdin pipe: %v", tag, err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatalf("%s | failed to create stdout pipe: %v", tag, err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		t.Fatalf("%s | failed to create stderr pipe: %v", tag, err)
	}

	var stdoutBuffer string
	var stderrBuffer string
	go enterSudoPassword(t, tag, stdin, stderr, &stderrBuffer)
	go copyReaderToBuffer(t, tag, stdout, &stdoutBuffer)

	if err := session.Start(command); err != nil {
		t.Fatalf("%s | failed to start session with command '%s': %v", tag, command, err)
	}

	if err := session.Wait(); err != nil {
		t.Fatalf("%s | command '%s' failed: %v\n%s", tag, command, err, stderrBuffer)
	}

	return stdoutBuffer, stderrBuffer
}

// ExecuteAsync starts a command on the remote host.
// stdout and stderr are copied to the strings pointed to by the passed in pointers
// A boolean pointer is returned, of which the value indicates if the process is still running
func ExecuteAsync(t *testing.T, tag string, host *ssh.Client, commandArgs []string, stdoutBuffer *string, stderrBuffer *string) *bool {

	command := strings.Join(commandArgs, " ")

	t.Logf("%s [exec-async] | %s\n", tag, command)

	// Set sudo to read the password from stdin
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		command = fmt.Sprintf(`sudo -S %s`, command)
	}

	if host == nil {
		t.Fatalf("%s | SSH client not initialized. Please connect to remote device first", tag)
	}

	session, err := host.NewSession()
	if err != nil {
		t.Fatalf("%s | failed to create session: %v", tag, err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		t.Fatalf("%s | failed to create stdin pipe: %v", tag, err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatalf("%s | failed to create stdout pipe: %v", tag, err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		t.Fatalf("%s | failed to create stderr pipe: %v", tag, err)
	}

	go enterSudoPassword(t, tag, stdin, stderr, stderrBuffer)
	go copyReaderToBuffer(t, tag, stdout, stdoutBuffer)

	if err := session.Start(command); err != nil {
		t.Fatalf("%s | failed to start session with command '%s': %v", tag, command, err)
	}

	// Do a session.Wait() and monitor the exit code
	running := true
	go monitorAsyncSession(t, tag, session, &running)

	t.Cleanup(func() {
		closeAsync(t, tag, session)
	})

	return &running
}

func monitorAsyncSession(t *testing.T, tag string, session *ssh.Session, running *bool) {
	err := session.Wait()
	if err != nil {
		*running = false
		if errors.Is(err, &ssh.ExitMissingError{}) {
			t.Logf("%s | session exited without an exit code", tag)
		} else {
			t.Logf("%s | %v", tag, err)
		}
	}
}

func closeAsync(t *testing.T, tag string, session *ssh.Session) {
	err := session.Signal(ssh.SIGTERM)
	if err != nil {
		t.Logf("%s | Failed to send SIGTERM: %v\n", tag, err)
	}
	time.Sleep(1 * time.Second) // Delay is required otherwise the client becomes an orphaned process

	// If sigterm succeeded, kill and session close will fail
	err = session.Signal(ssh.SIGKILL)
	if err != nil {
		if err.Error() == "EOF" {
			// expected error
		} else {
			t.Logf("%s | Failed to send SIGKILL: %v\n", tag, err)
		}
	}
	err = session.Close()
	if err != nil {
		if err.Error() == "EOF" {
			// expected error
		} else {
			t.Logf("%s | failed to close session: %v\n", tag, err)
		}
	}
	time.Sleep(1 * time.Second) // Delay is required otherwise the client becomes an orphaned process
}

// Monitor stderr for the sudo password request, and only pipe it in when it is requested
// https://stackoverflow.com/a/44501303
func enterSudoPassword(t *testing.T, tag string, in io.WriteCloser, out io.Reader, output *string) {
	var (
		line string
		r    = bufio.NewReader(out)
	)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		*output = *output + string(b)

		if b == byte('\n') {
			t.Logf("%s STDERR | %s", tag, line)
			line = ""
			continue
		}

		line += string(b)

		if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
			t.Logf("%s | Remote requested sudo password. Entering.\n", tag)
			_, err = in.Write([]byte(os.Getenv("REMOTE_PASSWORD") + "\n"))
			if err != nil {
				break
			}
		}
	}
}

func copyReaderToBuffer(t *testing.T, tag string, in io.Reader, out *string) {
	var (
		line string
		r    = bufio.NewReader(in)
	)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		*out = *out + string(b)

		if b == byte('\n') {
			t.Logf("%s STDOUT | %s", tag, line)
			line = ""
			continue
		}

		line += string(b)
	}
}
