package progress

import (
	"io"

	"github.com/osbuild/images/pkg/osbuild"
)

type (
	TerminalProgressBar = terminalProgressBar
	DebugProgressBar    = debugProgressBar
	VerboseProgressBar  = verboseProgressBar
)

var (
	OSStderr        = osStderr
)

func MockOsStderr(w io.Writer) (restore func()) {
	saved := osStderr
	osStderr = func() io.Writer { return w }
	return func() {
		osStderr = saved
	}
}

func MockIsattyIsTerminal(fn func(uintptr) bool) (restore func()) {
	saved := isattyIsTerminal
	isattyIsTerminal = fn
	return func() {
		isattyIsTerminal = saved
	}
}

func MockOsbuildCmd(s string) (restore func()) {
	saved := osbuild.OSBuildCmd
	osbuild.OSBuildCmd = s
	return func() {
		osbuild.OSBuildCmd = saved
	}
}
