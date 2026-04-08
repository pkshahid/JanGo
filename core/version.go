package core

import "fmt"

// VERSION is the string representation of the current GoDjango framework version.
const VERSION = "1.0.0"

// VersionInfo represents detailed version information.
type VersionInfo struct {
	Major      int
	Minor      int
	Micro      int
	ReleaseLevel string
	Serial     int
}

// String returns the formatted version string.
func (v VersionInfo) String() string {
	if v.ReleaseLevel == "final" && v.Serial == 0 {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Micro)
	}
	return fmt.Sprintf("%d.%d.%d-%s%d", v.Major, v.Minor, v.Micro, v.ReleaseLevel, v.Serial)
}

// VERSION_INFO is the parsed version of the current release.
var VERSION_INFO = VersionInfo{
	Major:        1,
	Minor:        0,
	Micro:        0,
	ReleaseLevel: "final",
	Serial:       0,
}
