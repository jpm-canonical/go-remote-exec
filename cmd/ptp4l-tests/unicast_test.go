package ptp4l_tests

import (
	"fmt"
	"os"
	"testing"

	linuxptp "go-remote-exec/pkg/linuxptp-testing"
	remote "go-remote-exec/pkg/remote-exec"
)

func createUnicastServerConfig(t *testing.T, transport linuxptp.Transport) string {

	// Use default server config but append transport
	config := `#
# Unicast master example configuration containing those attributes
# which differ from the defaults.  See the file, default.cfg, for the
# complete list of available options.
#
[global]
hybrid_e2e			1
inhibit_multicast_service	1
unicast_listen			1
`
	config = fmt.Sprintf("%s\nnetwork_transport\t\t%s\n", config, transport)

	// Create temp file for new config, and queue it to be deleted after test run
	newConfigFile, err := os.CreateTemp("", fmt.Sprintf("%s-server-*.cfg", t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(newConfigFile.Name())
	})
	// Write config
	_, err = newConfigFile.Write([]byte(config))
	if err != nil {
		t.Fatal(err)
	}
	err = newConfigFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return newConfigFile.Name()
}

func createUnicastClientConfig(t *testing.T, transport linuxptp.Transport) string {

	// Unicast slave config contains configs that are only for example purposes, and prevents the test from running.
	// We create a clean one to pass to the test.
	config := `#
# UNICAST slave example configuration with contrived master tables.
# This example will not work out of the box!
#
[global]
#
# Request service for sixty seconds.
#
unicast_req_duration	60
`
	config = fmt.Sprintf("%s\nnetwork_transport\t\t%s\n", config, transport)

	// Create temp file for new config, and queue it to be deleted after test run
	newConfigFile, err := os.CreateTemp("", fmt.Sprintf("%s-client-*.cfg", t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(newConfigFile.Name())
	})
	// Write config
	_, err = newConfigFile.Write([]byte(config))
	if err != nil {
		t.Fatal(err)
	}
	err = newConfigFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return newConfigFile.Name()
}

/*
TestUnicastUDPv4 sets up a client to synchronise from a server using UDP IPv4
*/
func TestUnicastUDPv4(t *testing.T) {
	transport := linuxptp.UDPv4

	remotePassword := os.Getenv("REMOTE_PASSWORD")
	if remotePassword == "" {
		t.Fatal("REMOTE_PASSWORD environment variable not set")
	}

	serverConfigFile := createUnicastServerConfig(t, transport)
	clientConfigFileName := createUnicastClientConfig(t, transport)

	testSetup := linuxptp.TestSetup{
		Server: linuxptp.HostSetup{
			Hostname:    "raspi-a.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  serverConfigFile,

			// The example does not add a unicast table for the master
			//AddUnicastTable:  true,
			//UnicastTransport: linuxptp.UDPv4,

			StartedSubstring: "assuming the grand master role",
		},
		Client: linuxptp.HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  clientConfigFileName,

			AddUnicastTable:  true,
			UnicastTransport: transport,

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}

	linuxptp.RunTest(t, testSetup)
}

/*
TestUnicastUDPv6 sets up a client to synchronise from a server using UDP IPv6
*/
func TestUnicastUDPv6(t *testing.T) {
	transport := linuxptp.UDPv6

	remotePassword := os.Getenv("REMOTE_PASSWORD")
	if remotePassword == "" {
		t.Fatal("REMOTE_PASSWORD environment variable not set")
	}

	serverConfigFile := createUnicastServerConfig(t, transport)
	clientConfigFileName := createUnicastClientConfig(t, transport)

	testSetup := linuxptp.TestSetup{
		Server: linuxptp.HostSetup{
			Hostname:    "raspi-a.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  serverConfigFile,

			// The example does not add a unicast table for the master
			//AddUnicastTable:  true,
			//UnicastTransport: linuxptp.UDPv4,

			StartedSubstring: "assuming the grand master role",
		},
		Client: linuxptp.HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  clientConfigFileName,

			AddUnicastTable:  true,
			UnicastTransport: transport,

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}

	linuxptp.RunTest(t, testSetup)
}

/*
TestUnicastL2 sets up a client to synchronise from a server using Layer 2 (MAC)
*/
func TestUnicastL2(t *testing.T) {
	transport := linuxptp.L2

	remotePassword := os.Getenv("REMOTE_PASSWORD")
	if remotePassword == "" {
		t.Fatal("REMOTE_PASSWORD environment variable not set")
	}

	serverConfigFile := createUnicastServerConfig(t, transport)
	clientConfigFileName := createUnicastClientConfig(t, transport)

	testSetup := linuxptp.TestSetup{
		Server: linuxptp.HostSetup{
			Hostname:    "raspi-a.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  serverConfigFile,

			// The example does not add a unicast table for the master
			//AddUnicastTable:  true,
			//UnicastTransport: linuxptp.UDPv4,

			StartedSubstring: "assuming the grand master role",
		},
		Client: linuxptp.HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",
			ConfigFile:  clientConfigFileName,

			AddUnicastTable:  true,
			UnicastTransport: transport,

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}

	linuxptp.RunTest(t, testSetup)
}
