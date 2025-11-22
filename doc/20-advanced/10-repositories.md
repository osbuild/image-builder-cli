# Repository Management

When building package based images `image-builder` will use repositories to get packages from. `image-builder` ships with built-in definitions and repositories for a [list of distributions](../10-faq.md#built-in-distributions). These are used when building artifacts.

A common request is to enable additional repositories, override the repositories used, redirect repositories, or include additional repositories in the produced artifact. For this we need to go through the way `image-builder` uses repositories for each step of the build process.
