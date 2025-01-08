package main

import (
	"github.com/osbuild/images/pkg/reporegistry"
)

// defaultDataDirs contains the default search paths to look for repository
// data. Note that the repositories are under a repositories/ sub-directory
// and contain a bunch of json files of the form "$distro_$version".json
// (but that is an implementation detail that the "images" library takes
// care of).
var defaultDataDirs = []string{
	"/etc/image-builder",
	"/usr/share/image-builder",
}

var newRepoRegistry = func(dataDir string) (*reporegistry.RepoRegistry, error) {
	var dataDirs []string
	if dataDir != "" {
		dataDirs = []string{dataDir}
	} else {
		dataDirs = defaultDataDirs
	}

	return reporegistry.New(dataDirs)
}
