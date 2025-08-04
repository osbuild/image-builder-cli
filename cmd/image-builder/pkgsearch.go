package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/osbuild/image-builder-cli/pkg/progress"
	"github.com/osbuild/images/pkg/dnfjson"
	"github.com/osbuild/images/pkg/rpmmd"
)

type pkgSearchFormatter interface {
	Output(io.Writer, rpmmd.PackageList) error
}

func newPkgSearchFormatter(format string) (pkgSearchFormatter, error) {
	switch format {
	case "json":
		return &jsonPkgSearchFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported results formatter %q (try --format=json)", format)
	}
}

type pkgSearchResultJSON struct {
	Packages rpmmd.PackageList
}
type jsonPkgSearchFormatter struct{}

func (*jsonPkgSearchFormatter) Output(w io.Writer, pkgs rpmmd.PackageList) error {
	enc := json.NewEncoder(w)
	return enc.Encode(struct {
		Packages rpmmd.PackageList
	}{
		Packages: pkgs,
	})
}

func cmdPkgSearch(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}

	pbar, err := progress.New("")
	if err != nil {
		return err
	}
	img, err := cmdGetOneImage(pbar, cmd, args)
	if err != nil {
		return err
	}
	// XXX: set a sensible cachedir here
	cacheDir := ""
	d := img.Distro

	solver := dnfjson.NewSolver(d.ModulePlatformID(), d.Releasever(), img.Arch.Name(), d.Name(), cacheDir)
	results, err := solver.SearchMetadata(img.Repos, args[1:])
	if err != nil {
		return err
	}
	fmter, err := newPkgSearchFormatter(format)
	if err != nil {
		return err
	}
	if err := fmter.Output(osStdout, results); err != nil {
		return err
	}

	return nil
}
