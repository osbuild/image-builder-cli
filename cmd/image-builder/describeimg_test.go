package main_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	testrepos "github.com/osbuild/images/test/data/repositories"

	"github.com/osbuild/image-builder-cli/cmd/image-builder"
)

func TestDescribeImage(t *testing.T) {
	restore := main.MockNewRepoRegistry(testrepos.New)
	defer restore()

	res, err := main.GetOneImage("", "centos-9", "tar", "x86_64")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = main.DescribeImage(res, &buf)
	assert.NoError(t, err)

	expectedOutput := `distro: centos-9
type: tar
arch: x86_64
os_vesion: 9-stream
bootmode: none
partition_type: ""
default_filename: root.tar.xz
packages:
  include:
    - policycoreutils
    - selinux-policy-targeted
    - selinux-policy-targeted
  exclude:
    - rng-tools
`
	assert.Equal(t, expectedOutput, buf.String())
}
