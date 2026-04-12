package validation

import (
	"errors"

	"github.com/osbuild/blueprint/pkg/blueprint"
	ImageModel "github.com/osbuild/image-builder-cli/pkg/image_model"
)

func firewall(imageFormat ImageModel.CLIOutputFormat, config *blueprint.FirewallCustomization) []string {
	warnings := []string{}
	if imageFormat == ImageModel.FormatGCE || imageFormat == ImageModel.FormatOpenStack {
		m := "The Google and OpenStack templates explicitly disable the firewall for their environment. This cannot be overridden by the blueprint."
		warnings = append(warnings, m)
	}

	return warnings
}

func filesystem(imageFormat ImageModel.CLIOutputFormat, config []blueprint.FilesystemCustomization) (error, []string) {
	warnings := []string{}
	if imageFormat == ImageModel.FormatImageInstaller || imageFormat == ImageModel.FormatEdgeInstaller || imageFormat == ImageModel.FormatEdgeSimplifiedInstaller {
		return errors.New("Filesystem customizations are currently not supported for the following image types: image-installer, edge-installer, edge-simplified-installer"), warnings
	}

	return nil, warnings
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

	if config.Filesystem != nil {
		if config.Disk != nil {
			m := "Filesystem and disk customizations are mutually exclusive."
			warnings = append(warnings, m)
			if failOnWarning {
				return errors.New(m), warnings
			}
		}
		err, messages := filesystem(imageFormat, config.Filesystem)
		warnings = append(warnings, messages...)
		if err != nil {
			return err, warnings
		}
		if failOnWarning && len(messages) > 0 {
			return errors.New(messages[0]), warnings
		}
	}

	return nil, warnings
}
