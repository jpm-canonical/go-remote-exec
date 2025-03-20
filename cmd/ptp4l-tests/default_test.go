package ptp4l_tests

import (
	"os"
	"testing"

	"go-remote-exec/pkg/remote"
)

func TestDefault(t *testing.T) {
	testSetup := TestSetup{
		Server: HostSetup{
			Hostname:    "raspi-a.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.4/default.cfg",

			StartedSubstring: "assuming the grand master role",
		},
		Client: HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.4/default.cfg",

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}
	runTest(t, testSetup)
}
