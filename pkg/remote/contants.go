package remote

const (
	PtpSyncThreshold = 100
	PtpSyncRepeats   = 5

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
	UdsAddress                        = "--uds_address"
	UdsRoAddress                      = "--uds_ro_address"

	// LinuxPTP applications
	Ptp4l     = "/usr/sbin/ptp4l"
	Ptp4lSnap = "linuxptp.ptp4l"

	// UDS paths
	Ptp4lUdsSnap   = "/run/snap.linuxptp/ptp4l"
	Ptp4lUdsRoSnap = "/run/snap.linuxptp/ptp4lro"
)

/*
Ptp4lRpi5Specific Raspberry Pi 5 needs overrides to work past some limitations of the hardware.
- Hardware Filters need to be set to full. The driver does not properly timestamp only certain packet types, so we enable time stamping on everything.
- The standard propagation delay between two RPi5's, connected directly is 16000ns. This is likely due to the RP1 chip that offloads IO, but adds latency.
*/
var Ptp4lRpi5Specific = []string{HwtsFilter, "full", NeighborPropagationDelayThreshold, "17000"}

/*
Ptp4lSnapSpecific contains ptp4l command line arguments for snaps.
- UDS paths. The snap needs to create Unix domain sockets under /run/snap.linuxptp/
*/
var Ptp4lSnapSpecific = []string{UdsAddress, Ptp4lUdsSnap, UdsRoAddress, Ptp4lUdsRoSnap}

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
