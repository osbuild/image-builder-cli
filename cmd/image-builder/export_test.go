package main

import (
	"fmt"
	"io"
	"os"

	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/distrofactory"
	"github.com/osbuild/images/pkg/reporegistry"
	"github.com/osbuild/images/pkg/rpmmd"
)

var (
	GetOneImage = getOneImage
	Run         = run
)

type ManifestGenerator = manifestGenerator

func MockOsArgs(new []string) (restore func()) {
	saved := os.Args
	os.Args = append([]string{"argv0"}, new...)
	return func() {
		os.Args = saved
	}
}

func MockOsStdout(new io.Writer) (restore func()) {
	saved := osStdout
	osStdout = new
	return func() {
		osStdout = saved
	}
}

func MockOsStderr(new io.Writer) (restore func()) {
	saved := osStderr
	osStderr = new
	return func() {
		osStderr = saved
	}
}

func MockNewRepoRegistry(f func() (*reporegistry.RepoRegistry, error)) (restore func()) {
	saved := newRepoRegistry
	newRepoRegistry = func(dataDir string) (*reporegistry.RepoRegistry, error) {
		if dataDir != "" {
			panic(fmt.Sprintf("cannot use custom dataDir %v in mock", dataDir))
		}
		return f()
	}
	return func() {
		newRepoRegistry = saved
	}
}

func MockDistrofactoryNew(f func() *distrofactory.Factory) (restore func()) {
	saved := distrofactoryNew
	distrofactoryNew = f
	return func() {
		distrofactoryNew = saved
	}
}

func MockDepsolve() (restore func()) {
	saved := depsolve
	// XXX: move to images
	depsolve = func(cacheDir string, packageSets map[string][]rpmmd.PackageSet, d distro.Distro, arch string) (map[string][]rpmmd.PackageSpec, map[string][]rpmmd.RepoConfig, error) {
		depsolvedSets := make(map[string][]rpmmd.PackageSpec)
		repoSets := make(map[string][]rpmmd.RepoConfig)
		for name, pkgSet := range packageSets {
			depsolvedSets[name] = []rpmmd.PackageSpec{
				{
					Name:     pkgSet[0].Include[0],
					Checksum: "sha256:01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b",
				},
			}
			//repoSets[name] = res.Repos
		}
		return depsolvedSets, repoSets, nil
	}
	return func() {
		depsolve = saved
	}
}
