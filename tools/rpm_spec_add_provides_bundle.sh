#!/usr/bin/bash

SPEC_FILE=${1:-"image-builder.spec"}

# Save the list of bundled packages into a file
WORKDIR=$(mktemp -d)
BUNDLES_FILE=${WORKDIR}/bundles.txt
./tools/rpm_spec_vendor2provides vendor/modules.txt > "${BUNDLES_FILE}"

# Remove the current bundle lines
sed -i '/^# BUNDLE_START/,/^# BUNDLE_END/{//p;d;}' "${SPEC_FILE}"
# Add the new bundle lines
sed -i "/^# BUNDLE_START/r ${BUNDLES_FILE}" "${SPEC_FILE}"
