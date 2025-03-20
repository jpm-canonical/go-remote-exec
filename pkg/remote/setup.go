package remote

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func InstallSnap(t *testing.T, host *ssh.Client) {
	// install linuxptp
	command := []string{"sudo", "snap", "install", "linuxptp", "--beta"}
	stdout, stderr := Execute(t, host, command)
	fmt.Printf("===stdout===\n%s============\n", stdout)
	fmt.Printf("===stderr===\n%s============\n", stderr)
}

// CopyFile copies a file from the local machine to the remote host.
// It is first copies to the home directory, then moves it to the final path using sudo.
// The test cleanup will remove the file.
func CopyFile(t *testing.T, localFilePath, remoteFilePath string, host *ssh.Client) {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer localFile.Close()

	sftpClient, err := sftp.NewClient(host)
	if err != nil {
		t.Fatal(err)
	}
	defer sftpClient.Close()

	// Create and copy file to home directory first
	tempHomeDirFileName := filepath.Base(remoteFilePath)
	remoteFile, err := sftpClient.Create(tempHomeDirFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		t.Fatal(err)
	}

	// Move from home dir to proper conf directory using sudo
	command := []string{"sudo", "mv", tempHomeDirFileName, remoteFilePath}
	_, _ = Execute(t, host, command)

	t.Cleanup(func() {
		command := []string{"sudo", "rm", remoteFilePath}
		_, _ = Execute(t, host, command)
	})
}
