package progress

import (
	"archive/tar"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/osbuild/image-builder-cli/pkg/util"
)

var superminInitScriptFmt = `#!/bin/sh
# inspired by https://github.com/coreos/coreos-assembler/blob/main/src/supermin-init-prelude.sh
# we need less because osbuild does most of its work via buildroots

set -e

export PATH=/usr/sbin:/usr/bin:/sbin:/bin

mount -t proc /proc /proc
mount -t sysfs /sys /sys
mount -t cgroup2 cgroup2 -o rw,nosuid,nodev,noexec,relatime,seclabel,nsdelegate,memory_recursiveprot /sys/fs/cgroup
mount -t devtmpfs devtmpfs /dev

# auto-reboot on kernel panic (e.g. if anything in this script dies)
echo 1 > /proc/sys/kernel/panic

# this is also normally set up by systemd in early boot
ln -s /proc/self/fd/0 /dev/stdin
ln -s /proc/self/fd/1 /dev/stdout
ln -s /proc/self/fd/2 /dev/stderr

# need /dev/shm for podman
mkdir -p /dev/shm
mount -t tmpfs tmpfs /dev/shm

# osbuild needs /run
mkdir -p /run
mount -t tmpfs tmpfs /run

### from here on we diverge from the above core-assembler script

# ensure network
/usr/sbin/dhclient eth0

# central for osbuild
mknod /dev/loop-control c 10 237

# map in the host /output dir, virtiofsd support uid remapping so this
# should be fine
mkdir -p /output
mount -t virtiofs osbuild_output /output

# We cannot put the host osbuild "store" on our /store dir even with
# virtiofsd as there will be issues with uid remapping and selinux
# labels.
mkdir -p /host/store
mount -t virtiofs osbuild_store /host/store
# rsync/cp the caches so that our build is faster
echo "Populate /store from host"
mkdir /store
rsync -a --exclude=./tmp /host/store/ /store

# fetch/cache sources first so that the host cache gets updated
echo "Fetching sources"
osbuild \
  --cache /store \
  /output/manifest.json
# rsync
rsync -a --exclude=./tmp /store/ /host/store

echo "Running osbuild"
osbuild \
  --export %s \
  --output-directory /output \
  --cache /store \
  /output/manifest.json

# trigger clean shutdown via sysreq
echo _suo > /proc/sysrq-trigger
# shutdown is async so we need to sleep here or PID=0 dies
sleep 999
`

func addInitTar(superminDir, superminInitScript string) error {
	initTarF, err := os.Create(filepath.Join(superminDir, "init.tgz"))
	if err != nil {
		return err
	}
	defer initTarF.Close()
	initTar := tar.NewWriter(initTarF)
	defer initTar.Close()
	if err := initTar.WriteHeader(&tar.Header{
		Name: "init",
		Size: int64(len(superminInitScript)),
		Mode: 0755,
	}); err != nil {
		return err
	}
	if _, err := initTar.Write([]byte(superminInitScript)); err != nil {
		return err
	}
	return nil
}

func superminPrepare(prepareDir, export string) error {
	superminInitScript := fmt.Sprintf(superminInitScriptFmt, export)

	// prepare supermin
	err := util.RunCmdSync(
		"supermin", "--prepare", "--use-installed",
		// fundamental
		"util-linux",
		// convenient
		"rsync",
		// basic networking
		"ca-certificates", "dhcp-client", "iproute",
		// loop-device support
		"kernel-modules",
		// osbuild and friends
		"osbuild", "osbuild-depsolve-dnf", "osbuild-lvm2", "osbuild-luks2", "osbuild-ostree",
		// lvm
		"lvm2",
		// target"
		"-o", prepareDir,
	)
	if err != nil {
		return fmt.Errorf("supermin prepare failed: %w", err)
	}
	if err := addInitTar(prepareDir, superminInitScript); err != nil {
		return err
	}
	return nil
}

func superminBuild(prepareDir, buildDir string) error {
	err := util.RunCmdSync(
		"supermin",
		"--build", prepareDir,
		// XXX: what is the right size?
		"--size", "40G",
		"-f", "ext2",
		"-o", buildDir,
	)
	if err != nil {
		return fmt.Errorf("supermin-build failed: %w", err)
	}
	return nil
}

