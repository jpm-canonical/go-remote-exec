package ptp4l_tests

import (
	"os"
	"testing"

	"go-remote-exec/pkg/remote"
)

func TestGptp(t *testing.T) {
	testSetup := TestSetup{
		Server: HostSetup{
			Hostname:    "raspi-a.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.2/gPTP.cfg",

			StartedSubstring: "selected local clock 2ccf67.fffe.1cbba1 as best master",
		},
		Client: HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  "../../default-configs-4.2/gPTP.cfg",

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}

	runTest(t, testSetup)
}
