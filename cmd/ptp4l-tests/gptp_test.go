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

type TestSetup struct {
	Server HostSetup
	Client HostSetup
}

type HostSetup struct {
	Hostname         string
	Username         string
	Password         string
	InstallType      remote.InstallType
	SystemType       remote.SystemType
	ConfigFile       string
	StartedSubstring string
}

func TestGptp(t *testing.T) {
	testSetup := TestSetup{
		Server: HostSetup{
			Hostname:         "raspi-a.lan",
			Username:         "jpmeijers",
			Password:         os.Getenv("REMOTE_PASSWORD"),
			InstallType:      remote.Snap,
			SystemType:       remote.Rpi5,
			ConfigFile:       "../../default-configs-4.2/gPTP.cfg",
			StartedSubstring: "selected local clock 2ccf67.fffe.1cbba1 as best master",
		},
		Client: HostSetup{
			Hostname:         "raspi-b.lan",
			Username:         "jpmeijers",
			Password:         os.Getenv("REMOTE_PASSWORD"),
			InstallType:      remote.Snap,
			SystemType:       remote.Rpi5,
			ConfigFile:       "../../default-configs-4.2/gPTP.cfg",
			StartedSubstring: "INITIALIZING to LISTENING on INIT_COMPLETE",
		},
	}

	runTest(t, testSetup)
}

func TestDefault(t *testing.T) {
	testSetup := TestSetup{
		Server: HostSetup{
			Hostname:         "raspi-a.lan",
			Username:         "jpmeijers",
			Password:         os.Getenv("REMOTE_PASSWORD"),
			InstallType:      remote.Snap,
			SystemType:       remote.Rpi5,
			ConfigFile:       "../../default-configs-4.4/default.cfg",
			StartedSubstring: "assuming the grand master role",
		},
		Client: HostSetup{
			Hostname:         "raspi-b.lan",
			Username:         "jpmeijers",
			Password:         os.Getenv("REMOTE_PASSWORD"),
			InstallType:      remote.Snap,
			SystemType:       remote.Rpi5,
			ConfigFile:       "../../default-configs-4.4/default.cfg",
			StartedSubstring: "INITIALIZING to LISTENING on INIT_COMPLETE",
		},
	}
	runTest(t, testSetup)
}

func runTest(t *testing.T, testSetup TestSetup) {

	startServer(t, testSetup)

	client := remote.Connect(t, testSetup.Client.Hostname, testSetup.Client.Username, testSetup.Client.Password)
	clientConfigPath := remote.GetConfigDirectory(t, testSetup.Client.InstallType) + fmt.Sprintf("%s-%d.cfg", t.Name(), time.Now().Unix())
	remote.CopyFile(t, testSetup.Client.ConfigFile, clientConfigPath, client)

	application := remote.Ptp4l
	if testSetup.Client.InstallType == remote.Snap {
		application = remote.Ptp4lSnap
	}
	clientCommand := []string{
		"sudo", application,
		remote.Interface, "eth0",
		remote.Verbose, "1",
		remote.UseSyslog, "0",
		remote.ConfigFile, clientConfigPath,
	}
	clientCommand = append(clientCommand, remote.ClientOnly, "1")

	// Append Rpi5 specific arguments
	if testSetup.Client.SystemType == remote.Rpi5 {
		clientCommand = append(clientCommand, remote.HwtsFilter, "full", remote.StepThreshold, "1")
	}

	t.Log("# Starting client")
	var clientStdOut string
	var clientStdErr string
	remote.ExecuteAsync(t, client, clientCommand, &clientStdOut, &clientStdErr)

	found := remote.WaitFor(&clientStdOut, "INITIALIZING to LISTENING on INIT_COMPLETE", 20*time.Second)
	if !found {
		t.Log(clientStdOut)
		t.Log(clientStdErr)
		t.Fatal("# Starting client failed")
	}
	t.Log("# Client started")

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

func startServer(t *testing.T, testSetup TestSetup) {

	// Connect to two remote devices
	server := remote.Connect(t, testSetup.Server.Hostname, testSetup.Server.Username, testSetup.Server.Password)
	// Use a unique name for the test config file
	serverConfigPath := remote.GetConfigDirectory(t, testSetup.Server.InstallType) + fmt.Sprintf("%s-%d.cfg", t.Name(), time.Now().Unix())
	// Copy config file to both machines
	remote.CopyFile(t, testSetup.Server.ConfigFile, serverConfigPath, server)

	application := remote.Ptp4l
	if testSetup.Server.InstallType == remote.Snap {
		application = remote.Ptp4lSnap
	}
	serverCommand := []string{
		"sudo", application,
		remote.Interface, "eth0",
		remote.Verbose, "1",
		remote.UseSyslog, "0",
		remote.ConfigFile, serverConfigPath,
	}
	serverCommand = append(serverCommand, remote.ServerOnly, "1")

	// Append Rpi5 specific arguments
	if testSetup.Server.SystemType == remote.Rpi5 {
		serverCommand = append(serverCommand, remote.HwtsFilter, "full", remote.StepThreshold, "1")
	}

	t.Log("# Starting server")
	var serverStdOut string
	var serverStdErr string
	remote.ExecuteAsync(t, server, serverCommand, &serverStdOut, &serverStdErr)

	found := remote.WaitFor(&serverStdOut, "assuming the grand master role", 20*time.Second)
	if !found {
		t.Log(serverStdOut)
		t.Log(serverStdErr)
		t.Fatal("# Starting server failed")
	}
	t.Log("# Server started")
}
