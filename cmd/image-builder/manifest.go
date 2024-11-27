package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/dnfjson"
	"github.com/osbuild/images/pkg/rpmmd"
	"github.com/osbuild/images/pkg/sbom"
)

// XXX: duplicated from cmd/build/main.go:depsolve (and probably more places)
// should go into a common helper in "images" or images should do this on
// its own
var depsolve = func(cacheDir string, packageSets map[string][]rpmmd.PackageSet, d distro.Distro, arch string) (map[string][]rpmmd.PackageSpec, map[string][]rpmmd.RepoConfig, error) {
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

type manifestGenerator struct {
	DataDir string
	Out     io.Writer
}

func (mg *manifestGenerator) Generate(distroStr, imgTypeStr, archStr string) error {
	// Note that distroStr/imgTypeStr/archStr may contain prefixes like
	// "distro:" so only the getOneImage() result should be used in the
	// rest of the function.
	res, err := getOneImage(mg.DataDir, distroStr, imgTypeStr, archStr)
	if err != nil {
		return err
	}

	var bp blueprint.Blueprint
	var options distro.ImageOptions
	reporeg, err := newRepoRegistry(mg.DataDir)
	if err != nil {
		return err
	}
	repos, err := reporeg.ReposByImageTypeName(res.Distro.Name(), res.Arch.Name(), res.ImgType.Name())
	if err != nil {
		return err
	}
	preManifest, warnings, err := res.ImgType.Manifest(&bp, options, repos, 0)
	if err != nil {
		return err
	}
	if len(warnings) > 0 {
		// XXX: what can we do here? for things like json output?
		// what are these warnings?
		return fmt.Errorf("warnings during manifest creation: %v", strings.Join(warnings, "\n"))
	}
	// XXX: add something like "--rpmmd" (like bib)
	cacheDir := ""
	packageSpecs, _, err := depsolve(cacheDir, preManifest.GetPackageSetChains(), res.Distro, res.Arch.Name())
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
	fmt.Fprintf(mg.Out, "%s\n", mf)

	return nil
}
