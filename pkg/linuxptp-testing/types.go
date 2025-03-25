package linuxptp_testing

import remote "go-remote-exec/pkg/remote-exec"

type TestSetup struct {
	Server HostSetup
	Client HostSetup
}

type HostSetup struct {
	Hostname                string
	Username                string
	Password                string
	InstallType             remote.InstallType
	SystemType              remote.SystemType
	Interface               string
	ConfigFile              string
	SecurityAssociationFile string

	AddUnicastTable  bool // default is false
	UnicastTransport Transport

	StartedSubstring          string
	RequireSyncBelowThreshold bool
}

type Transport string

const (
	L2    Transport = "L2"
	UDPv4 Transport = "UDPv4"
	UDPv6 Transport = "UDPv6"
)
