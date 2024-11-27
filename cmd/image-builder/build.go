package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/osbuild/images/pkg/arch"
	"github.com/osbuild/images/pkg/osbuild"
)

func buildImage(distroName, imgTypeStr, archStr string, opts *cmdlineOpts) error {
	res, err := getOneImage(opts.dataDir, distroName, imgTypeStr, archStr)
	if err != nil {
		return err
	}
	// cross arch building is not possible, we would have to download
	// a pre-populated buildroot (tar,container) with rpm for that
	if res.Arch.Name() != arch.Current().String() {
		return fmt.Errorf("cannot build for arch %q from %q", res.Arch.Name(), arch.Current().String())
	}

	imgType := res.ImgType

	var mf bytes.Buffer
	// XXX: so messy, do not abuse cmdlineOpts.out for this buffer,
	// refactor outputManifest instead
	opts.out = &mf
	if err := outputManifest(res.Distro.Name(), res.ImgType.Name(), res.Arch.Name(), opts); err != nil {
		return err
	}

	osbuildStoreDir := ".store"
	outputDir := "."
	buildName := fmt.Sprintf("%s-%s-%s", res.Distro.Name(), res.ImgType.Name(), res.Arch.Name())
	jobOutputDir := filepath.Join(outputDir, buildName)
	// XXX: support stremaing via statusWriter
	_, err = osbuild.RunOSBuild(mf.Bytes(), osbuildStoreDir, jobOutputDir, imgType.Exports(), nil, nil, false, os.Stderr)
	return err
}
