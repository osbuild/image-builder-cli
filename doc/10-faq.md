# Frequently Asked Questions

As we receive questions we'll fill in the frequent ones here.

## Why does `image-builder-cli` need `root` permissions?

For image types where we need to work with filesystems we need root. Mounting and working with filesystems is not namespaced in the Linux kernel and mounting filesystems is generally considered to be "running untrusted code in the kernel" hence it requires root permissions.
