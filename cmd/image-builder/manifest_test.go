package main_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osbuild/images/pkg/distrofactory"
	testrepos "github.com/osbuild/images/test/data/repositories"

	"github.com/osbuild/image-builder-cli/cmd/image-builder"
)

func TestManifestGeneratorSad(t *testing.T) {
	restore := main.MockNewRepoRegistry(testrepos.New)
	defer restore()

	mg := &main.ManifestGenerator{}
	assert.NotNil(t, mg)
	err := mg.Generate("bad-distro", "bad-type", "bad-arch")
	assert.EqualError(t, err, `cannot find image for: distro:"bad-distro" type:"bad-type" arch:"bad-arch"`)
}

// XXX: move to images:testutil.go or something
func pipelineNamesFrom(t *testing.T, osbuildManifest []byte) []string {
	var manifest map[string]interface{}

	err := json.Unmarshal(osbuildManifest, &manifest)
	assert.NoError(t, err)
	assert.NotNil(t, manifest["pipelines"])
	pipelines := manifest["pipelines"].([]interface{})
	pipelineNames := make([]string, len(pipelines))
	for idx, pi := range pipelines {
		pipelineNames[idx] = pi.(map[string]interface{})["name"].(string)
	}
	return pipelineNames
}

func TestManifestGenerator(t *testing.T) {
	var osbuildManifest bytes.Buffer

	restore := main.MockDistrofactoryNew(distrofactory.NewTestDefault)
	defer restore()
	restore = main.MockDepsolve()
	defer restore()

	mg := &main.ManifestGenerator{DataDir: "../../test/data/", Out: &osbuildManifest}
	assert.NotNil(t, mg)
	err := mg.Generate("test-distro-1", "test_type", "test_arch")
	assert.NoError(t, err)

	pipelineNames := pipelineNamesFrom(t, osbuildManifest.Bytes())
	assert.Equal(t, []string{"build", "os"}, pipelineNames)
}
