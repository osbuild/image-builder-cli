package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"

	"github.com/osbuild/bootc-image-builder/bib/pkg/progress"
	"github.com/osbuild/images/pkg/cloud"
	"github.com/osbuild/images/pkg/cloud/awscloud"
)

// ErrMissingUploadConfig is returned when the upload configuration is missing
var ErrMissingUploadConfig = errors.New("missing upload configuration")

// ErrUploadConfigNotProvided is returned when all the upload configuration is missing
var ErrUploadConfigNotProvided = errors.New("missing all upload configuration")

// ErrUploadTypeUnsupported is returned when the upload type is not supported
var ErrUploadTypeUnsupported = errors.New("unsupported type")

var awscloudNewUploader = awscloud.NewUploader

func uploadImageWithProgress(uploader cloud.Uploader, imagePath string) error {
	f, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// setup basic progress
	st, err := f.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat upload: %v", err)
	}
	pbar := pb.New64(st.Size())
	pbar.Set(pb.Bytes, true)
	pbar.SetWriter(osStdout)
	r := pbar.NewProxyReader(f)
	pbar.Start()
	defer pbar.Finish()

	return uploader.UploadAndRegister(r, osStderr)
}

func uploaderCheckWithProgress(pbar progress.ProgressBar, uploader cloud.Uploader) error {
	pr, pw := io.Pipe()
	defer pw.Close()

	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			pbar.SetMessagef("%s", scanner.Text())
		}
	}()
	return uploader.Check(pw)
}

func uploaderFor(cmd *cobra.Command, typeOrCloud string) (cloud.Uploader, error) {
	switch typeOrCloud {
	case "ami", "aws":
		return uploaderForCmdAWS(cmd)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUploadTypeUnsupported, typeOrCloud)
	}

}

func uploaderForCmdAWS(cmd *cobra.Command) (cloud.Uploader, error) {
	amiName, err := cmd.Flags().GetString("aws-ami-name")
	if err != nil {
		return nil, err
	}
	bucketName, err := cmd.Flags().GetString("aws-bucket")
	if err != nil {
		return nil, err
	}
	region, err := cmd.Flags().GetString("aws-region")
	if err != nil {
		return nil, err
	}

	var missing []string
	requiredArgs := []string{"aws-ami-name", "aws-bucket", "aws-region"}
	for _, argName := range requiredArgs {
		arg, err := cmd.Flags().GetString(argName)
		if err != nil {
			return nil, err
		}
		if arg == "" {
			missing = append(missing, fmt.Sprintf("--%s", argName))
		}
	}
	if len(missing) > 0 {
		if len(missing) == len(requiredArgs) {
			return nil, fmt.Errorf("%w: %q", ErrUploadConfigNotProvided, missing)
		}

		return nil, fmt.Errorf("%w: %q", ErrMissingUploadConfig, missing)
	}

	return awscloudNewUploader(region, bucketName, amiName, nil)
}

func cmdUpload(cmd *cobra.Command, args []string) error {
	uploadTo, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}
	if uploadTo == "" {
		return fmt.Errorf("missing --to parameter, try --to=aws")
	}

	imagePath := args[0]
	uploader, err := uploaderFor(cmd, uploadTo)
	if err != nil {
		return err
	}

	return uploadImageWithProgress(uploader, imagePath)
}
