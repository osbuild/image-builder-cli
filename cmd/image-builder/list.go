package main

import (
	"os"

	"github.com/osbuild/images/pkg/imagefilter"
)

func listImages(dataDir string, extraRepos []string, output string, filterExprs []string) error {
	imageFilter, err := newImageFilterDefault(dataDir, extraRepos)
	if err != nil {
		return err
	}

	filteredResult, err := imageFilter.Filter(filterExprs...)
	if err != nil {
		return err
	}

	fmter, err := imagefilter.NewResultsFormatter(imagefilter.OutputFormat(output))
	if err != nil {
		return err
	}
	if err := fmter.Output(os.Stdout, filteredResult); err != nil {
		return err
	}

	return nil
}
