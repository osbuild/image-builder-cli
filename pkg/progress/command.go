package progress

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/osbuild/images/pkg/osbuild"
)

// XXX: merge variant back into images/pkg/osbuild/osbuild-exec.go
func RunOSBuild(pb ProgressBar, manifest []byte, exports []string, opts *osbuild.OSBuildOptions) error {
	if opts == nil {
		opts = &osbuild.OSBuildOptions{}
	}

	// To keep maximum compatibility keep the old behavior to run osbuild
	// directly and show all messages unless we have a "real" progress bar.
	//
	// This should ensure that e.g. "podman bootc" keeps working as it
	// is currently expecting the raw osbuild output. Once we double
	// checked with them we can remove the runOSBuildNoProgress() and
	// just run with the new runOSBuildWithProgress() helper.
	switch pb.(type) {
	case *terminalProgressBar, *debugProgressBar:
		return runOSBuildWithProgress(pb, manifest, exports, opts)
	default:
		return runOSBuildNoProgress(pb, manifest, exports, opts)
	}
}

func runOSBuildNoProgress(pb ProgressBar, manifest []byte, exports []string, opts *osbuild.OSBuildOptions) error {
	cmd := osbuild.NewOSBuildCmd(manifest, exports, opts)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running osbuild: %w", err)
	}
	return nil
}

func runOSBuildWithProgress(pb ProgressBar, manifest []byte, exports []string, opts *osbuild.OSBuildOptions) (err error) {
	rp, wp, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("cannot create pipe for osbuild: %w", err)
	}
	defer rp.Close()
	defer wp.Close()


	opts.Monitor = osbuild.MonitorJSONSeq
	opts.MonitorFD = 3
	opts.MonitorFile = wp
	var stdio bytes.Buffer
	var mu sync.Mutex
	buildLog := opts.BuildLog
	opts.BuildLogMu = &mu
	if opts.BuildLog == nil {
		mw := &stdio
		opts.Stdout = mw
		opts.Stderr = mw
		buildLog = io.Discard
	} else {
		opts.Stdout = &stdio
	}
	cmd := osbuild.NewOSBuildCmd(manifest, exports, opts)

	osbuildStatus := osbuild.NewStatusScanner(rp)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting osbuild: %v", err)
	}
	wp.Close()
	defer func() {
		// Try to stop osbuild if we exit early, we are gentle
		// here to give osbuild the chance to release its
		// resources (like mounts in the buildroot). This is
		// best effort only (but also a pretty uncommon error
		// condition). If ProcessState is set the process has
		// already exited and we have nothing to do.
		if err != nil && cmd.Process != nil && cmd.ProcessState == nil {
			sigErr := cmd.Process.Signal(syscall.SIGINT)
			err = errors.Join(err, sigErr)
		}
	}()

	var tracesMsgs []string
	for {
		st, err := osbuildStatus.Status()
		if err != nil {
			// This should never happen but if it does we try
			// to be helpful. We need to exit here (and kill
			// osbuild in the defer) or we would appear to be
			// handing as cmd.Wait() would wait to finish but
			// no progress or other message is reported. We
			// can also not (in the general case) recover as
			// the underlying osbuildStatus.scanner maybe in
			// an unrecoverable state (like ErrTooBig).
			return fmt.Errorf(`error parsing osbuild status, please report a bug and try with "--progress=verbose": %w`, err)
		}
		if st == nil {
			break
		}
		i := 0
		for p := st.Progress; p != nil; p = p.SubProgress {
			if err := pb.SetProgress(i, p.Message, p.Done, p.Total); err != nil {
				log.Printf("WARNING: cannot set progress: %v", err)
			}
			i++
		}
		// forward to user
		if st.Message != "" {
			pb.SetMessagef(st.Message)
		}

		// keep internal log for error reporting, forward to
		// external build log
		if st.Message != "" {
			tracesMsgs = append(tracesMsgs, st.Message)
			mu.Lock()
			fmt.Fprintln(buildLog, st.Message)
			mu.Unlock()
		}
		if st.Trace != "" {
			tracesMsgs = append(tracesMsgs, st.Trace)
			mu.Lock()
			fmt.Fprintln(buildLog, st.Trace)
			mu.Unlock()
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error running osbuild: %w\nBuildLog:\n%s\nOutput:\n%s", err, strings.Join(tracesMsgs, "\n"), stdio.String())
	}

	return nil
}
