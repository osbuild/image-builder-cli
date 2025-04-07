package progress_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osbuild/image-builder-cli/pkg/progress"
)

func TestEnoughPrivsSmoke(t *testing.T) {
	enoughPrivs, err := progress.EnoughPrivsForOsbuild()
	assert.NoError(t, err)
	assert.Equal(t, enoughPrivs, os.Getuid() == 0)
}
