package main

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"

	"github.com/osbuild/images/pkg/distrofactory"
	"github.com/osbuild/images/pkg/imagefilter"

	ImageModel "github.com/osbuild/image-builder-cli/pkg/image_model"
)

func newImageFilterDefault(repoDir string, extraRepos []string) (*imagefilter.ImageFilter, error) {
	fac := distrofactory.NewDefault()
	repos, err := newRepoRegistry(repoDir, extraRepos)
	if err != nil {
		return nil, err
	}

	return imagefilter.New(fac, repos)
}

type repoOptions struct {
	// RepoDir contains the base dir for the repo definition search path, it will also look
	// in the `repositories` subdirectory to RepoDir
	RepoDir string

	// ExtraRepos contains extra baseURLs that get added to the depsolving
	ExtraRepos []string

	// ForceRepos contains baseURLs that replace *all* base repositories
	ForceRepos []string
}

// should this be moved to images:imagefilter?
func getOneImage(distroName string, imgFormat ImageModel.CLIOutputFormat, archStr string, repoOpts *repoOptions) (*imagefilter.Result, error) {
	if repoOpts == nil {
		repoOpts = &repoOptions{}
	}

	imageFilter, err := newImageFilterDefault(repoOpts.RepoDir, repoOpts.ExtraRepos)
	if err != nil {
		return nil, err
	}
	// strip prefixes to make ib copy/paste friendly when pasting output
	// from "list"
	distroName = strings.TrimPrefix(distroName, "distro:")
	// Not needed anymore, because validation was done earlier?
	// imgTypeStr = strings.TrimPrefix(imgTypeStr, "type:")
	archStr = strings.TrimPrefix(archStr, "arch:")

	// imgFormat check not needed anymore, because validation was done earlier?
	// for _, s := range []string{distroName, imgFormat, archStr} {
	// error early when globs are used
	for _, s := range []string{distroName, archStr} {
		if glob.QuoteMeta(s) != s {
			return nil, fmt.Errorf("cannot use globs in %q when getting a single image", s)
		}
	}

	filterExprs := []string{
		fmt.Sprintf("distro:%s", distroName),
		fmt.Sprintf("arch:%s", archStr),
		fmt.Sprintf("type:%s", imgFormat),
	}
	filteredResults, err := imageFilter.Filter(filterExprs...)
	if err != nil {
		return nil, err
	}
	switch len(filteredResults) {
	case 0:
		return nil, fmt.Errorf("cannot find image for: distro:%q type:%q arch:%q", distroName, imgFormat, archStr)
	case 1:
		return &filteredResults[0], nil
	default:
		// This condition should never be hit in practise as we
		// disallow globs above.
		// XXX: imagefilter.Result should have a String() method so
		// that this output can actually show the results
		return nil, fmt.Errorf("internal error: found %v results for %q %q %q", len(filteredResults), distroName, imgFormat, archStr)
	}
}

// getAllImages returns all images matching the filter expressions.
func getAllImages(repoOpts *repoOptions, filterExprs ...string) ([]imagefilter.Result, error) {
	if repoOpts == nil {
		repoOpts = &repoOptions{}
	}

	imageFilter, err := newImageFilterDefault(repoOpts.RepoDir, repoOpts.ExtraRepos)
	if err != nil {
		return nil, err
	}

	filteredResults, err := imageFilter.Filter(filterExprs...)
	if err != nil {
		return nil, err
	}

	return filteredResults, nil
}
