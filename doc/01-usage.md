# Usage

After [installing](./00-installation.md) you probably want to use `image-builder-cli`. A general workflow would be to find the image type you want to build and then build it.

Let's take a look at the available `x86_64` image types for Fedora 41 and build one.

```
€ image-builder list-images --filter arch:x86_64 --filter distro:fedora-41
fedora-41 type:ami arch:x86_64
fedora-41 type:container arch:x86_64
fedora-41 type:image-installer arch:x86_64
fedora-41 type:iot-bootable-container arch:x86_64
fedora-41 type:iot-commit arch:x86_64
fedora-41 type:iot-container arch:x86_64
fedora-41 type:iot-installer arch:x86_64
fedora-41 type:iot-qcow2-image arch:x86_64
fedora-41 type:iot-raw-image arch:x86_64
fedora-41 type:iot-simplified-installer arch:x86_64
fedora-41 type:live-installer arch:x86_64
fedora-41 type:minimal-raw arch:x86_64
fedora-41 type:oci arch:x86_64
fedora-41 type:openstack arch:x86_64
fedora-41 type:ova arch:x86_64
fedora-41 type:qcow2 arch:x86_64
fedora-41 type:vhd arch:x86_64
fedora-41 type:vmdk arch:x86_64
fedora-41 type:wsl arch:x86_64
€ image-builder build --distro fedora-41 qcow2
# ...
```

# Blueprints

Images can be customized with [blueprints](https://osbuild.org/docs/user-guide/blueprint-reference). For example we can build the `qcow2` we built above with some customizations applied.

We'll be adding the `nginx`, and `haproxy` packages and enabling their services so they start on boot. We'll also add a user by the name `user` with an ssh key and set the hostname of the machine:

```
€ cat blueprint.toml 
packages = [
    { name = "nginx" },
    { name = "haproxy" },
]

[customizations]
hostname = "mynewmachine.home.arpa"

[customizations.services]
enabled = ["nginx", "haproxy"]

[[customizations.user]]
name = "user"
key = "REDACTED"
€ image-builder build --distro fedora-41 qcow2 blueprint.toml
# ...
```

_Note that blueprints are aimed at end-user or systems administrator customizations. They take care to not break image builds and due to that they are not the right abstraction level for those who want to build or maintain entire distributions. For those interested in that usecase please take a look at the roadmap._
