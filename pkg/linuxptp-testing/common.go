package linuxptp_testing

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	remote "go-remote-exec/pkg/remote-exec"
	"golang.org/x/crypto/ssh"
)

func InstallSnap(t *testing.T, tag string, host *ssh.Client) {
	command := []string{"sudo", "snap", "install", "linuxptp", "--beta"}
	stdout, stderr := remote.Execute(t, tag, host, command)
	fmt.Printf("===stdout===\n%s============\n", stdout)
	fmt.Printf("===stderr===\n%s============\n", stderr)
}

func RunTest(t *testing.T, testSetup TestSetup) {
	startServer(t, testSetup.Server)
	startClient(t, testSetup.Client)
}

func startClient(t *testing.T, config HostSetup) {
	tag := "client"

	client := remote.Connect(t, tag, config.Hostname, config.Username, config.Password)

	t.Log("# Copying config to client")
	clientConfigPath := GetConfigDirectory(t, config.InstallType) + fmt.Sprintf("%s-%d.cfg", t.Name(), time.Now().Unix())
	remote.CopyFile(t, tag, config.ConfigFile, clientConfigPath, client)

	application := Ptp4l
	if config.InstallType == remote.Snap {
		application = Ptp4lSnap
	}
	clientCommand := []string{
		"sudo", application,
		Interface, config.Interface,
		Verbose, "1",
		UseSyslog, "0",
		StepThreshold, "1", // include this to allow quicker syncs by stepping clock
		ConfigFile, clientConfigPath,
	}
	clientCommand = append(clientCommand, ClientOnly, "1")

	// If Security Association is set, copy its config file, and add the argument
	if config.SecurityAssociationFile != "" {
		secAssocPath := GetConfigDirectory(t, config.InstallType) + fmt.Sprintf("%s-sa-%d.cfg", t.Name(), time.Now().Unix())
		remote.CopyFile(t, tag, config.SecurityAssociationFile, secAssocPath, client)
		clientCommand = append(clientCommand, SaFile, secAssocPath)
	}

	// Append Rpi5 specific arguments
	if config.SystemType == remote.Rpi5 {
		clientCommand = append(clientCommand, Ptp4lRpi5Specific...)
	}

	// Append Snap specific arguments
	if config.InstallType == remote.Snap {
		clientCommand = append(clientCommand, Ptp4lSnapSpecific...)

		// Also make sure the directory exists
		CreatePtp4lSnapUds(t, client)
	}

	t.Log("# Starting client")
	var clientStdOut string
	var clientStdErr string
	runningPtr := remote.ExecuteAsync(t, tag, client, clientCommand, &clientStdOut, &clientStdErr)

	found := remote.WaitFor(runningPtr, &clientStdOut, config.StartedSubstring, 20*time.Second)
	if !found {
		t.Logf("%s STDOUT | %s", tag, clientStdOut)
		t.Logf("%s STDERR | %s", tag, clientStdErr)
		t.Fatal("# Starting client failed")
	}
	t.Log("# Client started")

	// Watch client logs for synchronisation with server
	clientSynchronising := false
	clientSyncBelowThreshold := false
	syncRepeats := 0

	clientStdOutCopy := ""
	period := 30 * time.Second
	endTime := time.Now().Add(period)
	t.Logf("# Waiting for sync, until %s", endTime)

	// Monitor Client's stdout, split into lines, and check for sync message
	for time.Now().Before(endTime) {
		before, after, completeLine := strings.Cut(clientStdOut, "\n")
		if completeLine {
			clientStdOut = after
			clientStdOutCopy += before + "\n" // make a copy for printing later

			fields := strings.Fields(before)

			/*
				ptp4l[247021.552]: master offset      -1968 s0 freq   +7050 path delay     16919
			*/
			if fields[1] == "master" && fields[2] == "offset" {
				if config.RequireSyncBelowThreshold {
					masterOffset, err := strconv.Atoi(fields[3])
					if err != nil {
						t.Log(err)
					} else {
						if math.Abs(float64(masterOffset)) < PtpSyncThreshold {
							syncRepeats++

							if syncRepeats >= PtpSyncRepeats {
								t.Logf("# Client synchronised. Master Offset %snS", fields[3])
								clientSyncBelowThreshold = true
								break
							}
						} else {
							syncRepeats = 0
						}
					}
				} else {
					t.Logf("# Client synchronising. Master Offset %snS", fields[3])
					clientSynchronising = true
					break
				}
			}

			/*
				ptp4l[246963.314]: rms    8 max   15 freq  +7051 +/-  12
				ptp4l[246964.315]: rms    9 max   18 freq  +7052 +/-  12 delay 16909 +/-   0
			*/
			if fields[1] == "rms" {
				if config.RequireSyncBelowThreshold {
					rmsOffset, err := strconv.Atoi(fields[2])
					if err != nil {
						t.Log(err)
					} else {
						if math.Abs(float64(rmsOffset)) < PtpSyncThreshold {
							syncRepeats++

							if syncRepeats >= PtpSyncRepeats {
								t.Logf("# Client synchronised. RMS Offset %snS", fields[2])
								clientSyncBelowThreshold = true
								break
							}
						} else {
							syncRepeats = 0
						}
					}
				} else {
					t.Logf("# Client synchronising. RMS Offset %snS", fields[2])
					clientSynchronising = true
					break
				}
			}
		}
	}

	if (config.RequireSyncBelowThreshold && !clientSyncBelowThreshold) ||
		(!config.RequireSyncBelowThreshold && !clientSynchronising) {
		t.Log(clientStdOutCopy)
		t.Log(clientStdErr)
		t.Fatal("# Synchronisation failed!")
	}
}

