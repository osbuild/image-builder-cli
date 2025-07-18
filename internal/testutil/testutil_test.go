package testutil_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osbuild/image-builder-cli/internal/testutil"
)

func TestMockCommand(t *testing.T) {
	fakeCmd := testutil.MockCommand(t, "false", "exit 0")

	err := exec.Command("false", "run1-arg1", "run1-arg2").Run()
	assert.NoError(t, err)
	err = exec.Command("false", "run2-arg1", "run2-arg2").Run()
	assert.NoError(t, err)

	assert.Equal(t, [][]string{
		{"run1-arg1", "run1-arg2"},
		{"run2-arg1", "run2-arg2"},
	}, fakeCmd.CallArgsList())
}

func TestCaptureStdout(t *testing.T) {
	stdout, stderr := testutil.CaptureStdio(t, func() {
		fmt.Fprintf(os.Stdout, "output on stdout")
		fmt.Fprintf(os.Stderr, "output on stderr")
	})
	assert.Equal(t, "output on stdout", stdout)
	assert.Equal(t, "output on stderr", stderr)
}

func TestChroot(t *testing.T) {
	tmpdir := t.TempDir()
	testutil.Chdir(t, tmpdir, func() {
		cwd, err := os.Getwd()
		assert.NoError(t, err)
		assert.Equal(t, tmpdir, cwd)
	})
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NotEqual(t, tmpdir, cwd)
}
