import random
import string
import subprocess
import textwrap

import pytest


# XXX: copied from bib
@pytest.fixture(name="build_container", scope="session")
def build_container_fixture():
    """Build a container from the Containerfile and returns the name"""

    container_tag = "ibcli-test-" + "".join(random.choices(
        string.ascii_lowercase + string.digits, k=4))

    subprocess.check_call([
        "podman", "build",
        "-f", "Containerfile",
        "-t", container_tag,
    ])
    yield container_tag
    subprocess.check_call(["podman", "rmi", container_tag])


# XXX: copied from bib
@pytest.fixture(name="build_fake_container", scope="session")
def build_fake_container_fixture(tmpdir_factory, build_container):
    """Build a container with a fake osbuild and returns the name"""
    tmp_path = tmpdir_factory.mktemp("build-fake-container")

    fake_osbuild_path = tmp_path / "fake-osbuild"
    fake_osbuild_path.write_text(textwrap.dedent("""\
    #!/bin/bash -e

    # injest generated manifest from the images library, if we do not
    # do this images may fail with "broken" pipe errors
    cat - >/dev/null

    echo "osbuild-stdout-output"
    mkdir -p /output/qcow2
    echo "fake-disk.qcow2" > /output/qcow2/disk.qcow2

    """), encoding="utf8")

    cntf_path = tmp_path / "Containerfile"

    cntf_path.write_text(textwrap.dedent(f"""\n
    FROM {build_container}
    COPY fake-osbuild /usr/bin/osbuild
    RUN chmod 755 /usr/bin/osbuild
    """), encoding="utf8")

    container_tag = "ibcli-test-faked-osbuild-" + "".join(random.choices(
        string.ascii_lowercase + string.digits, k=4))
    subprocess.check_call([
        "podman", "build",
        "-t", container_tag,
        tmp_path,
    ])
    yield container_tag
    subprocess.check_call(["podman", "rmi", container_tag])
