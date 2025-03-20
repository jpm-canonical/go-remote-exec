package remote

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

func Execute(t *testing.T, host *ssh.Client, commandArgs []string) (string, string) {

	command := strings.Join(commandArgs, " ")

	t.Logf("[exec-block] %s\n", command)

	// Set sudo to read the password from stdin
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		command = fmt.Sprintf(`sudo -S %s`, command)
	}

	if host == nil {
		t.Fatal("SSH client not initialized. Please connect to remote device first")
	}

	session, err := host.NewSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	var stderrBuffer string
	go enterSudoPassword(t, stdin, stderr, &stderrBuffer)

	if err := session.Start(command); err != nil {
		t.Fatalf("failed to start session with command '%s': %v", command, err)
	}

	stdoutBuffer, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("failed to read command output: %v", err)
	}

	if err := session.Wait(); err != nil {
		t.Fatalf("command '%s' failed: %v\n%s", command, err, stderrBuffer)
	}

	return string(stdoutBuffer), stderrBuffer
}

// ExecuteAsync starts a command on the remote host.
// stdout and stderr are copied to the strings pointed to by the passed in pointers
// A boolean pointer is returned, of which the value indicates if the process is still running
func ExecuteAsync(t *testing.T, host *ssh.Client, commandArgs []string, stdoutBuffer *string, stderrBuffer *string) *bool {

	command := strings.Join(commandArgs, " ")

	t.Logf("[exec-async] %s\n", command)

	// Set sudo to read the password from stdin
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		command = fmt.Sprintf(`sudo -S %s`, command)
	}

	if host == nil {
		t.Fatalf("SSH client not initialized. Please connect to remote device first")
	}

	session, err := host.NewSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	go enterSudoPassword(t, stdin, stderr, stderrBuffer)
	go copyReaderToBuffer(t, stdout, stdoutBuffer)

	if err := session.Start(command); err != nil {
		t.Fatalf("failed to start session with command '%s': %v", command, err)
	}

	// Do a session.Wait() and monitor the exit code
	running := true
	go monitorAsyncSession(t, session, &running)

	t.Cleanup(func() {
		closeAsync(t, session)
	})

	return &running
}

func monitorAsyncSession(t *testing.T, session *ssh.Session, running *bool) {
	err := session.Wait()
	if err != nil {
		*running = false
		if errors.Is(err, &ssh.ExitMissingError{}) {
			t.Fatal("session exited without an exit code")
		} else {
			t.Fatal(err)
		}
	}
}

func closeAsync(t *testing.T, session *ssh.Session) {
	err := session.Signal(ssh.SIGTERM)
	if err != nil {
		t.Logf("Failed to send SIGTERM: %v\n", err)
	}
	time.Sleep(1 * time.Second) // Delay is required otherwise the client becomes an orphaned process

	// If sigterm succeeded, kill and session close will fail
	err = session.Signal(ssh.SIGKILL)
	if err != nil {
		if err.Error() == "EOF" {
			// expected error
		} else {
			t.Logf("Failed to send SIGKILL: %v\n", err)
		}
	}
	err = session.Close()
	if err != nil {
		if err.Error() == "EOF" {
			// expected error
		} else {
			t.Logf("failed to close session: %v\n", err)
		}
	}
	time.Sleep(1 * time.Second) // Delay is required otherwise the client becomes an orphaned process
}

// Monitor stderr for the sudo password request, and only pipe it in when it is requested
// https://stackoverflow.com/a/44501303
func enterSudoPassword(t *testing.T, in io.WriteCloser, out io.Reader, output *string) {
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
			t.Logf("STDERR | %s", line)
			line = ""
			continue
		}

		line += string(b)

		if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
			t.Logf("Remote requested sudo password. Entering.\n")
			_, err = in.Write([]byte(os.Getenv("REMOTE_PASSWORD") + "\n"))
			if err != nil {
				break
			}
		}
	}
}

func copyReaderToBuffer(t *testing.T, in io.Reader, out *string) {
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
			t.Logf("STDOUT | %s", line)
			line = ""
			continue
		}

		line += string(b)
	}
}
