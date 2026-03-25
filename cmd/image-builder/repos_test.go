package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osbuild/images/pkg/rpmmd"
)

func TestParseRepoURLsHappy(t *testing.T) {
	checkGPG := false

	cfg, err := parseRepoURLs([]string{
		"file:///path/to/repo",
		"https://example.com/repo",
	}, "forced")
	assert.NoError(t, err)
	assert.Equal(t, []rpmmd.RepoConfig{
		{
			Id:           "forced-repo-0",
			Name:         "forced repo#0 /path/to/repo",
			BaseURLs:     []string{"file:///path/to/repo"},
			CheckGPG:     &checkGPG,
			CheckRepoGPG: &checkGPG,
		},
		{
			Id:           "forced-repo-1",
			Name:         "forced repo#1 example.com/repo",
			BaseURLs:     []string{"https://example.com/repo"},
			CheckGPG:     &checkGPG,
			CheckRepoGPG: &checkGPG,
		},
	}, cfg)
}

func TestParseExtraRepoSad(t *testing.T) {
	_, err := parseRepoURLs([]string{"/just/a/path"}, "forced")
	assert.EqualError(t, err, `scheme missing in "/just/a/path", please prefix with e.g. file:// or https://`)

	_, err = parseRepoURLs([]string{"https://example.com", "/just/a/path"}, "forced")
	assert.EqualError(t, err, `scheme missing in "/just/a/path", please prefix with e.g. file:// or https://`)
}

func TestParseRepoURLsFileRepo(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "test.repo")
	err := os.WriteFile(repoPath, []byte(`[fedora]
name=Fedora
baseurl=https://download.fedoraproject.org/pub/fedora/linux/releases/42/Everything/x86_64/os/
gpgcheck=1
repo_gpgcheck=0
`), 0644)
	require.NoError(t, err)

	cfg, err := parseRepoURLs([]string{"file://" + repoPath}, "extra")
	require.NoError(t, err)
	require.Len(t, cfg, 1)
	assert.Equal(t, "extra-repo-0-fedora", cfg[0].Id)
	assert.Equal(t, "Fedora", cfg[0].Name)
	assert.Equal(t, []string{"https://download.fedoraproject.org/pub/fedora/linux/releases/42/Everything/x86_64/os/"}, cfg[0].BaseURLs)
	assert.True(t, cfg[0].CheckGPG != nil && *cfg[0].CheckGPG)
	assert.True(t, cfg[0].CheckRepoGPG != nil && !*cfg[0].CheckRepoGPG)
}

func TestParseRepoURLsFileRepoWithGPGKey(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "RPM-GPG-KEY-test")
	keyContent := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\nxyz\n-----END PGP PUBLIC KEY BLOCK-----")
	require.NoError(t, os.WriteFile(keyPath, keyContent, 0644))

	repoPath := filepath.Join(dir, "test.repo")
	err := os.WriteFile(repoPath, []byte(`[myrepo]
name=My Repo
baseurl=https://example.com/repo
gpgcheck=1
gpgkey=file://`+keyPath+`
`), 0644)
	require.NoError(t, err)

	cfg, err := parseRepoURLs([]string{"file://" + repoPath}, "extra")
	require.NoError(t, err)
	require.Len(t, cfg, 1)
	require.Len(t, cfg[0].GPGKeys, 1)
	assert.Equal(t, string(keyContent), cfg[0].GPGKeys[0])
}

func TestParseRepoURLsFileDirectoryTreatedAsURL(t *testing.T) {
	// file:// to a directory is not a .repo file: treat as single base-URL repo
	dir := t.TempDir()
	cfg, err := parseRepoURLs([]string{"file://" + dir}, "extra")
	require.NoError(t, err)
	require.Len(t, cfg, 1)
	assert.Equal(t, "extra-repo-0", cfg[0].Id)
	assert.Equal(t, []string{"file://" + dir}, cfg[0].BaseURLs)
}

func TestNewRepoRegistryImplSmoke(t *testing.T) {
	registry, err := newRepoRegistryImpl("", nil)
	require.NoError(t, err)
	repos, err := registry.DistroHasRepos("rhel-10.2", "x86_64")
	require.NoError(t, err)
	assert.True(t, len(repos) > 0)
}

func TestNewRepoRegistryImplExtraReposGetAppended(t *testing.T) {
	registry, err := newRepoRegistryImpl("", []string{"https://example.com/my/repo"})
	require.NoError(t, err)
	repos, err := registry.DistroHasRepos("rhel-10.2", "x86_64")
	require.NoError(t, err)
	assert.Equal(t, repos[len(repos)-1].BaseURLs[0], "https://example.com/my/repo")
}

func TestNewRepoRegistryImplRepodir(t *testing.T) {
	// prereq test: no testdistro-1 in the default repos
	registry, err := newRepoRegistryImpl("", nil)
	require.NoError(t, err)
	assert.NotContains(t, registry.ListDistros(), "testdistro-1")
	_, err = registry.DistroHasRepos("testdistro-1", "x86_64")
	require.EqualError(t, err, `requested repository not found: for distribution "testdistro-1"`)

	// create a custom repodir with testdistro-1.json, the basefilename
	// must match a distro nameVer
	repoDir := t.TempDir()
	repoFile := filepath.Join(repoDir, "repositories", "testdistro-1.json")
	err = os.Mkdir(filepath.Dir(repoFile), 0755)
	require.NoError(t, err)
	repoContents := `{
	"x86_64": [
		{
			"name": "testdistro-1-repo",
			"baseurl": "https://example.com/test/test/distro/1"
		}
	]
}
`
	err = os.WriteFile(repoFile, []byte(repoContents), 0644)
	require.NoError(t, err)

	// and ensure we have testdistro-1 now
	registry, err = newRepoRegistryImpl(repoDir, nil)
	require.NoError(t, err)
	repos, err := registry.DistroHasRepos("testdistro-1", "x86_64")
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, repos[0].Name, "testdistro-1-repo")
}

func TestNewRepoRegistryImplRepodirNoSubDir(t *testing.T) {
	// prereq test: no testdistro-1 in the default repos
	registry, err := newRepoRegistryImpl("", nil)
	require.NoError(t, err)
	assert.NotContains(t, registry.ListDistros(), "testdistro-1")
	_, err = registry.DistroHasRepos("testdistro-1", "x86_64")
	require.EqualError(t, err, `requested repository not found: for distribution "testdistro-1"`)

	// create a custom repodir with testdistro-1.json, the basefilename
	// must match a distro nameVer
	repoDir := t.TempDir()
	repoFile := filepath.Join(repoDir, "testdistro-1.json")
	repoContents := `{
	"x86_64": [
		{
			"name": "testdistro-1-repo",
			"baseurl": "https://example.com/test/test/distro/1"
		}
	]
}
`
	err = os.WriteFile(repoFile, []byte(repoContents), 0644)
	require.NoError(t, err)

	// and ensure we have testdistro-1 now
	registry, err = newRepoRegistryImpl(repoDir, nil)
	require.NoError(t, err)
	repos, err := registry.DistroHasRepos("testdistro-1", "x86_64")
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, repos[0].Name, "testdistro-1-repo")
}
