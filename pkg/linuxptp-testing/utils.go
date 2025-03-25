package linuxptp_testing

import (
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	remote "go-remote-exec/pkg/remote-exec"
	"golang.org/x/crypto/ssh"
)

// getConfigDirectory returns the correct config directory depending on the installation type
func getConfigDirectory(t *testing.T, installType remote.InstallType) string {
	remotePath := ""
	if installType == remote.Deb {
		remotePath = "/etc/linuxptp/"
	} else if installType == remote.Snap {
		remotePath = "/var/snap/linuxptp/common/"
	}
	return remotePath
}

func buildCommand(t *testing.T, tag string, hostSetup HostSetup, remoteConfigPath string) []string {

	// Build command
	application := Ptp4l
	if hostSetup.InstallType == remote.Snap {
		application = Ptp4lSnap
	}
	command := []string{
		"sudo", application,
		Interface, hostSetup.Interface,
		Verbose, "1",
		UseSyslog, "0",
		ConfigFile, remoteConfigPath,
	}

	// Append Rpi5 specific arguments
	if hostSetup.SystemType == remote.Rpi5 {
		command = append(command, Ptp4lRpi5Specific...)
	}

	// Append Snap specific arguments
	if hostSetup.InstallType == remote.Snap {
		command = append(command, Ptp4lSnapSpecific...)
	}

	return command
}

func configureSecurityAssociation(t *testing.T, tag string, hostSetup HostSetup, host *ssh.Client) []string {
	if hostSetup.SecurityAssociationFile == "" {
		return []string{}
	}

	secAssocPath := getConfigDirectory(t, hostSetup.InstallType) + fmt.Sprintf("%s-sa-%d.cfg", t.Name(), time.Now().Unix())
	remote.CopyFile(t, tag, hostSetup.SecurityAssociationFile, secAssocPath, host)
	return []string{SaFile, secAssocPath}
}

func putConfigFile(t *testing.T, tag string, config HostSetup, host *ssh.Client) string {
	// Use a unique name for the test config file
	remoteConfigPath := getConfigDirectory(t, config.InstallType) + fmt.Sprintf("%s-%d.cfg", t.Name(), time.Now().Unix())
	remote.CopyFile(t, tag, config.ConfigFile, remoteConfigPath, host)
	return remoteConfigPath
}

func addUnicastConfig(t *testing.T, tag string, testSetup TestSetup, server *ssh.Client, client *ssh.Client) TestSetup {
	if !testSetup.Server.AddUnicastTable && !testSetup.Client.AddUnicastTable {
		return testSetup
	}

	// For unicast comms, server gets client address, client gets server address

	if testSetup.Server.AddUnicastTable {
		switch testSetup.Server.UnicastTransport {
		case L2:
			clientMac, err := remote.GetMacAddress(t, tag, client, testSetup.Client.Interface)
			if err != nil {
				t.Fatalf("Can't find client MAC: %s", err)
			}
			testSetup.Server.ConfigFile = appendUnicastL2(t, testSetup.Server.ConfigFile, testSetup.Server.Interface, clientMac)
		case UDPv4:
			clientIPv4, err := remote.GetIPv4Address(t, tag, client, testSetup.Client.Interface)
			if err != nil {
				t.Fatalf("Can't find client IPv4: %s", err)
			}
			testSetup.Server.ConfigFile = appendUnicastUDPv4(t, testSetup.Server.ConfigFile, testSetup.Server.Interface, clientIPv4)
		case UDPv6:
			clientIPv6, err := remote.GetIPv6GlobalAddress(t, tag, client, testSetup.Client.Interface)
			if err != nil {
				t.Logf("Can't find client global IPv6: %s", err)
				// If no global IPv6 exists, fall back to link local address
				clientIPv6, err = remote.GetIPv6LocalAddress(t, tag, client, testSetup.Client.Interface)
				if err != nil {
					t.Fatalf("Can't find client link local IPv6: %s", err)
				}
			}
			testSetup.Server.ConfigFile = appendUnicastUDPv6(t, testSetup.Server.ConfigFile, testSetup.Server.Interface, clientIPv6)
		}

	}

	if testSetup.Client.AddUnicastTable {
		switch testSetup.Client.UnicastTransport {
		case L2:
			serverMac, err := remote.GetMacAddress(t, tag, server, testSetup.Server.Interface)
			if err != nil {
				t.Fatalf("Can't find server MAC: %s", err)
			}
			testSetup.Client.ConfigFile = appendUnicastL2(t, testSetup.Client.ConfigFile, testSetup.Client.Interface, serverMac)
		case UDPv4:
			serverIPv4, err := remote.GetIPv4Address(t, tag, server, testSetup.Server.Interface)
			if err != nil {
				t.Fatalf("Can't find server IPv4: %s", err)
			}
			testSetup.Client.ConfigFile = appendUnicastUDPv4(t, testSetup.Client.ConfigFile, testSetup.Client.Interface, serverIPv4)
		case UDPv6:
			serverIPv6, err := remote.GetIPv6GlobalAddress(t, tag, server, testSetup.Server.Interface)
			if err != nil {
				t.Logf("Can't find server global IPv6: %s", err)
				// If no global IPv6 exists, fall back to link local address
				serverIPv6, err = remote.GetIPv6LocalAddress(t, tag, server, testSetup.Server.Interface)
				if err != nil {
					t.Fatalf("Can't find server link local IPv6: %s", err)
				}
			}
			testSetup.Client.ConfigFile = appendUnicastUDPv6(t, testSetup.Client.ConfigFile, testSetup.Client.Interface, serverIPv6)
		}
	}

	// Return a copy of the test setup with the unicast configs added
	return testSetup
}

