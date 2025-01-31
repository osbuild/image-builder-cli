package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/osbuild/images/data/repositories"
	"github.com/osbuild/images/pkg/reporegistry"
	"github.com/osbuild/images/pkg/rpmmd"
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

var newRepoRegistry = func(dataDir, extraRepoPath string) (*reporegistry.RepoRegistry, error) {
	var dataDirs []string
	if dataDir != "" {
		dataDirs = []string{dataDir}
	} else {
		dataDirs = defaultDataDirs
	}

	// XXX: think about sharing this with reporegistry?
	var fses []fs.FS
	for _, d := range dataDirs {
		fses = append(fses, os.DirFS(filepath.Join(d, "repositories")))
	}
	fses = append(fses, repos.FS)

	// XXX: should we support disabling the build-ins somehow?
	conf, err := reporegistry.LoadAllRepositoriesFromFS(fses)
	if err != nil {
		return nil, err
	}

	// XXX: this should probably go into manifestgen.Options as
	// a new Options.ExtraRepoConf eventually (just like OverrideRepos)
	if extraRepoPath != "" {
		// XXX: this loads the extra repo unconditionally to all
		// distro versions. good luck with that!
		// Jokes aside, its unclear if we can do better without
		// burdens like forcing the user to name the file distro
		// after the $distro-$version.json
		// XXX2: should we just support yum repo formats here and
		// just internally convert to our json format?
		extraRepo, err := rpmmd.LoadRepositoriesFromFile(extraRepoPath)
		if err != nil {
			return nil, err
		}
		for _, repoArchConfigs := range conf {
			for arch := range repoArchConfigs {
				archCfg := repoArchConfigs[arch]
				archCfg = append(archCfg, extraRepo[arch]...)
				repoArchConfigs[arch] = archCfg
			}
		}
	}

	return reporegistry.NewFromDistrosRepoConfigs(conf), nil
}
