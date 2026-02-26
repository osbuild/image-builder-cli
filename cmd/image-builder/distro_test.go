package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	main "github.com/osbuild/image-builder-cli/cmd/image-builder"
	"github.com/osbuild/image-builder-cli/internal/testutil"
)

func TestFindDistro(t *testing.T) {
	for _, tc := range []struct {
		argDistro      string
		bpDistro       string
		expectedDistro string
		expectedErr    string
	}{
		{"arg", "", "arg", ""},
		{"", "bp", "bp", ""},
		{"arg", "bp", "", `error selecting distro name, cmdline argument "arg" is different from blueprint "bp"`},
		// the argDistro,bpDistro == "" case is tested below
	} {
		distro, err := main.FindDistro(tc.argDistro, tc.bpDistro)
		if tc.expectedErr != "" {
			assert.Equal(t, tc.expectedErr, err.Error())
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedDistro, distro)
		}
	}
}

func TestFindDistroAutoDetect(t *testing.T) {
	restore := main.MockDistroGetHostDistroName(func() (string, error) {
		return "mocked-host-distro", nil
	})
	defer restore()

	var err error
	var distro string
	_, stderr := testutil.CaptureStdio(t, func() {
		distro, err = main.FindDistro("", "")
	})
	assert.NoError(t, err)
	assert.Equal(t, "mocked-host-distro", distro)
	assert.Equal(t, "No distro name specified, selecting \"mocked-host-distro\" based on host, use --distro to override\n", stderr)
}
