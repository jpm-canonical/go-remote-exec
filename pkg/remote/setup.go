package remote

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func CreateHost(host string, username string, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

func Setup(host *ssh.Client) error {
	// install linuxptp
	stdout, stderr, err := execute(host, "sudo snap install linuxptp")
	fmt.Printf("===stdout===\n%s============\n", stdout)
	fmt.Printf("===stderr===\n%s============\n", stderr)
	return err
}

func StartServer(host *ssh.Client) (*ssh.Session, *string, *string, error) {
	command := `sudo linuxptp.ptp4l -i eth0 --step_threshold=1 --hwts_filter full --serverOnly 1 -m -f /snap/linuxptp/current/etc/default.cfg`
	var stdoutBuffer string
	var stderrBuffer string
	session, err := executeAsync(host, command, &stdoutBuffer, &stderrBuffer)
	return session, &stdoutBuffer, &stderrBuffer, err
}

func StartClient(host *ssh.Client) (*ssh.Session, *string, *string, error) {
	command := `sudo linuxptp.ptp4l -i eth0 --step_threshold=1 --hwts_filter full --clientOnly 1 -m -f /snap/linuxptp/current/etc/default.cfg`
	var stdoutBuffer string
	var stderrBuffer string
	session, err := executeAsync(host, command, &stdoutBuffer, &stderrBuffer)
	return session, &stdoutBuffer, &stderrBuffer, err
}

func Stop(session *ssh.Session) {
	closeAsync(session)
}

func execute(host *ssh.Client, command string) (string, string, error) {

	fmt.Printf("[exec-wait] %s\n", command)

	// Set sudo to read the password from stdin
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		command = fmt.Sprintf(`sudo -S %s`, command)
	}

	if host == nil {
		return "", "", fmt.Errorf("SSH client not initialized. Please connect to remote device first")
	}

	session, err := host.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	var stderrBuffer string
	go enterSudoPassword(stdin, stderr, &stderrBuffer)

	if err := session.Start(command); err != nil {
		return "", "", fmt.Errorf("failed to start session with command '%s': %v", command, err)
	}

	stdoutBuffer, err := io.ReadAll(stdout)
	if err != nil {
		return "", "", fmt.Errorf("failed to read command output: %v", err)
	}

	if err := session.Wait(); err != nil {
		return "", "", fmt.Errorf("command '%s' failed: %v\n%s", command, err, stderrBuffer)
	}

	return string(stdoutBuffer), string(stderrBuffer), nil
}

func executeAsync(host *ssh.Client, command string, stdoutBuffer *string, stderrBuffer *string) (*ssh.Session, error) {
	fmt.Printf("[exec-async] %s\n", command)

	// Set sudo to read the password from stdin
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		command = fmt.Sprintf(`sudo -S %s`, command)
	}

	if host == nil {
		return nil, fmt.Errorf("SSH client not initialized. Please connect to remote device first")
	}

	session, err := host.NewSession()
	if err != nil {
		return session, fmt.Errorf("failed to create session: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return session, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return session, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return session, fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	go enterSudoPassword(stdin, stderr, stderrBuffer)
	go copyReaderToBuffer(stdout, stdoutBuffer)

	if err := session.Start(command); err != nil {
		return session, fmt.Errorf("failed to start session with command '%s': %v", command, err)
	}

	return session, nil
}

func closeAsync(session *ssh.Session) {
	err := session.Signal(ssh.SIGTERM)
	if err != nil {
		fmt.Printf("failed to send SIGTERM: %v\n", err)
	}
	time.Sleep(1 * time.Second) // Delay is required otherwise the client becomes an orphaned process

	// If sigterm succeeded, kill and session close will fail
	err = session.Signal(ssh.SIGKILL)
	if err != nil {
		if err.Error() == "EOF" {
			// expected error
		} else {
			fmt.Printf("failed to send SIGKILL: %v\n", err)
		}
	}
	err = session.Close()
	if err != nil {
		if err.Error() == "EOF" {
			// expected error
		} else {
			fmt.Printf("failed to close session: %v\n", err)
		}
	}
	time.Sleep(1 * time.Second) // Delay is required otherwise the client becomes an orphaned process
}

// Monitor stderr for the sudo password request, and only pipe it in when it is requested
// https://stackoverflow.com/a/44501303
func enterSudoPassword(in io.WriteCloser, out io.Reader, output *string) {
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
			line = ""
			continue
		}

		line += string(b)

		if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
			fmt.Printf("Remote requested sudo password. Entering.\n")
			_, err = in.Write([]byte(os.Getenv("REMOTE_PASSWORD") + "\n"))
			if err != nil {
				break
			}
		}
	}
}

func copyReaderToBuffer(in io.Reader, out *string) {
	var (
		r = bufio.NewReader(in)
	)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		*out = *out + string(b)
	}
}

func WaitFor(buffer *string, search string, timeout time.Duration) error {
	start := time.Now()

	for time.Now().Before(start.Add(timeout)) {
		if strings.Contains(*buffer, search) {
			return nil
		}
	}
	return fmt.Errorf("not found")
}
