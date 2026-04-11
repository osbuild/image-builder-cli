package imagemodel

type CLIOutputFormat string

const (
	FormatQCOW2                   CLIOutputFormat = "qcow2"
	FormatTar                     CLIOutputFormat = "tar"
	FormatAMI                     CLIOutputFormat = "ami"
	FormatVHD                     CLIOutputFormat = "vhd"
	FormatGCE                     CLIOutputFormat = "gce"
	FormatVMDK                    CLIOutputFormat = "vmdk"
	FormatOVA                     CLIOutputFormat = "ova"
	FormatOpenStack               CLIOutputFormat = "openstack"
	FormatEdgeCommit              CLIOutputFormat = "edge-commit"
	FormatIoTCommit               CLIOutputFormat = "iot-commit"
	FormatEdgeContainer           CLIOutputFormat = "edge-container"
	FormatIoTContainer            CLIOutputFormat = "iot-container"
	FormatEdgeInstaller           CLIOutputFormat = "edge-installer"
	FormatIoTInstaller            CLIOutputFormat = "iot-installer"
	FormatEdgeRawImage            CLIOutputFormat = "edge-raw-image"
	FormatIoTRawImage             CLIOutputFormat = "iot-raw-image"
	FormatEdgeSimplifiedInstaller CLIOutputFormat = "edge-simplified-installer"
	FormatIoTSimplifiedInstaller  CLIOutputFormat = "iot-simplified-installer"
	FormatEdgeAMI                 CLIOutputFormat = "edge-ami"
	FormatIoTAMI                  CLIOutputFormat = "iot-ami"
	FormatEdgeVSphere             CLIOutputFormat = "edge-vsphere"
	FormatImageInstaller          CLIOutputFormat = "image-installer"
	FormatLiveInstaller           CLIOutputFormat = "live-installer"
	FormatOCI                     CLIOutputFormat = "oci"
)

var AllCLIOutputFormats = []CLIOutputFormat{
	FormatQCOW2,
	FormatTar,
	FormatAMI,
	FormatVHD,
	FormatGCE,
	FormatVMDK,
	FormatOVA,
	FormatOpenStack,
	FormatEdgeCommit,
	FormatIoTCommit,
	FormatEdgeContainer,
	FormatIoTContainer,
	FormatEdgeInstaller,
	FormatIoTInstaller,
	FormatEdgeRawImage,
	FormatIoTRawImage,
	FormatEdgeSimplifiedInstaller,
	FormatIoTSimplifiedInstaller,
	FormatEdgeAMI,
	FormatIoTAMI,
	FormatEdgeVSphere,
	FormatImageInstaller,
	FormatLiveInstaller,
	FormatOCI,
}
