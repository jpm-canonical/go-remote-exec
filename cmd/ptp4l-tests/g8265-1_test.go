package ptp4l_tests

import (
	"fmt"
	"io"
	"net"
	"os"
	"testing"

	linuxptp "go-remote-exec/pkg/linuxptp-testing"
	remote "go-remote-exec/pkg/remote-exec"
)

/*
TestG8265_1 runs a test using the G.8265.1 telecoms profile.
This test does not currently work, as the config file needs to be edited to point to the unicast IP address of the remote host.
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

			StartedSubstring: "assuming the grand master role",
		},
		Client: linuxptp.HostSetup{
			Hostname:    "raspi-b.lan",
			Username:    "jpmeijers",
			Password:    os.Getenv("REMOTE_PASSWORD"),
			InstallType: remote.Snap,
			SystemType:  remote.Rpi5,
			Interface:   "eth0",

			StartedSubstring:          "INITIALIZING to LISTENING on INIT_COMPLETE",
			RequireSyncBelowThreshold: true,
		},
	}

	clientIp := findIpAddress(t, "client", testSetup.Client)
	serverIp := findIpAddress(t, "server", testSetup.Server)

	// Client gets server IP, server gets client IP
	testSetup.Client.ConfigFile = createConfigFile(t, testSetup.Client.Interface, serverIp)
	testSetup.Server.ConfigFile = createConfigFile(t, testSetup.Server.Interface, clientIp)

	linuxptp.RunTest(t, testSetup)
}

func findIpAddress(t *testing.T, tag string, hostSetup linuxptp.HostSetup) net.IP {
	host := remote.Connect(t, tag, hostSetup.Hostname, hostSetup.Username, hostSetup.Password)
	interfaceIp := remote.GetIpV4Address(t, tag, host, hostSetup.Interface)
	err := host.Close()
	if err != nil {
		t.Fatal(err)
	}
	return interfaceIp
}

func createConfigFile(t *testing.T, interfaceName string, remoteIp net.IP) string {
	genericConfigFile, err := os.Open("../../default-configs-4.4/G.8265.1.cfg")
	if err != nil {
		t.Fatal(err)
	}
	genericConfig, err := io.ReadAll(genericConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	configFormatString := `%[1]s
[unicast_master_table]
table_id			1
logQueryInterval		2
UDPv4				%[2]s

[%[3]s]
unicast_master_table		1
`

	hostConfig := fmt.Sprintf(configFormatString, genericConfig, remoteIp, interfaceName)
	hostConfigFile, err := os.CreateTemp("", fmt.Sprintf("%s-server-*.cfg", t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(hostConfigFile.Name())
	})
	_, err = hostConfigFile.Write([]byte(hostConfig))
	if err != nil {
		t.Fatal(err)
	}
	err = hostConfigFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return hostConfigFile.Name()
}
