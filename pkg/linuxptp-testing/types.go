package linuxptp_testing

import remote "go-remote-exec/pkg/remote-exec"

type TestSetup struct {
	Server          HostSetup
	Client          HostSetup
	AddUnicastTable bool // default is false
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

	StartedSubstring          string
	RequireSyncBelowThreshold bool
}
