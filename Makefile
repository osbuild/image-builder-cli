#
# Maintenance Helpers
#
# This makefile contains targets used for development, as well as helpers to
# aid automatization of maintenance. Unless a target is documented in
# `make help`, it is not supported and is only meant to be used by developers
# to aid their daily development work.
#
# All supported targets honor the `SRCDIR` variable to find the source-tree.
# For most unsupported targets, you are expected to have the source-tree as
# your working directory. To specify a different source-tree, simply override
# the variable via `SRCDIR=<path>` on the commandline. By default, the working
# directory is used for build output, but `BUILDDIR=<path>` allows overriding
# it.
#

BUILDDIR ?= .
SRCDIR ?= .

RST2MAN ?= rst2man

# see https://hub.docker.com/r/docker/golangci-lint/tags
# v1.55 to get golang 1.21 (1.21.3)
# v1.53 to get golang 1.20 (1.20.5)
GOLANGCI_LINT_VERSION=v1.55
GOLANGCI_LINT_CACHE_DIR=$(HOME)/.cache/golangci-lint/$(GOLANGCI_LINT_VERSION)
GOLANGCI_COMPOSER_IMAGE=composer_golangci
#
# Automatic Variables
#
# This section contains a bunch of automatic variables used all over the place.
# They mostly try to fetch information from the repository sources to avoid
# hard-coding them in this makefile.
#
# Most of the variables here are pre-fetched so they will only ever be
# evaluated once. This, however, means they are always executed regardless of
# which target is run.
#
#     VERSION:
#         This evaluates the `Version` field of the specfile. Therefore, it will
#         be set to the latest version number of this repository without any
#         prefix (just a plain number).
#
#     COMMIT:
#         This evaluates to the latest git commit sha. This will not work if
#         the source is not a git checkout. Hence, this variable is not
#         pre-fetched but evaluated at time of use.
#

VERSION := $(shell (cd "$(SRCDIR)" && grep "^Version:" image-builder-cli.spec | sed 's/[^[:digit:]]*\([[:digit:]]\+\).*/\1/'))
COMMIT = $(shell (cd "$(SRCDIR)" && git rev-parse HEAD))

#
# Generic Targets
#
# The following is a set of generic targets used across the makefile. The
# following targets are defined:
#
#     help
#         This target prints all supported targets. It is meant as
#         documentation of targets we support and might use outside of this
#         repository.
#         This is also the default target.
#
#     $(BUILDDIR)/
#     $(BUILDDIR)/%/
#         This target simply creates the specified directory. It is limited to
#         the build-dir as a safety measure. Note that this requires you to use
#         a trailing slash after the directory to not mix it up with regular
#         files. Lastly, you mostly want this as order-only dependency, since
#         timestamps on directories do not affect their content.
#

.PHONY: help
help:
	@echo "make [TARGETS...]"
	@echo
	@echo "This is the maintenance makefile of image-builder-cli. The following"
	@echo "targets are available:"
	@echo
	@echo "    help:               Print this usage information."
	@echo "    rpm:                Build the RPM"
	@echo "    srpm:               Build the source RPM"
	@echo "    scratch:            Quick scratch build of RPM"
	@echo "    clean:              Remove all built binaries"

$(BUILDDIR)/:
	mkdir -p "$@"

$(BUILDDIR)/%/:
	mkdir -p "$@"


#
# Maintenance Targets
#
# The following targets are meant for development and repository maintenance.
# They are not supported nor is their use recommended in scripts.
#

.PHONY: build
build: $(BUILDDIR)/bin/
	go build -o $<image-builder ./cmd/image-builder/

.PHONY: clean
clean:
	rm -rf $(BUILDDIR)/bin/
	rm -rf $(CURDIR)/rpmbuild
	rm -rf $(CURDIR)/release_artifacts

#
# Building packages
#
# The following rules build image-builder-cli packages from the current HEAD
# commit, based on the spec file in this directory.  The resulting packages
# have the commit hash in their version, so that they don't get overwritten
# when calling `make rpm` again after switching to another branch.
#
# All resulting files (spec files, source rpms, rpms) are written into
# ./rpmbuild, using rpmbuild's usual directory structure.
#

RPM_SPECFILE=rpmbuild/SPECS/image-builder-cli.spec
RPM_TARBALL=rpmbuild/SOURCES/image-builder-cli-$(COMMIT).tar.gz

.PHONY: $(RPM_SPECFILE)
$(RPM_SPECFILE):
	mkdir -p $(CURDIR)/rpmbuild/SPECS
	git show HEAD:image-builder-cli.spec > $(RPM_SPECFILE)
	./tools/rpm_spec_add_provides_bundle.sh $(RPM_SPECFILE)

RPM_TARBALL_UNCOMPRESSED=$(RPM_TARBALL:.tar.gz=.tar)

$(RPM_TARBALL):
	mkdir -p $(CURDIR)/rpmbuild/SOURCES
	git archive --prefix=image-builder-cli-$(COMMIT)/ --format=tar.gz HEAD > $(RPM_TARBALL)
	gunzip -f $(RPM_TARBALL)
	go mod vendor
	tar --append --owner=0 --group=0 --transform "s;^;image-builder-cli-$(COMMIT)/;" --file $(RPM_TARBALL_UNCOMPRESSED) vendor/
	gzip $(RPM_TARBALL_UNCOMPRESSED)

.PHONY: srpm
srpm: $(RPM_SPECFILE) $(RPM_TARBALL)
	rpmbuild -bs \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--define "commit $(COMMIT)" \
		--with tests \
		$(RPM_SPECFILE)

.PHONY: rpm
rpm: $(RPM_SPECFILE) $(RPM_TARBALL)
	rpmbuild -bb \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--define "commit $(COMMIT)" \
		--with tests \
		$(RPM_SPECFILE)

.PHONY: scratch
scratch: $(RPM_SPECFILE) $(RPM_TARBALL)
	rpmbuild -bb \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--define "commit $(COMMIT)" \
		--without tests \
		--nocheck \
		$(RPM_SPECFILE)

RPM_TARBALL_FILENAME=$(notdir $(RPM_TARBALL))

.PHONY: release_artifacts
release_artifacts: $(RPM_TARBALL)
	mkdir -p release_artifacts
	cp $< release_artifacts/
