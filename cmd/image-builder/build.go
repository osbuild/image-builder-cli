package main

import (
	"fmt"
	"path/filepath"

	"github.com/osbuild/images/pkg/imagefilter"
	"github.com/osbuild/images/pkg/osbuild"
)

func buildImage(res *imagefilter.Result, osbuildManifest []byte) error {
	osbuildStoreDir := ".store"
	// XXX: support output dir via commandline
	outputDir := "."
	buildName := fmt.Sprintf("%s-%s-%s", res.Distro.Name(), res.ImgType.Name(), res.Arch.Name())
	jobOutputDir := filepath.Join(outputDir, buildName)

	// XXX: support stremaing via statusWriter
	_, err := osbuild.RunOSBuild(osbuildManifest, osbuildStoreDir, jobOutputDir, res.ImgType.Exports(), nil, nil, false, osStderr)
	return err

}
