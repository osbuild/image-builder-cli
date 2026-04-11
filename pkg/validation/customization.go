package validation

import (
	"github.com/osbuild/blueprint/pkg/blueprint"
	ImageModel "github.com/osbuild/image-builder-cli/pkg/image_model"
)

// This would be more of a "policy"-based validation. Check very specific cases. Not sure if I love it, would be cool to figure out a more declarative way
func CustomizationConflicts(imageFormat ImageModel.CLIOutputFormat, config *blueprint.Customizations) ValidationResult {
	if imageFormat == ImageModel.FormatGCE || imageFormat == ImageModel.FormatOpenStack {
		if config.Firewall != nil {
			return ValidationResult{
				Ok: false,
				Level: LevelWarning,
				Message: "The Google and OpenStack templates explicitly disable the firewall for their environment. This cannot be overridden by the blueprint.",
			}
		}

	}

	return ValidationResult{Ok: true}
}
