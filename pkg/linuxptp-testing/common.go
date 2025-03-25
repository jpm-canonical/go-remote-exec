package linuxptp_testing

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	remote "go-remote-exec/pkg/remote-exec"
	"golang.org/x/crypto/ssh"
)

func RunTest(t *testing.T, testSetup TestSetup) {
	serverTag := "server"
	clientTag := "client"

	// Connect to remote devices
	server := remote.Connect(t, serverTag, testSetup.Server.Hostname, testSetup.Server.Username, testSetup.Server.Password)
	client := remote.Connect(t, clientTag, testSetup.Client.Hostname, testSetup.Client.Username, testSetup.Client.Password)

	if testSetup.AddUnicastTable {
		serverIp := findIpAddress(t, serverTag, testSetup.Server, server)
		clientIp := findIpAddress(t, clientTag, testSetup.Client, client)

		// For unicast comms, client gets server IP, server gets client IP
		testSetup.Client.ConfigFile = appendUnicast(t, testSetup.Client.ConfigFile, testSetup.Client.Interface, serverIp)
		testSetup.Server.ConfigFile = appendUnicast(t, testSetup.Server.ConfigFile, testSetup.Server.Interface, clientIp)
	}

	startServer(t, serverTag, testSetup.Server, server)
	startClient(t, clientTag, testSetup.Client, client)
}

func startClient(t *testing.T, tag string, config HostSetup, client *ssh.Client) {
	t.Logf("%s | Copying config", tag)
	serverConfigPath := putConfigFile(t, tag, config, client)

	// Build command
	clientCommand := buildCommand(t, tag, config, serverConfigPath)
	clientCommand = append(clientCommand, configureSecurityAssociation(t, tag, config, client)...)
	clientCommand = append(clientCommand, ClientOnly, "1")

	t.Logf("%s | Starting", tag)
	var clientStdOut string
	var clientStdErr string
	runningPtr := remote.ExecuteAsync(t, tag, client, clientCommand, &clientStdOut, &clientStdErr)

	found := remote.WaitFor(runningPtr, &clientStdOut, config.StartedSubstring, 20*time.Second)
	if !found {
		t.Logf("%s STDOUT | %s", tag, clientStdOut)
		t.Logf("%s STDERR | %s", tag, clientStdErr)
		t.Fatalf("%s | Startup failed", tag)
	}
	t.Logf("%s | Started", tag)

	// Watch client logs for synchronisation with server
	clientSynchronising := false
	clientSyncBelowThreshold := false
	syncRepeats := 0

	clientStdOutCopy := ""
	period := 30 * time.Second
	endTime := time.Now().Add(period)
	t.Logf("%s | Waiting for sync, until %s", tag, endTime)

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
								t.Logf("%s | Synchronised. Master Offset %snS", tag, fields[3])
								clientSyncBelowThreshold = true
								break
							}
						} else {
							syncRepeats = 0
						}
					}
				} else {
					t.Logf("%s | Synchronising. Master Offset %snS", tag, fields[3])
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
								t.Logf("%s | Synchronised. RMS Offset %snS", tag, fields[2])
								clientSyncBelowThreshold = true
								break
							}
						} else {
							syncRepeats = 0
						}
					}
				} else {
					t.Logf("%s | Synchronising. RMS Offset %snS", tag, fields[2])
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
		t.Fatalf("%s | Synchronisation failed!", tag)
	}
}

func startServer(t *testing.T, tag string, config HostSetup, server *ssh.Client) {
	t.Logf("%s | Copying config", tag)
	serverConfigPath := putConfigFile(t, tag, config, server)

	// Build command
	serverCommand := buildCommand(t, tag, config, serverConfigPath)
	serverCommand = append(serverCommand, configureSecurityAssociation(t, tag, config, server)...)
	serverCommand = append(serverCommand, ServerOnly, "1")

	t.Logf("%s | Starting", tag)
	var serverStdOut string
	var serverStdErr string
	runningPtr := remote.ExecuteAsync(t, tag, server, serverCommand, &serverStdOut, &serverStdErr)

	found := remote.WaitFor(runningPtr, &serverStdOut, config.StartedSubstring, 20*time.Second)
	if !found {
		t.Fatalf("%s | Startup failed", tag)
	}
	t.Logf("%s | Started", tag)
}
