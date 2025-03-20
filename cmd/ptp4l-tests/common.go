package ptp4l_tests

import (
	"fmt"
	"log"
	"math"
	"strconv"
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
	Hostname    string
	Username    string
	Password    string
	InstallType remote.InstallType
	SystemType  remote.SystemType
	Interface   string
	ConfigFile  string

	StartedSubstring          string
	RequireSyncBelowThreshold bool
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
		remote.Interface, testSetup.Client.Interface,
		remote.Verbose, "1",
		remote.UseSyslog, "0",
		remote.StepThreshold, "1", // include this to allow quicker syncs by stepping clock
		remote.ConfigFile, clientConfigPath,
	}
	clientCommand = append(clientCommand, remote.ClientOnly, "1")

	// Append Rpi5 specific arguments
	if testSetup.Client.SystemType == remote.Rpi5 {
		clientCommand = append(clientCommand, remote.Ptp4lRpi5Specific...)
	}

	// Append Snap specific arguments
	if testSetup.Client.InstallType == remote.Snap {
		clientCommand = append(clientCommand, remote.Ptp4lSnapSpecific...)

		// Also make sure the directory exists
		remote.CreatePtp4lSnapUds(t, client)
	}

	t.Log("# Starting client")
	var clientStdOut string
	var clientStdErr string
	runningPtr := remote.ExecuteAsync(t, client, clientCommand, &clientStdOut, &clientStdErr)

	found := remote.WaitFor(runningPtr, &clientStdOut, testSetup.Client.StartedSubstring, 20*time.Second)
	if !found {
		t.Log(clientStdOut)
		t.Log(clientStdErr)
		t.Fatal("# Starting client failed")
	}
	t.Log("# Client started")

	// Watch client logs for synchronisation with server
	clientSynchronised := false
	syncRepeats := 0

	clientStdOutCopy := ""
	period := 30 * time.Second
	endTime := time.Now().Add(period)
	log.Printf("# Waiting for sync, until %s", endTime)

	// Monitor Client's stdout, split into lines, and check for sync message
	for time.Now().Before(endTime) {
		before, after, completeLine := strings.Cut(clientStdOut, "\n")
		if completeLine {
			clientStdOut = after
			clientStdOutCopy += before // make a copy for printing later

			fields := strings.Fields(before)

			/*
				ptp4l[247021.552]: master offset      -1968 s0 freq   +7050 path delay     16919
			*/
			if fields[1] == "master" && fields[2] == "offset" {
				if testSetup.Client.RequireSyncBelowThreshold {
					masterOffset, err := strconv.Atoi(fields[3])
					if err != nil {
						t.Log(err)
					} else {
						if math.Abs(float64(masterOffset)) < remote.PtpSyncThreshold {
							syncRepeats++

							if syncRepeats >= remote.PtpSyncRepeats {
								t.Logf("# Client synchronised. Master Offset %snS", fields[3])
								clientSynchronised = true
								break
							}
						} else {
							syncRepeats = 0
						}
					}
				} else {
					t.Logf("# Client synchronising. Master Offset %snS", fields[3])
					clientSynchronised = true
					break
				}
			}

			/*
				ptp4l[246963.314]: rms    8 max   15 freq  +7051 +/-  12
				ptp4l[246964.315]: rms    9 max   18 freq  +7052 +/-  12 delay 16909 +/-   0
			*/
			if fields[1] == "rms" {
				if testSetup.Client.RequireSyncBelowThreshold {
					rmsOffset, err := strconv.Atoi(fields[2])
					if err != nil {
						t.Log(err)
					} else {
						if math.Abs(float64(rmsOffset)) < remote.PtpSyncThreshold {
							syncRepeats++

							if syncRepeats >= remote.PtpSyncRepeats {
								t.Logf("# Client synchronised. RMS Offset %snS", fields[2])
								clientSynchronised = true
								break
							}
						} else {
							syncRepeats = 0
						}
					}
				} else {
					t.Logf("# Client synchronising. RMS Offset %snS", fields[2])
					clientSynchronised = true
					break
				}
			}
		}
	}

	if !clientSynchronised {
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
		remote.Interface, testSetup.Server.Interface,
		remote.Verbose, "1",
		remote.UseSyslog, "0",
		remote.ConfigFile, serverConfigPath,
	}
	serverCommand = append(serverCommand, remote.ServerOnly, "1")

	// Append Rpi5 specific arguments
	if testSetup.Server.SystemType == remote.Rpi5 {
		serverCommand = append(serverCommand, remote.Ptp4lRpi5Specific...)
	}

	// Append Snap specific arguments
	if testSetup.Server.InstallType == remote.Snap {
		serverCommand = append(serverCommand, remote.Ptp4lSnapSpecific...)

		// Also make sure the directory exists
		remote.CreatePtp4lSnapUds(t, server)
	}

	t.Log("# Starting server")
	var serverStdOut string
	var serverStdErr string
	runningPtr := remote.ExecuteAsync(t, server, serverCommand, &serverStdOut, &serverStdErr)

	found := remote.WaitFor(runningPtr, &serverStdOut, testSetup.Server.StartedSubstring, 20*time.Second)
	if !found {
		t.Fatal("# Starting server failed")
	}
	t.Log("# Server started")
}