func startServer(t *testing.T, config HostSetup) {
	tag := "server"

	// Connect to two remote devices
	server := remote.Connect(t, tag, config.Hostname, config.Username, config.Password)

	t.Log("# Copying config to server")
	// Use a unique name for the test config file
	serverConfigPath := GetConfigDirectory(t, config.InstallType) + fmt.Sprintf("%s-%d.cfg", t.Name(), time.Now().Unix())
	// Copy config file to both machines
	remote.CopyFile(t, tag, config.ConfigFile, serverConfigPath, server)

	// Build command
	application := Ptp4l
	if config.InstallType == remote.Snap {
		application = Ptp4lSnap
	}
	serverCommand := []string{
		"sudo", application,
		Interface, config.Interface,
		Verbose, "1",
		UseSyslog, "0",
		ConfigFile, serverConfigPath,
	}
	serverCommand = append(serverCommand, ServerOnly, "1")

	// If Security Association is set, copy its config file, and add the argument
	if config.SecurityAssociationFile != "" {
		secAssocPath := GetConfigDirectory(t, config.InstallType) + fmt.Sprintf("%s-sa-%d.cfg", t.Name(), time.Now().Unix())
		remote.CopyFile(t, tag, config.SecurityAssociationFile, secAssocPath, server)
		serverCommand = append(serverCommand, SaFile, secAssocPath)
	}

	// Append Rpi5 specific arguments
	if config.SystemType == remote.Rpi5 {
		serverCommand = append(serverCommand, Ptp4lRpi5Specific...)
	}

	// Append Snap specific arguments
	if config.InstallType == remote.Snap {
		serverCommand = append(serverCommand, Ptp4lSnapSpecific...)

		// Also make sure the directory exists
		CreatePtp4lSnapUds(t, server)
	}

	t.Log("# Starting server")
	var serverStdOut string
	var serverStdErr string
	runningPtr := remote.ExecuteAsync(t, tag, server, serverCommand, &serverStdOut, &serverStdErr)

	found := remote.WaitFor(runningPtr, &serverStdOut, config.StartedSubstring, 20*time.Second)
	if !found {
		t.Fatal("# Starting server failed")
	}
	t.Log("# Server started")
}
