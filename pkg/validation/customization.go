package validation

import (
	"errors"

	"github.com/osbuild/blueprint/pkg/blueprint"
	ImageModel "github.com/osbuild/image-builder-cli/pkg/image_model"
)

// This would be more of a "policy"-based validation. Check very specific cases. Not sure if I love it, would be cool to figure out a more declarative way
func CustomizationConflicts(imageFormat ImageModel.CLIOutputFormat, config *blueprint.Customizations, failOnWarning bool) (error, []string) {
	warnings := []string{}
	if imageFormat == ImageModel.FormatGCE || imageFormat == ImageModel.FormatOpenStack {
		if config.Firewall != nil {
			message := "The Google and OpenStack templates explicitly disable the firewall for their environment. This cannot be overridden by the blueprint."
			if failOnWarning {
				return errors.New(message), warnings
			}
			warnings = append(warnings, message)
		}

	}

	return nil, warnings
}
