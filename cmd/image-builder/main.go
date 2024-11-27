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
	osStdout io.Writer = os.Stdout
	osStderr io.Writer = os.Stderr
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

func cmdManifest(cmd *cobra.Command, args []string) error {
	dataDir, err := cmd.Flags().GetString("datadir")
	if err != nil {
		return err
	}

	distroName := args[0]
	imgType := args[1]
	var archStr string
	if len(args) > 2 {
		archStr = args[2]
	} else {
		archStr = arch.Current().String()
	}

	opts := &cmdlineOpts{
		out:     osStdout,
		dataDir: dataDir,
	}
	return outputManifest(distroName, imgType, archStr, opts)
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
	// XXX: add blueprint switch
	rootCmd.AddCommand(manifestCmd)

	return rootCmd.Execute()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(osStderr, "error: %s\n", err)
		os.Exit(1)
	}
}
