package manifestgen_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/distrofactory"
	"github.com/osbuild/images/pkg/imagefilter"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/reporegistry"
	"github.com/osbuild/images/pkg/rpmmd"

	"github.com/osbuild/image-builder-cli/internal/manifestgen"
	"github.com/osbuild/image-builder-cli/internal/manifesttest"
)

func init() {
	// silence logrus by default, it is quite verbose
	logrus.SetLevel(logrus.WarnLevel)
}

func sha256For(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("sha256:%x", bs)
}

func TestManifestGeneratorDepsolve(t *testing.T) {
	var osbuildManifest bytes.Buffer

	dataDir := "../../test/data/"
	repos, err := reporegistry.New([]string{dataDir})
	assert.NoError(t, err)

	fac := distrofactory.NewTestDefault()
	filter, err := imagefilter.New(fac, repos)
	assert.NoError(t, err)
	res, err := filter.Filter("distro:test-distro-1", "type:test_type", "arch:test_arch")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	opts := &manifestgen.Options{
		Output:            &osbuildManifest,
		Depsolver:         fakeDepsolve,
		CommitResolver:    panicCommitResolver,
		ContainerResolver: panicContainerResolver,
	}
	mg, err := manifestgen.New(repos, opts)
	assert.NoError(t, err)
	assert.NotNil(t, mg)
	var bp blueprint.Blueprint
	err = mg.Generate(&bp, res[0].Distro, res[0].ImgType, res[0].Arch, nil)
	assert.NoError(t, err)

	pipelineNames, err := manifesttest.PipelineNamesFrom(osbuildManifest.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, []string{"build", "os"}, pipelineNames)

	// we expect at least a "include-rpm-kernel" package in the manifest,
	// sadly the test distro does not really generate much here so we
	// need to use this as a canary that resolving happend
	// XXX: add testhelper to manifesttest for this
	expectedSha256 := sha256For("include-rpm-kernel")
	assert.Contains(t, osbuildManifest.String(), expectedSha256)
}

func TestManifestGeneratorWithOstreeCommit(t *testing.T) {
	var osbuildManifest bytes.Buffer

	dataDir := "../../test/data/"
	repos, err := reporegistry.New([]string{dataDir})
	assert.NoError(t, err)

	fac := distrofactory.NewTestDefault()
	filter, err := imagefilter.New(fac, repos)
	assert.NoError(t, err)
	res, err := filter.Filter("distro:test-distro-1", "type:rhel-edge-commit", "arch:test_arch3")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	opts := &manifestgen.Options{
		Output:            &osbuildManifest,
		Depsolver:         fakeDepsolve,
		CommitResolver:    fakeCommitResolver,
		ContainerResolver: panicContainerResolver,
	}
	mg, err := manifestgen.New(repos, opts)
	assert.NoError(t, err)
	assert.NotNil(t, mg)
	var bp blueprint.Blueprint
	err = mg.Generate(&bp, res[0].Distro, res[0].ImgType, res[0].Arch, nil)
	assert.NoError(t, err)

	pipelineNames, err := manifesttest.PipelineNamesFrom(osbuildManifest.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, []string{"build", "os"}, pipelineNames)

	// XXX: add testhelper to manifesttest for this
	assert.Contains(t, osbuildManifest.String(), `{"url":"resolved-url-for-test/1/x86_64/edge"}`)
	// we expect at least a "include-rpm-kernel" package in the manifest,
	// sadly the test distro does not really generate much here so we
	// need to use this as a canary that resolving happend
	// XXX: add testhelper to manifesttest for this
	expectedSha256 := sha256For("include-rpm-kernel")
	assert.Contains(t, osbuildManifest.String(), expectedSha256)
}

func fakeDepsolve(cacheDir string, packageSets map[string][]rpmmd.PackageSet, d distro.Distro, arch string) (map[string][]rpmmd.PackageSpec, map[string][]rpmmd.RepoConfig, error) {
	depsolvedSets := make(map[string][]rpmmd.PackageSpec)
	repoSets := make(map[string][]rpmmd.RepoConfig)
	for name, pkgSets := range packageSets {
		var resolvedSet []rpmmd.PackageSpec
		for _, pkgSet := range pkgSets {
			for _, pkgName := range pkgSet.Include {
				fakeName := fmt.Sprintf("include-rpm-%s", pkgName)
				resolvedSet = append(resolvedSet, rpmmd.PackageSpec{
					Name:     fakeName,
					Checksum: sha256For(fakeName),
				})
			}
		}
		depsolvedSets[name] = resolvedSet
	}
	return depsolvedSets, repoSets, nil
}

func fakeCommitResolver(commitSources map[string][]ostree.SourceSpec) (map[string][]ostree.CommitSpec, error) {
	commits := make(map[string][]ostree.CommitSpec, len(commitSources))
	for name, commitSources := range commitSources {
		commitSpecs := make([]ostree.CommitSpec, len(commitSources))
		for idx, commitSource := range commitSources {
			commitSpecs[idx] = ostree.CommitSpec{
				URL: fmt.Sprintf("resolved-url-for-%s", commitSource.Ref),
			}
		}
		commits[name] = commitSpecs
	}
	return commits, nil

}

func panicCommitResolver(commitSources map[string][]ostree.SourceSpec) (map[string][]ostree.CommitSpec, error) {
	if len(commitSources) > 0 {
		panic("panicCommitResolver")
	}
	return nil, nil
}

func fakeContainerResolver(containerSources map[string][]container.SourceSpec, archName string) (map[string][]container.Spec, error) {
	containerSpecs := make(map[string][]container.Spec, len(containerSources))
	for plName, sourceSpecs := range containerSources {
		var containers []container.Spec
		for _, spec := range sourceSpecs {
			containers = append(containers, container.Spec{
				Source: fmt.Sprintf("resolved-cnt-%s", spec.Source),
			})
		}
		containerSpecs[plName] = containers
	}
	return containerSpecs, nil
}

func panicContainerResolver(containerSources map[string][]container.SourceSpec, archName string) (map[string][]container.Spec, error) {
	if len(containerSources) > 0 {
		panic("panicContainerResolver")
	}
	return nil, nil
}

// XXX: containers cannot be tested because the test_distro package
// does not provide a test distro that has container content, this
// will need to get added via the manifest.NewContentTest helper
