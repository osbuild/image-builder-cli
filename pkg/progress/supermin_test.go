package progress_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/osbuild/image-builder-cli/pkg/progress"
)

func TestWaitForFilesFile(t *testing.T) {
	tmpdir := t.TempDir()

	// trivial case, no file appears, we error
	start := time.Now()
	canary1 := filepath.Join(tmpdir, "f1.txt")
	err := progress.WaitForFiles(200*time.Millisecond, canary1)
	assert.EqualError(t, err, fmt.Sprintf("files missing after 200ms: [%s]", canary1))
	// ensure we waited untilthe timeout
	assert.True(t, time.Since(start) >= 200*time.Millisecond)

	// trivial case, file is already there before we wait
	err = os.WriteFile(canary1, nil, 0644)
	assert.NoError(t, err)
	// use an absurd high time to ensure test timeouts if we would
	// wait a long time here
	err = progress.WaitForFiles(1*time.Hour, canary1)
	assert.NoError(t, err)

	// new file appears after 100ms
	canary2 := filepath.Join(tmpdir, "f2.txt")
	start = time.Now()
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := os.WriteFile(canary2, nil, 0644)
		assert.NoError(t, err)
	}()
	err = progress.WaitForFiles(1*time.Hour, canary1, canary2)
	assert.NoError(t, err)
	assert.True(t, time.Since(start) >= 100*time.Millisecond)
	// it should take 100-200msec to get the file but to avoid
	// races in heavy loaded CI VMs we are conservative here
	assert.True(t, time.Since(start) <= time.Second)
}
