package ptp4l_tests

import (
	"os"
	"testing"

	linuxptp "go-remote-exec/pkg/linuxptp-testing"
	remote "go-remote-exec/pkg/remote-exec"
)

/*
TestG8265_1 runs a test using the G.8265.1 telecoms profile.
*/
func TestG8265_1(t *testing.T) {
	remotePassword := os.Getenv("REMOTE_PASSWORD")
	if remotePassword == "" {
		t.Fatal("REMOTE_PASSWORD environment variable not set")
	}

	testSetup := linuxptp.TestSetup{
		Server: linuxptp.HostSetup{
			Hostname:    "raspi-a.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.4/G.8265.1.cfg",

			AddUnicastTable:  true,
			UnicastTransport: linuxptp.UDPv4,

			StartedSubstring: "assuming the grand master role",
		},
		Client: linuxptp.HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.4/G.8265.1.cfg",

			AddUnicastTable:  true,
			UnicastTransport: linuxptp.UDPv4,

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}

	linuxptp.RunTest(t, testSetup)
}
