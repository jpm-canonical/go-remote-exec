package remote_exec

type InstallType string

const (
	Deb  InstallType = "deb"
	Snap InstallType = "snap"
)

type SystemType string

const (
	Rpi5    SystemType = "rpi5"
	Generic SystemType = "generic"
)
