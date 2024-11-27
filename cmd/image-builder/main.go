package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/osbuild/images/pkg/arch"
)

var (
	osStdin  io.ReadCloser = os.Stdin
	osStdout io.Writer     = os.Stdout
	osStderr io.Writer     = os.Stderr
)

type cmdlineOpts struct {
	dataDir string
	out     io.Writer
}

func cmdListImages(cmd *cobra.Command, args []string) error {
	filter, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return err
	}
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	dataDir, err := cmd.Flags().GetString("datadir")
	if err != nil {
		return err
	}

	opts := &cmdlineOpts{
		out:     osStdout,
		dataDir: dataDir,
	}
	return listImages(output, filter, opts)
}

func distroTypeArchFromArgs(args []string) (distroName, imgType, archStr string, err error) {
	distroName = args[0]
	imgType = args[1]
	switch {
	case len(args) == 2:
		archStr = arch.Current().String()
	case len(args) == 3:
		archStr = args[2]
	default:
		return "", "", "", fmt.Errorf("unexpected extra arguments: %q", args[2:])
	}
	return distroName, imgType, archStr, nil
}

func cmdManifest(cmd *cobra.Command, args []string) error {
	dataDir, err := cmd.Flags().GetString("datadir")
	if err != nil {
		return err
	}
	blueprintPath, err := cmd.Flags().GetString("blueprint")
	if err != nil {
		return err
	}

	distroName, imgType, archStr, err := distroTypeArchFromArgs(args)
	if err != nil {
		return err
	}

	opts := &cmdlineOpts{
		out:     osStdout,
		dataDir: dataDir,
	}
	return outputManifest(distroName, imgType, archStr, opts, blueprintPath)
}

func cmdBuild(cmd *cobra.Command, args []string) error {
	dataDir, err := cmd.Flags().GetString("datadir")
	if err != nil {
		return err
	}
	blueprintPath, err := cmd.Flags().GetString("blueprint")
	if err != nil {
		return err
	}

	distroName, imgType, archStr, err := distroTypeArchFromArgs(args)
	if err != nil {
		return err
	}

	opts := &cmdlineOpts{
		dataDir: dataDir,
	}
	return buildImage(distroName, imgType, archStr, opts, blueprintPath)
}

func run() error {
	// images logs a bunch of stuff to Debug/Info that is distracting
	// the user (at least by default, like what repos being loaded)
	logrus.SetLevel(logrus.WarnLevel)

	rootCmd := &cobra.Command{
		Use:   "image-builder",
		Short: "Build operating system images from a given distro/image-type/blueprint",
		Long: `Build operating system images from a given distribution,
image-type and blueprint.

Image-builder builds operating system images for a range of predefined
operating sytsems like centos and RHEL with easy customizations support.`,
		SilenceErrors: true,
	}
	rootCmd.PersistentFlags().String("datadir", "", `Override the default data direcotry for e.g. custom repositories/*.json data`)
	rootCmd.SetOut(osStdout)
	rootCmd.SetErr(osStderr)

	listImagesCmd := &cobra.Command{
		Use:          "list-images",
		Short:        "List buildable images, use --filter to limit further",
		RunE:         cmdListImages,
		SilenceUsage: true,
	}
	listImagesCmd.Flags().StringArray("filter", nil, `Filter distributions by a specific criteria (e.g. "type:rhel*")`)
	listImagesCmd.Flags().String("output", "", "Output in a specific format (text, json)")
	rootCmd.AddCommand(listImagesCmd)

	manifestCmd := &cobra.Command{
		Use:          "manifest <distro> <image-type> [<arch>]",
		Short:        "Build manifest for the given distro/image-type, e.g. centos-9 qcow2",
		RunE:         cmdManifest,
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(2),
		Hidden:       true,
	}
	// XXX: share with build
	manifestCmd.Flags().String("blueprint", "", `pass a blueprint file`)
	rootCmd.AddCommand(manifestCmd)

	buildCmd := &cobra.Command{
		Use:          "build <distro> <image-type>",
		Short:        "Build the given distro/image-type, e.g. centos-9 qcow2",
		RunE:         cmdBuild,
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(2),
	}
	rootCmd.AddCommand(buildCmd)
	// XXX: move to rootCmd
	buildCmd.Flags().String("datadir", "", `Override the default data direcotry for e.g. custom repositories/*.json data`)
	// XXX: share with manifest
	buildCmd.Flags().String("blueprint", "", `pass a blueprint file`)
	// XXX2: add --output=text,json and streaming

	return rootCmd.Execute()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(osStderr, "error: %s\n", err)
		os.Exit(1)
	}
}
