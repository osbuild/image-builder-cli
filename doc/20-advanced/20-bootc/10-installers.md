# Advanced `bootc` Topics

## Installer

Since there are no pre-provided installer container images at this moment, you can use a Containerfile similar to:

```Dockerfile
FROM your-favorite-bootc-container:latest
RUN dnf install -y \
     anaconda \
     anaconda-install-env-deps \
     anaconda-dracut \
     dracut-config-generic \
     dracut-network \
     net-tools \
     squashfs-tools \
     grub2-efi-x64-cdboot \
     python3-mako \
     lorax-templates-* \
     biosdevname \
     prefixdevname \
     && dnf clean all

# On Fedora 42 this is necessary to get files in the right places
# RUN dnf reinstall -y shim-x64

# On Fedora 43 and up this is necessary to get files in the right
# places
RUN mkdir -p /boot/efi && cp -ra /usr/lib/efi/*/*/EFI /boot/efi

# lorax wants to create a symlink in /mnt which points to /var/mnt
# on bootc but /var/mnt does not exist on some images.
#
# If https://gitlab.com/fedora/bootc/base-images/-/merge_requests/294
# gets merged this will be no longer needed
RUN mkdir /var/mnt
```

To produce your own Anaconda-based installer that can be used in combination with the `bootc-installer-payload-ref` argument like so:

```
$ sudo podman build -t localhost/anaconda -f Containerfile
$ sudo image-builder build --bootc-ref localhost/anaconda:latest --bootc-installer-payload-ref quay.io/centos-bootc/centos-bootc:stream10 bootc-installer
# ...
```
