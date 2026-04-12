package validation

import (
	"fmt"
	ImageModel "github.com/osbuild/image-builder-cli/pkg/image_model"
)

// Maybe its better be a parse function (string, error)
func ValidateImageFormat(format string) error {
	for _, format := range ImageModel.AllCLIOutputFormats {
		if ImageModel.CLIOutputFormat(format) == format {
			return nil
		}
	}
	return fmt.Errorf("Invalid image format argument %s", format)
}
