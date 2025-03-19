package ptp4l_tests

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"go-remote-exec/pkg/remote"
)

func TestGptp(t *testing.T) {
	localConfigFile := "../../default-configs/gPTP.cfg"
	testWithConfigFile(t, localConfigFile)
}

func TestDefault(t *testing.T) {
	localConfigFile := "../../default-configs/default.cfg"
	testWithConfigFile(t, localConfigFile)
}

func testWithConfigFile(t *testing.T, localConfigFile string) {
	remoteUser := "jpmeijers"
	remotePassword := os.Getenv("REMOTE_PASSWORD")

	remoteConfigFile := fmt.Sprintf("%s-%d.cfg", t.Name(), time.Now().Unix())

	// Connect to two remote devices
	hostA := remote.Connect(t, "raspi-a.lan", remoteUser, remotePassword)
	hostB := remote.Connect(t, "raspi-b.lan", remoteUser, remotePassword)

	// Copy config file to both machines
	remote.CopyFile(t, localConfigFile, remoteConfigFile, hostA)
	remote.CopyFile(t, localConfigFile, remoteConfigFile, hostB)

	commonCommand := []string{
		"sudo", remote.Ptp4l,
		remote.Interface, "eth0",
		remote.StepThreshold, "1",
		remote.HwtsFilter, "full",
		remote.Verbose, "1",
		remote.UseSyslog, "0",
		remote.ConfigFile, "$HOME/" + remoteConfigFile,
	}
	serverCommand := append(commonCommand, remote.ServerOnly, "1")
	clientCommand := append(commonCommand, remote.ClientOnly, "1")

	t.Log("# Starting server")
	var serverStdOut string
	var serverStdErr string
	remote.ExecuteAsync(t, hostA, serverCommand, &serverStdOut, &serverStdErr)

	found := remote.WaitFor(&serverStdOut, "assuming the grand master role", 20*time.Second)
	if !found {
		t.Log(serverStdOut)
		t.Log(serverStdErr)
		t.Fatal("# Starting server failed")
	}
	t.Log("# Server started")

	t.Log("# Starting client")
	var clientStdOut string
	var clientStdErr string
	remote.ExecuteAsync(t, hostB, clientCommand, &clientStdOut, &clientStdErr)

	found = remote.WaitFor(&clientStdOut, "INITIALIZING to LISTENING on INIT_COMPLETE", 20*time.Second)
	if !found {
		t.Log(clientStdOut)
		t.Log(clientStdErr)
		t.Fatal("# Starting server failed")
	}
	t.Log("# Server started")

	// Watch client logs for synchronisation with server
	foundSyncMessage := false
	clientStdOutCopy := ""
	period := 20 * time.Second
	endTime := time.Now().Add(period)
	log.Printf("# Waiting for sync, until %s", endTime)

	// Monitor Client's stdout, split into lines, and check for sync message
	for time.Now().Before(endTime) {
		before, after, completeLine := strings.Cut(clientStdOut, "\n")
		if completeLine {
			clientStdOut = after
			clientStdOutCopy += before // make a copy for printing later

			fields := strings.Fields(before)
			if fields[1] == "master" && fields[2] == "offset" {
				foundSyncMessage = true
				log.Printf("# Synchronising. Offset %snS", fields[3])
				break
			}
		}
	}

	if !foundSyncMessage {
		t.Log(clientStdOutCopy)
		t.Log(clientStdErr)
		t.Fatal("# Synchronisation failed!")
	}
}
