package remote

import (
	"bufio"
	"fmt"
	"io"
	"log"
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

func StartServer(host *ssh.Client) error {
	//command := `sudo linuxptp.ptp4l -i eth0 --step_threshold=1 --hwts_filter full --serverOnly 1 -m -f /snap/linuxptp/current/etc/default.cfg`
	//executeAsync(host, command)
	return nil
	// Need to return a handle to this process, to stop it later
}

func StartClient(host *ssh.Client) error {
	// sudo linuxptp.ptp4l -i eth0 --step_threshold=1 --hwts_filter full --clientOnly 1 -m -f /snap/linuxptp/current/etc/default.cfg
	return nil
}

func execute(host *ssh.Client, command string) (string, string, error) {

	log.Printf("[exec-ssh] %s", command)

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

	var stderrBuffer []byte
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

// Monitor stderr for the sudo password request, and only pipe it in when it is requested
// https://stackoverflow.com/a/44501303
func enterSudoPassword(in io.WriteCloser, out io.Reader, output *[]byte) {
	var (
		line string
		r    = bufio.NewReader(out)
	)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		*output = append(*output, b)

		if b == byte('\n') {
			line = ""
			continue
		}

		line += string(b)

		if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
			log.Printf("Remote requested sudo passwword. Entering.")
			_, err = in.Write([]byte(os.Getenv("REMOTE_PASSWORD") + "\n"))
			if err != nil {
				break
			}

			// Append newline to stderr, as that is how it looks in a teminal normally
			*output = append(*output, '\n')
		}
	}
}