func setupVirtiofsd(tmpDir, outputDir, storeDir string) (func(), error) {
	var cmds []*exec.Cmd
	var cleanupFunc = func() {
		for _, cmd := range cmds {
			if cmd != nil && cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
	}

	for _, mnt := range []struct {
		path, tag string
	}{
		{outputDir, "output"},
		{storeDir, "store"},
	} {
		socketPath := filepath.Join(tmpDir, fmt.Sprintf("vfsd_%s.sock", mnt.tag))
		var args []string
		// run virtiofsd in user namespace if non-root to make
		// chown() and friends inside the VM work
		if os.Getuid() != 0 {
			args = append(args, "podman", "unshare", "--")
		}
		// if this runs as root we will be inside a unprivileged
		// container so we need won't be able to do most of the
		// sandboxing (e.g. setfcap)
		args = append(args, "/usr/libexec/virtiofsd")
		args = append(args, []string{
			"--sandbox=none",
			"--seccomp=none",
			// workaround https://gitlab.com/virtio-fs/virtiofsd/-/merge_requests/197
			"--modcaps=-mknod:-setfcap",
			"--socket-path", socketPath,
			"--shared-dir", mnt.path,
		}...)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			cleanupFunc()
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	// XXX: wait for socket to appear
	time.Sleep(2 * time.Second)

	return cleanupFunc, nil
}

func runOSBuildWithSupermin(pbar ProgressBar, manifest []byte, exports []string, opts *OSBuildOptions) error {
	// XXX: add support for progress via e.g. a virtio serial port
	// that the osbuld output is piped to
	pbar.Stop()
	fmt.Fprintf(os.Stderr, "Running osbuild in supermin\n")

	if _, err := os.Stat("/dev/kvm"); err != nil {
		return fmt.Errorf("cannot use supermin without /dev/kvm: %w", err)
	}
	if len(exports) != 1 {
		return fmt.Errorf("only a single export supported right now")
	}

	superminTmp, err := os.MkdirTemp("/var/tmp", "supermin")
	if err != nil {
		return err
	}
	defer os.RemoveAll(superminTmp)

	// we need to prepare and then build supermin
	prepareDir := filepath.Join(superminTmp, "prepare")
	if err := superminPrepare(prepareDir, exports[0]); err != nil {
		return err
	}
	runDir := filepath.Join(superminTmp, "run")
	if err := superminBuild(prepareDir, runDir); err != nil {
		return err
	}

	// then we can run osbuild inside supermin
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return err
	}
	manifestPath := filepath.Join(opts.OutputDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifest, 0644); err != nil {
		return err
	}
	defer os.Remove(manifestPath)

	// map /output, /store into VM
	cleanup, err := setupVirtiofsd(superminTmp, opts.OutputDir, opts.StoreDir)
	if err != nil {
		return err
	}
	defer cleanup()

	return util.RunCmdSync(
		"qemu-kvm",
		"-nodefaults", "-nographic",
		"-accel", "kvm",
		"-cpu", "host",
		"-m", "4G",
		// exit on reboot, we need this to catch crashes in the init
		// shell script
		"-no-reboot",
		// XXX: see colins osbuildbootc/cosa for $arch specific setup
		// for qemu
		"-netdev", "user,id=eth0",
		"-device", "virtio-net-pci,netdev=eth0",
		"-object", "memory-backend-memfd,id=mem,size=4G,share=on",
		"-numa", "node,memdev=mem",
		// supermin generates those
		"-kernel", filepath.Join(runDir, "kernel"),
		"-initrd", filepath.Join(runDir, "initrd"),
		// virtiosfds stuff
		"-chardev", "socket,id=char0,path="+superminTmp+"/vfsd_output.sock",
		"-device", "vhost-user-fs-pci,queue-size=1024,chardev=char0,tag=osbuild_output",
		"-chardev", "socket,id=char1,path="+superminTmp+"/vfsd_store.sock",
		"-device", "vhost-user-fs-pci,queue-size=1024,chardev=char1,tag=osbuild_store",
		// XXX: see colins osbuildbootc/cosa for $arch options
		"-hda", filepath.Join(runDir, "root"),
		"-serial", "stdio",
		"-append", "console=ttyS0 quiet root=/dev/sda",
	)
}
