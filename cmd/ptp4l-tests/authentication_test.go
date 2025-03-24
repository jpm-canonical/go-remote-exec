package ptp4l_tests

import (
	"os"
	"testing"

	linuxptp "go-remote-exec/pkg/linuxptp-testing"
	remote "go-remote-exec/pkg/remote-exec"
)

/*
TestAuthentication does a time sync using authentication.
This needs LinuxPTP v4.4 or later, compiled with auth support (ex. gnutls).
*/
func TestAuthentication(t *testing.T) {
	remotePassword := os.Getenv("REMOTE_PASSWORD")
	if remotePassword == "" {
		t.Fatal("REMOTE_PASSWORD environment variable not set")
	}

	securityAssocFile := "../../default-configs-4.4/sa.cfg"

	testSetup := linuxptp.TestSetup{
		Server: linuxptp.HostSetup{
			Hostname:                "raspi-a.lan",
			Username:                "jpmeijers",
			Password:                os.Getenv("REMOTE_PASSWORD"),
			InstallType:             remote.Snap,
			SystemType:              remote.Rpi5,
			Interface:               "eth0",
			ConfigFile:              "../../default-configs-4.4/authentication.cfg",
			SecurityAssociationFile: securityAssocFile,

			StartedSubstring: "assuming the grand master role",
		},
		Client: linuxptp.HostSetup{
			Hostname:                "raspi-b.lan",
			Username:                "jpmeijers",
			Password:                os.Getenv("REMOTE_PASSWORD"),
			InstallType:             remote.Snap,
			SystemType:              remote.Rpi5,
			Interface:               "eth0",
			ConfigFile:              "../../default-configs-4.4/authentication.cfg",
			SecurityAssociationFile: securityAssocFile,

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}
	linuxptp.RunTest(t, testSetup)
}
