package main

import (
	"io"
	"slices"

	"gopkg.in/yaml.v3"

	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/imagefilter"
)

// Use yaml output by default because it is both nicely human and
// machine readable and parts of our image defintions will be written
// in yaml too.  This means this should be a possible input a
// "flattended" image definiton.
type describeImgYAML struct {
	Distro string `yaml:"distro"`
	Type   string `yaml:"type"`
	Arch   string `yaml:"arch"`

	// XXX: think about ordering (as this is what the user will see)
	OsVersion string `yaml:"os_vesion"`

	Bootmode        string `yaml:"bootmode"`
	PartitionType   string `yaml:"partition_type"`
	DefaultFilename string `yaml:"default_filename"`

	// XXX: add pipelines here? maybe at least exports?
	Packages *packagesYAML `yaml:"packages"`
}

type packagesYAML struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

func packageSetsFor(imgType distro.ImageType) (inc, exc []string, err error) {
	var bp blueprint.Blueprint
	manifest, _, err := imgType.Manifest(&bp, distro.ImageOptions{}, nil, 0)
	if err != nil {
		return nil, nil, err
	}

	// XXX: or should this just do what osbuild-package-sets does
	// and inlcude what pipeline needs the package set too?
	for pipelineName, pkgSets := range manifest.GetPackageSetChains() {
		// XXX: or shouldn't we exclude the build pipeline here?
		if pipelineName == "build" {
			continue
		}
		for _, pkgSet := range pkgSets {
			inc = append(inc, pkgSet.Include...)
			exc = append(exc, pkgSet.Exclude...)
		}
	}
	slices.Sort(inc)
	slices.Sort(exc)
	return inc, exc, nil
}

// XXX: should this live in images instead?
func describeImage(img *imagefilter.Result, out io.Writer) error {
	// see
	// https://github.com/osbuild/images/pull/1019#discussion_r1832376568
	// for what is available on an image (without depsolve or partitioning)
	inc, exc, err := packageSetsFor(img.ImgType)
	if err != nil {
		return err
	}

	outYaml := &describeImgYAML{
		Distro:          img.Distro.Name(),
		OsVersion:       img.Distro.OsVersion(),
		Arch:            img.Arch.Name(),
		Type:            img.ImgType.Name(),
		Bootmode:        img.ImgType.BootMode().String(),
		PartitionType:   img.ImgType.PartitionType().String(),
		DefaultFilename: img.ImgType.Filename(),
		Packages: &packagesYAML{
			Include: inc,
			Exclude: exc,
		},
	}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	return enc.Encode(outYaml)
}
