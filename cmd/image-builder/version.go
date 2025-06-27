package main

import (
	"fmt"
	"runtime/debug"
)

// Get the version for this project based on the passed in string, the passed in string
// can potentially be set to something externally (for example through build arguments).
// When the string has any value other than the sentinel value "DEVEL" we pass it along
// as-is. Otherwise we try and get information from the build info.
func GetVersion(v string) string {
	// buildversion was set at build time, don't do any special handling but just
	// return it as-is
	if v != "DEVEL" {
		return v
	} else {
		// otherwise try to get it from the build information
		return versionFromBuildInfo()
	}
}

// Try and get a version (commit) from the build info included in the binary since Go 1.18.
func versionFromBuildInfo() string {
	version := "DEVEL"

	if nfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range nfo.Settings {
			if setting.Key == "vcs.revision" {
				version = fmt.Sprintf("DEVEL-%s", setting.Value)
			}
		}
	}

	return version
}
