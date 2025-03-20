package ptp4l_tests

import (
	"os"
	"testing"

	"go-remote-exec/pkg/remote"
)

func TestAutomotive(t *testing.T) {
	remotePassword := os.Getenv("REMOTE_PASSWORD")
	if remotePassword == "" {
		t.Fatal("REMOTE_PASSWORD environment variable not set")
	}

	testSetup := TestSetup{
		Server: HostSetup{
			Hostname:    "raspi-a.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.4/automotive-master.cfg",

			StartedSubstring: "INITIALIZING to MASTER on INIT_COMPLETE",
		},
		Client: HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.4/automotive-slave.cfg",

			StartedSubstring:          "INITIALIZING to SLAVE on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}
	runTest(t, testSetup)
}
