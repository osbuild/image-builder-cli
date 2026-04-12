package validation

import (
	"errors"

	"github.com/osbuild/blueprint/pkg/blueprint"
	ImageModel "github.com/osbuild/image-builder-cli/pkg/image_model"
)

func firewall(imageFormat ImageModel.CLIOutputFormat, config *blueprint.FirewallCustomization) []string {
	warnings := []string{}
	if imageFormat == ImageModel.FormatGCE || imageFormat == ImageModel.FormatOpenStack {
		message := "The Google and OpenStack templates explicitly disable the firewall for their environment. This cannot be overridden by the blueprint."
		warnings = append(warnings, message)
	}

	return warnings
}

// This would be more of a "policy"-based validation. Check very specific cases. Not sure if I love it, would be cool to figure out a more declarative way
func CustomizationConflicts(imageFormat ImageModel.CLIOutputFormat, config *blueprint.Customizations, failOnWarning bool) (error, []string) {
	warnings := []string{}
	if config.Firewall != nil {
		messages := firewall(imageFormat, config.Firewall)
		warnings = append(warnings, messages...)
		if failOnWarning && len(messages) > 0 {
			return errors.New(messages[0]), warnings
		}
	}

	return nil, warnings
}
