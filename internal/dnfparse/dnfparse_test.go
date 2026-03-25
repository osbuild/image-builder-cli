package dnfparse

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trueVal  = true
	falseVal = false
)

// wantRepo describes expected Repo fields for table-driven tests. All booleans are pointers to match RepoConfig.
type wantRepo struct {
	id            string
	name          string
	baseURLs      []string
	checkGPG      *bool
	checkRepoGPG  *bool
	ignoreSSL     *bool
	gpgKeys       []string
	sslCACert     string
	sslClientCert string
	sslClientKey  string
}

func assertOptionalBool(t *testing.T, want, got *bool, label string) {
	t.Helper()
	if want == nil {
		assert.Nil(t, got, "%s (expected nil)", label)
		return
	}
	require.NotNil(t, got, "%s", label)
	assert.Equal(t, *want, *got, "%s", label)
}

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []wantRepo
	}{
		{
			name: "fedora",
			input: `
[main]
# global defaults ignored

[fedora]
name=Fedora $releasever - $basearch
baseurl=https://download.fedoraproject.org/pub/fedora/linux/releases/$releasever/Everything/$basearch/os/
gpgcheck=1
repo_gpgcheck=1

[updates]
name=Fedora $releasever - $basearch - Updates
baseurl=https://download.fedoraproject.org/pub/fedora/linux/updates/$releasever/Everything/$basearch/
gpgcheck=1
repo_gpgcheck=0
`,
			want: []wantRepo{
				{
					id:           "fedora",
					name:         "Fedora $releasever - $basearch",
					baseURLs:     []string{"https://download.fedoraproject.org/pub/fedora/linux/releases/$releasever/Everything/$basearch/os/"},
					checkGPG:     &trueVal,
					checkRepoGPG: &trueVal,
					ignoreSSL:    nil,
					gpgKeys:      nil,
				},
				{
					id:           "updates",
					name:         "Fedora $releasever - $basearch - Updates",
					baseURLs:     []string{"https://download.fedoraproject.org/pub/fedora/linux/updates/$releasever/Everything/$basearch/"},
					checkGPG:     &trueVal,
					checkRepoGPG: &falseVal,
					ignoreSSL:    nil,
					gpgKeys:      nil,
				},
			},
		},
		{
			name: "rhel",
			input: `
# Managed by (rhsm) subscription-manager

[rhel-10-for-x86_64-baseos-rpms]
name = Red Hat Enterprise Linux 10 for x86_64 - BaseOS (RPMs)
baseurl = https://satellite.example.com/pulp/content/Default_Organization/Library/content/dist/rhel10/$releasever/x86_64/baseos/os
enabled = 1
gpgcheck = 1
gpgkey = file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release
sslverify = 1
sslcacert = /etc/rhsm/ca/katello-server-ca.pem
sslclientkey = /etc/pki/entitlement/3619656444745875922-key.pem
sslclientcert = /etc/pki/entitlement/3619656444745875922.pem
metadata_expire = 1
enabled_metadata = 1

[rhel-10-for-x86_64-appstream-rpms]
name = Red Hat Enterprise Linux 10 for x86_64 - AppStream (RPMs)
baseurl = https://satellite.example.com/pulp/content/Default_Organization/Library/content/dist/rhel10/$releasever/x86_64/appstream/os
enabled = 1
gpgcheck = 1
gpgkey = file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release
sslverify = 1
sslcacert = /etc/rhsm/ca/katello-server-ca.pem
sslclientkey = /etc/pki/entitlement/3619656444745875922-key.pem
sslclientcert = /etc/pki/entitlement/3619656444745875922.pem
metadata_expire = 1
enabled_metadata = 1`,
			want: []wantRepo{
				{
					id:            "rhel-10-for-x86_64-baseos-rpms",
					name:          "Red Hat Enterprise Linux 10 for x86_64 - BaseOS (RPMs)",
					baseURLs:      []string{"https://satellite.example.com/pulp/content/Default_Organization/Library/content/dist/rhel10/$releasever/x86_64/baseos/os"},
					checkGPG:      &trueVal,
					checkRepoGPG:  nil,
					ignoreSSL:     &falseVal,
					gpgKeys:       []string{"file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release"},
					sslCACert:     "/etc/rhsm/ca/katello-server-ca.pem",
					sslClientCert: "/etc/pki/entitlement/3619656444745875922.pem",
					sslClientKey:  "/etc/pki/entitlement/3619656444745875922-key.pem",
				},
				{
					id:            "rhel-10-for-x86_64-appstream-rpms",
					name:          "Red Hat Enterprise Linux 10 for x86_64 - AppStream (RPMs)",
					baseURLs:      []string{"https://satellite.example.com/pulp/content/Default_Organization/Library/content/dist/rhel10/$releasever/x86_64/appstream/os"},
					checkGPG:      &trueVal,
					checkRepoGPG:  nil,
					ignoreSSL:     &falseVal,
					gpgKeys:       []string{"file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release"},
					sslCACert:     "/etc/rhsm/ca/katello-server-ca.pem",
					sslClientCert: "/etc/pki/entitlement/3619656444745875922.pem",
					sslClientKey:  "/etc/pki/entitlement/3619656444745875922-key.pem",
				},
			},
		},
		{
			name: "multiple baseurls",
			input: `
[myrepo]
name=My Repo
baseurl=https://one.com/repo
baseurl=https://two.com/repo
gpgcheck=0
`,
			want: []wantRepo{
				{
					id:           "myrepo",
					name:         "My Repo",
					baseURLs:     []string{"https://one.com/repo", "https://two.com/repo"},
					checkGPG:     &falseVal,
					checkRepoGPG: nil,
					ignoreSSL:    nil,
					gpgKeys:      nil,
				},
			},
		},
		{
			name: "sslverify=0",
			input: `
[insecure]
name=Insecure Repo
baseurl=https://insecure.example.com/repo
sslverify=0
`,
			want: []wantRepo{
				{
					id:           "insecure",
					name:         "Insecure Repo",
					baseURLs:     []string{"https://insecure.example.com/repo"},
					checkGPG:     nil,
					checkRepoGPG: nil,
					ignoreSSL:    &trueVal,
					gpgKeys:      nil,
				},
			},
		},
		{
			name: "skips section without baseurl",
			input: `
[main]
name=Main config

[hasbase]
name=Has baseurl
baseurl=https://example.com/repo
`,
			want: []wantRepo{
				{
					id:           "hasbase",
					name:         "Has baseurl",
					baseURLs:     []string{"https://example.com/repo"},
					checkGPG:     nil,
					checkRepoGPG: nil,
					ignoreSSL:    nil,
					gpgKeys:      nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, err := Parse(strings.NewReader(tt.input), "test.repo")
			require.NoError(t, err)
			require.Len(t, repos, len(tt.want), "number of repos")
			for i, w := range tt.want {
				r := repos[i]
				assert.Equal(t, w.id, r.Id, "repo[%d].Id", i)
				assert.Equal(t, w.name, r.Name, "repo[%d].Name", i)
				assert.Equal(t, w.baseURLs, r.BaseURLs, "repo[%d].BaseURLs", i)
				assertOptionalBool(t, w.checkGPG, r.CheckGPG, fmt.Sprintf("repo[%d].CheckGPG", i))
				assertOptionalBool(t, w.checkRepoGPG, r.CheckRepoGPG, fmt.Sprintf("repo[%d].CheckRepoGPG", i))
				assertOptionalBool(t, w.ignoreSSL, r.IgnoreSSL, fmt.Sprintf("repo[%d].IgnoreSSL", i))
				if w.gpgKeys != nil {
					assert.Equal(t, w.gpgKeys, r.GPGKeys, "repo[%d].GPGKeys", i)
				}
				assert.Equal(t, w.sslCACert, r.SSLCACert, "repo[%d].SSLCACert", i)
				assert.Equal(t, w.sslClientCert, r.SSLClientCert, "repo[%d].SSLClientCert", i)
				assert.Equal(t, w.sslClientKey, r.SSLClientKey, "repo[%d].SSLClientKey", i)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"1", "1", true},
		{"yes", "yes", true},
		{"true", "true", true},
		{"0", "0", false},
		{"no", "no", false},
		{"false", "false", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBool(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGPGKeyContents(t *testing.T) {
	tests := []struct {
		name        string
		repo        func(t *testing.T) *Repo
		wantContent []byte
		wantErr     string
	}{
		{
			name: "single file",
			repo: func(t *testing.T) *Repo {
				dir := t.TempDir()
				path := dir + "/RPM-GPG-KEY-test"
				content := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\nxyz\n-----END PGP PUBLIC KEY BLOCK-----")
				require.NoError(t, os.WriteFile(path, content, 0644))
				return &Repo{GPGKeys: []string{"file://" + path}}
			},
			wantContent: []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\nxyz\n-----END PGP PUBLIC KEY BLOCK-----"),
			wantErr:     "",
		},
		{
			name: "multiple files",
			repo: func(t *testing.T) *Repo {
				dir := t.TempDir()
				k1, k2 := dir+"/key1", dir+"/key2"
				require.NoError(t, os.WriteFile(k1, []byte("key1"), 0644))
				require.NoError(t, os.WriteFile(k2, []byte("key2"), 0644))
				return &Repo{GPGKeys: []string{"file://" + k1, "file://" + k2}}
			},
			wantContent: []byte("key1\nkey2"),
			wantErr:     "",
		},
		{
			name: "non-file URI",
			repo: func(t *testing.T) *Repo {
				return &Repo{GPGKeys: []string{"https://example.com/key.asc"}}
			},
			wantErr: "only file:// URIs are supported",
		},
		{
			name: "empty GPGKeys",
			repo: func(t *testing.T) *Repo {
				return &Repo{GPGKeys: nil}
			},
			wantContent: nil,
			wantErr:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo(t)
			got, err := repo.GPGKeyContents()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, got)
		})
	}
}
