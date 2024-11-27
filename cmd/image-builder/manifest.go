package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/dnfjson"
	"github.com/osbuild/images/pkg/rpmmd"
	"github.com/osbuild/images/pkg/sbom"
)

// XXX: duplicated from cmd/build/main.go:depsolve (and probably more places)
// should go into a common helper in "images" or images should do this on
// its own
func depsolve(cacheDir string, packageSets map[string][]rpmmd.PackageSet, d distro.Distro, arch string) (map[string][]rpmmd.PackageSpec, map[string][]rpmmd.RepoConfig, error) {
	solver := dnfjson.NewSolver(d.ModulePlatformID(), d.Releasever(), arch, d.Name(), cacheDir)
	depsolvedSets := make(map[string][]rpmmd.PackageSpec)
	repoSets := make(map[string][]rpmmd.RepoConfig)
	for name, pkgSet := range packageSets {
		res, err := solver.Depsolve(pkgSet, sbom.StandardTypeNone)
		if err != nil {
			return nil, nil, err
		}
		depsolvedSets[name] = res.Packages
		repoSets[name] = res.Repos
	}
	return depsolvedSets, repoSets, nil
}

func outputManifest(distroName, imgTypeStr, archStr string, opts *cmdlineOpts, blueprintPath string) error {
	res, err := getOneImage(opts.dataDir, distroName, imgTypeStr, archStr)
	if err != nil {
		return err
	}

	// XXX: share with build
	bp, err := loadBlueprint(blueprintPath)
	if err != nil {
		return err
	}

	// XXX: what/how much do we expose here?
	options := distro.ImageOptions{}
	distro := res.Distro
	imgType := res.ImgType

	reporeg, err := newRepoRegistry(opts.dataDir)
	if err != nil {
		return err
	}
	// do not use distroName/imgTypeStr/archStr as they may contain search
	// prefixes
	repos, err := reporeg.ReposByImageTypeName(res.Distro.Name(), res.Arch.Name(), res.ImgType.Name())
	if err != nil {
		return err
	}
	preManifest, warnings, err := imgType.Manifest(bp, options, repos, 0)
	if err != nil {
		return err
	}
	if len(warnings) > 0 {
		// XXX: what can we do here? for things like json output?
		// what are these warnings?
		return fmt.Errorf("warnings during manifest creation: %v", strings.Join(warnings, "\n"))
	}
	// XXX: cleanup, use shared dir,etc
	cacheDir, err := os.MkdirTemp("", "depsolve")
	if err != nil {
		return err
	}
	packageSpecs, _, err := depsolve(cacheDir, preManifest.GetPackageSetChains(), distro, res.Arch.Name())
	if err != nil {
		return err
	}
	if packageSpecs == nil {
		return fmt.Errorf("depsolve did not return any packages")
	}
	// XXX: support commit/container specs
	mf, err := preManifest.Serialize(packageSpecs, nil, nil, nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(opts.out, "%s\n", mf)

	return nil
}