func appendUnicastL2(t *testing.T, originalConfigFilePath string, interfaceName string, mac string) string {

	configFormatString := `
[unicast_master_table]
table_id			1
logQueryInterval	2
L2					%[1]s

[%[2]s]
unicast_master_table	1
`

	unicastConfig := fmt.Sprintf(configFormatString, mac, interfaceName)
	return appendUnicastToConfigFile(t, originalConfigFilePath, unicastConfig)
}

func appendUnicastUDPv4(t *testing.T, originalConfigFilePath string, interfaceName string, remoteIp net.IP) string {

	configFormatString := `
[unicast_master_table]
table_id			1
logQueryInterval	2
UDPv4				%[1]s

[%[2]s]
unicast_master_table	1
`

	unicastConfig := fmt.Sprintf(configFormatString, remoteIp, interfaceName)
	return appendUnicastToConfigFile(t, originalConfigFilePath, unicastConfig)
}

func appendUnicastUDPv6(t *testing.T, originalConfigFilePath string, interfaceName string, remoteIp net.IP) string {

	configFormatString := `
[unicast_master_table]
table_id			1
logQueryInterval	2
UDPv6				%[1]s

[%[2]s]
unicast_master_table	1
`

	unicastConfig := fmt.Sprintf(configFormatString, remoteIp, interfaceName)
	return appendUnicastToConfigFile(t, originalConfigFilePath, unicastConfig)
}

/*
appendUnicastToConfigFile reads the original config file, appends the unicast config to it, and then writes it out to a temp file
The path to the temp file is returned.
*/
func appendUnicastToConfigFile(t *testing.T, originalConfigFilePath string, unicastConfig string) string {
	// Read original config file
	originalConfigFile, err := os.Open(originalConfigFilePath)
	if err != nil {
		t.Fatal(err)
	}
	originalConfig, err := io.ReadAll(originalConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	// Create temp file for new config, and queue it to be deleted after test run
	newConfigFile, err := os.CreateTemp("", fmt.Sprintf("%s-server-*.cfg", t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(newConfigFile.Name())
	})

	// First write original config
	_, err = newConfigFile.Write(originalConfig)
	if err != nil {
		t.Fatal(err)
	}
	// Then append unicast config
	_, err = newConfigFile.Write([]byte(unicastConfig))
	if err != nil {
		t.Fatal(err)
	}

	err = newConfigFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	return newConfigFile.Name()
}
