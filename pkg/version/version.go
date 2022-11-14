package version

import (
	"fmt"
	"runtime"
)

var (
	commit   string
	version  = "dev"
	platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	outputTemplate = `
Version: %s
Commit: %s
Platform: %s
Go: %s
`
)

// Version contains all of the information about the build of the current release.
// It is expected that this is ran on tagged releases and populated by `-ldflags` of the `go build` commmand.
// See the .gorelease.yml ldflags section for details.
type Version struct {
	Commit    string
	GoVersion string
	Platform  string
	Version   string
}

// ToString populates the version/build information into the output template and returns it for use.
func (v *Version) ToString() string {
	return fmt.Sprintf(outputTemplate, v.Version, v.Commit, v.Platform, v.GoVersion)
}

// GetVersion is used to retrieve the current version information of the application.
func GetVersion() *Version {
	return &Version{
		Commit:    commit,
		GoVersion: runtime.Version(),
		Platform:  platform,
		Version:   version,
	}
}
