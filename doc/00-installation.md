# Installation

As `image-builder-cli` is still in development its packages are not yet included in any distribution. There are however multiple ways that you can already use it. After you have `image-builder-cli` installed take a look at its [usage](./01-usage.md).

## Container

We build a container for `x86` and `aarch64` from our main branch; you can use it like this. We need to use a privileged container due to the way filesystems work in Linux. The below command will build a Fedora 41 Minimal Raw disk image and put it into the mounted output directory.

```
€ sudo podman run \
    --privileged \
    -v ./output:/output \
    ghcr.io/osbuild/image-builder-cli:latest \
    --distro fedora-41 \
    build minimal-raw
# ...
```

## RPM

If you're on a distribution that uses the RPM package format we also build RPMs from our main branch. These are hosted on COPR and you could use them:

```
€ sudo dnf copr enable @osbuild/image-builder-cli
# ...
€ sudo dnf install image-builder-cli
# ...
€ sudo image-builder build --distro fedora-41 minimal-raw
# ...
```

## Go

You can also run `image-builder-cli` directly with Go's package management.

_Doing this requires the pre-installation of `osbuild-composer` from your distributions package manager._

```
€ sudo go run 'github.com/osbuild/image-builder-cli/cmd/image-builder@main' build --distro fedora-41 minimal-raw
# ...
```

While not recommended you can also install `image-builder-cli` directly with Go:

```
€ sudo go install 'github.com/osbuild/image-builder-cli/cmd/image-builder'
# ...
€ sudo image-builder build --distro fedora-41 minimal-raw
# ...
```

## Source

Another option, and this might be most useful while hacking on the source is to run directly from a source checkout.

_Doing this requires the pre-installation of `osbuild-composer` from your distributions package manager._


```
€ git clone github.com/osbuild/image-builder-cli
# ...
€ sudo go run ./cmd/image-builder --distro fedora-41 build minimal-raw
# ...
```
