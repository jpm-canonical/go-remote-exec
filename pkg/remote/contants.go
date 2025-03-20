package remote

const (
	PtpSyncThreshold = 100

	// LinuxPTP arguments
	StepThreshold                     = "--step_threshold"
	HwtsFilter                        = "--hwts_filter"
	ServerOnly                        = "--serverOnly"
	ClientOnly                        = "--clientOnly"
	ConfigFile                        = "-f"
	Verbose                           = "--verbose" // Log messages to stdout
	UseSyslog                         = "--use_syslog"
	Version                           = "-v"
	Interface                         = "-i"
	NeighborPropagationDelayThreshold = "--neighborPropDelayThresh"

	// LinuxPTP applications
	Ptp4l     = "/usr/sbin/ptp4l"
	Ptp4lSnap = "linuxptp.ptp4l"
)

/*
Rpi5Specific Raspberry Pi 5 needs overrides to work past some limitations of the hardware.
- Hardware Filters need to be set to full. The driver does not properly timestamp only certain packet types, so we enable time stamping on everything.
- The standard propagation delay between two RPi5's, connected directly is 16000ns. This is likely due to the RP1 chip that offloads IO, but adds latency.
*/
var Rpi5Specific = []string{HwtsFilter, "full", NeighborPropagationDelayThreshold, "17000"}

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
