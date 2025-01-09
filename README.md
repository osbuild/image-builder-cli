# image-builder CLI

Build images from the command line in a convenient way.

## Run via container

```console
$ sudo podman run --privileged \
   -v ./output:/output \
   ghcr.io/osbuild/image-builder-cli:latest \
   build \
   --distro fedora-41 \
   minimal-raw
```

## Installation

This project is under development right now and needs to be run via:
```console
$ go run github.com/osbuild/image-builder-cli/cmd/image-builder@main
```
or install it into `$GOPATH/bin`
```console
$ go install github.com/osbuild/image-builder-cli/cmd/image-builder@main
```

we plan to provide rpm packages as well.


## Prerequisites

Make sure to have the required `osbuild` RPMs installed:
```console
$ sudo dnf install osbuild osbuild-depsolve-dnf
```

## Examples

To see the list of buildable images run:
```console
$ image-builder list-images
...
centos-9 type:qcow2 arch:x86_64
...
rhel-10.0 type:ami arch:x86_64
...
```

To actually build an image run:
```console
$ sudo image-builder build qcow2 --distro centos-9
...
```
this will create a directory `centos-9-qcow2-x86_64` under which the
output is stored.


It is possible to filter:
```console
$ image-builder list-images --filter ami
...
centos-9 type:ami arch:x86_64
...
rhel-8.5 type:ami arch:aarch64
...
rhel-10.0 type:ami arch:aarch64
```
or be more specific
```console
$ image-builder list-images --filter "arch:x86*" --filter "distro:*centos*"
centos-9 type:ami arch:x86_64
...
centos-9 type:qcow2 arch:x86_64
...
```

The following filters are currently supported, shell-style globbing is supported:
 * distro: the distro name (e.g. fedora-41)
 * arch: the architecture name (e.g. x86_64)
 * type: the image type name (e.g. qcow2)
 * bootmode: the bootmode (legacy, UEFI, hybrid)

The output can also be switched, supported are "text", "json":
```console
$ image-builder list-images --output=json
[
  {
    "distro": {
      "name": "centos-9"
    },
    "arch": {
      "name": "aarch64"
    },
    "image_type": {
      "name": "ami"
    }
  },
...
  {
    "distro": {
      "name": "rhel-10.0"
    },
    "arch": {
      "name": "x86_64"
    },
    "image_type": {
      "name": "wsl"
    }
  }
]

```


## FAQ

Q: Does this require a backend.
A: The osbuild binary is used to actually build the images but beyond that
   no setup is required, i.e. no daemons like osbuild-composer.

Q: Can I have custom repository files?
A: Sure! The repostories are encoded in json in "<distro>-<vesion>.json",
   files, e.g. "fedora-41.json". See these [examples](https://github.com/osbuild/images/tree/main/data/repositories). Use the "--datadir" switch and
   place them under "repositories/name-version.json", e.g. for:
   "--datadir ~/my-project --distro foo-1" a json file must be put under 
   "~/my-project/repositories/foo-1.json.

## Project

 * **Website**: <https://www.osbuild.org>
 * **Bug Tracker**: <https://github.com/osbuild/image-builder-cli/issues>
 * **Discussions**: <https://github.com/orgs/osbuild/discussions>
 * **Matrix (chat)**: [Image Builder channel on Fedora Chat](https://matrix.to/#/#image-builder:fedoraproject.org?web-instance[element.io]=chat.fedoraproject.org)
 * **Changelog**: <https://github.com/osbuild/image-builder-cli/releases>

### Repository

 - **web**:   <https://github.com/osbuild/image-builder-cli>
 - **https**: `https://github.com/osbuild/image-builder-cli.git`
 - **ssh**:   `git@github.com:osbuild/image-builder-cli.git`
