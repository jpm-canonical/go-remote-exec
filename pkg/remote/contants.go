package remote

const (
	// --step_threshold=1 --hwts_filter full --serverOnly 1 -m -f /snap/linuxptp/current/etc/default.cfg
	StepThreshold = "--step_threshold"
	HwtsFilter    = "--hwts_filter"
	ServerOnly    = "--serverOnly"
	ClientOnly    = "--clientOnly"
	ConfigFile    = "-f"
	Verbose       = "--verbose" // Log messages to stdout
	UseSyslog     = "--use_syslog"
	Version       = "-v"
	Interface     = "-i"

	Ptp4l     = "/usr/sbin/ptp4l"
	Ptp4lSnap = "linuxptp.ptp4l"
)

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
