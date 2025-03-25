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

func findIpAddress(t *testing.T, tag string, hostSetup HostSetup, host *ssh.Client) net.IP {
	interfaceIp := remote.GetIpV4Address(t, tag, host, hostSetup.Interface)
	return interfaceIp
}

func appendUnicast(t *testing.T, baseConfigFile string, interfaceName string, remoteIp net.IP) string {
	genericConfigFile, err := os.Open(baseConfigFile)
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
