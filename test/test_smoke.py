import os
import subprocess

import pytest
import yaml


@pytest.mark.skipif(os.getuid() != 0, reason="needs root")
def test_manifest_version_smoke(build_container):
    output = subprocess.check_output([
        "podman", "run",
        "--privileged",
        build_container,
        "--version",
    ])

    ver_yaml = yaml.load(output, yaml.loader.SafeLoader)

    assert ver_yaml["image-builder"]["version"] != ""
    assert ver_yaml["image-builder"]["commit"] != ""
    assert ver_yaml["image-builder"]["dependencies"]["images"] != ""
