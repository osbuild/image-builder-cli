package main

import (
	"fmt"
	"path/filepath"

	"github.com/osbuild/images/pkg/imagefilter"
	"github.com/osbuild/images/pkg/osbuild/progress"
)

func buildImage(pbar progress.ProgressBar, res *imagefilter.Result, osbuildManifest []byte, osbuildStoreDir string) error {
	// XXX: support output dir via commandline
	// XXX2: support output filename via commandline (c.f.
	//   https://github.com/osbuild/images/pull/1039)
	outputDir := "."
	buildName := fmt.Sprintf("%s-%s-%s", res.Distro.Name(), res.ImgType.Name(), res.Arch.Name())
	jobOutputDir := filepath.Join(outputDir, buildName)

	return progress.RunOSBuild(pbar, osbuildManifest, osbuildStoreDir, jobOutputDir, res.ImgType.Exports(), nil)
}
