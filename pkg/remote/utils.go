package remote

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func CopyFile(t *testing.T, localFileName, remoteFileName string, host *ssh.Client) {
	localFile, err := os.Open(localFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer localFile.Close()

	sftpClient, err := sftp.NewClient(host)
	if err != nil {
		t.Fatal(err)
	}
	defer sftpClient.Close()

	remoteFile, err := sftpClient.Create(remoteFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer remoteFile.Close()

	err = sftpClient.Chmod(remoteFileName, 0777)
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		t.Fatal(err)
	}
}

func WaitFor(buffer *string, search string, timeout time.Duration) bool {
	start := time.Now()

	for time.Now().Before(start.Add(timeout)) {
		if strings.Contains(*buffer, search) {
			return true
		}
	}
	return false
}
