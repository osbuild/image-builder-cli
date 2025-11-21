import os
import subprocess
import shlex
from dataclasses import dataclass

import pytest
import yaml

# put common podman run args in once place
podman_run = ["podman", "run", "--rm", "--privileged"]


def test_smoke_has_expected_images_centos(build_container):
    """
    Ensure that image types that are built in by CentOS are available
    and do not disappear from the list. See:
    https://gitlab.com/redhat/centos-stream/release-engineering/releng-tools/-/blob/master/scripts/images-build-gen2.py
    """

    output = subprocess.check_output(podman_run + [
        build_container,
        "list",
    ], text=True)

    type_arch = {
        "tar": ["aarch64", "x86_64", "ppc64le", "s390x"],
        "qcow2": ["aarch64", "x86_64", "ppc64le", "s390x"],
        "ec2": ["x86_64", "aarch64"],
        "azure": ["x86_64", "aarch64"],
        "wsl": ["x86_64", "aarch64"],
        "vagrant-libvirt": ["x86_64"],
        "vagrant-virtualbox": ["x86_64"],
        "image-installer": ["x86_64", "aarch64"],
        "container": ["x86_64", "aarch64"],
        "container-minimal": ["x86_64", "aarch64"],
    }

    for distro in ["centos-9", "centos-10"]:
        for type_, arches in type_arch.items():
            for arch in arches:
                assert f"{distro} type:{type_} arch:{arch}" in output


def test_smoke_has_expected_images_almalinux_8(build_container):
    output = subprocess.check_output(podman_run + [
        build_container,
        "list",
    ], text=True)

    type_arch = {
        "ami": ["aarch64", "x86_64"],
        "ec2": ["aarch64", "x86_64"],
        "gce": ["x86_64"],
        "image-installer": ["aarch64", "x86_64"],
        "minimal-raw": ["aarch64", "x86_64"],
        "network-installer": ["aarch64", "x86_64"],
        "oci": ["x86_64"],
        "openstack": ["aarch64", "x86_64"],
        "ova": ["x86_64"],
        "qcow2": ["aarch64", "x86_64", "ppc64le", "s390x"],
        "tar": ["aarch64", "x86_64", "ppc64le", "s390x"],
        "vhd": ["aarch64", "x86_64"],
        "vmdk": ["x86_64"],
        "wsl": ["aarch64", "x86_64"],
    }

    versions = [
        "almalinux-8.4",
        "almalinux-8.5",
        "almalinux-8.6",
        "almalinux-8.7",
        "almalinux-8.8",
        "almalinux-8.9",
        "almalinux-8.10",
    ]

    expected_to_not_exist = [
        # alma linux did not have ppc64le before 8.5 and no s390x before 8.6, lets
        # skip testing their existence
        ("almalinux-8.4", "tar", "ppc64le"),
        ("almalinux-8.4", "tar", "s390x"),
        ("almalinux-8.4", "qcow2", "ppc64le"),
        ("almalinux-8.4", "qcow2", "s390x"),

        ("almalinux-8.5", "tar", "s390x"),
        ("almalinux-8.5", "qcow2", "s390x"),

        # there was no vhd for aarch64 before 8.6
        ("almalinux-8.4", "vhd", "aarch64"),
        ("almalinux-8.5", "vhd", "aarch64"),
    ]

    for distro in versions:
        for type_, arches in type_arch.items():
            for arch in arches:
                if (distro, type_, arch) in expected_to_not_exist:
                    continue
                assert f"{distro} type:{type_} arch:{arch}" in output


def test_smoke_has_expected_images_almalinux_9_and_10(build_container):
    output = subprocess.check_output(podman_run + [
        build_container,
        "list",
    ], text=True)

    type_arch = {
        "tar": ["aarch64", "x86_64", "ppc64le", "s390x"],
        "qcow2": ["aarch64", "x86_64", "ppc64le", "s390x"],
        "ec2": ["x86_64", "aarch64"],
        "azure": ["x86_64", "aarch64"],
        "wsl": ["x86_64", "aarch64"],
        "vagrant-libvirt": ["x86_64"],
        "vagrant-virtualbox": ["x86_64"],
        "image-installer": ["x86_64", "aarch64"],
        "container": ["x86_64", "aarch64"],
        "container-minimal": ["x86_64", "aarch64"],
    }

    versions = [
        "almalinux-9.0",
        "almalinux-9.1",
        "almalinux-9.2",
        "almalinux-9.3",
        "almalinux-9.4",
        "almalinux-9.5",
        "almalinux-9.6",
        "almalinux-9.7",
        "almalinux-9.8",

        "almalinux-10.0",
        "almalinux-10.1",
        "almalinux-10.2",

        "almalinux_kitten-10",
    ]

    expected_to_not_exist = [
    ]

    for distro in versions:
        for type_, arches in type_arch.items():
            for arch in arches:
                if (distro, type_, arch) in expected_to_not_exist:
                    continue
                assert f"{distro} type:{type_} arch:{arch}" in output


def test_smoke_has_expected_images_fedora(build_container):
    """
    Ensure that image types that are built in by Fedora are available
    and do not disappear from the list. See:
    https://pagure.io/pungi-fedora/blob/main/f/fedora.conf
    and
    https://pagure.io/fedora-iot/pungi-iot/blob/main/f/fedora-iot.conf
    """

    output = subprocess.check_output(podman_run + [
        build_container,
        "list",
    ], text=True)

    type_arch = {
        "minimal-raw-xz": ["aarch64"],
        "iot-raw-xz": ["x86_64", "aarch64"],
        "iot-installer": ["x86_64", "aarch64"],
        "iot-simplified-installer": ["x86_64", "aarch64"],
    }

    for distro in ["fedora-42", "fedora-43", "fedora-44"]:
        for type_, arches in type_arch.items():
            for arch in arches:
                assert f"{distro} type:{type_} arch:{arch}" in output


@pytest.mark.skipif(os.getuid() != 0, reason="needs root")
def test_smoke_version_smoke(build_container):
    output = subprocess.check_output(podman_run + [
        build_container,
        "--version",
    ])

    ver_yaml = yaml.load(output, yaml.loader.SafeLoader)

    assert ver_yaml["image-builder"]["version"] != ""
    assert ver_yaml["image-builder"]["commit"] != ""
    assert ver_yaml["image-builder"]["dependencies"]["images"] != ""
    assert ver_yaml["image-builder"]["dependencies"]["osbuild"] != ""


@dataclass
class ProgressTestCase:
    """Test case for progress output tests."""
    progress: str
    pty: bool
    needle: str
    forbidden: str


@pytest.mark.parametrize("case", [
    ProgressTestCase("verbose", True, "osbuild-stdout-output", "[|]"),
    ProgressTestCase("term", True, "[|]", "osbuild-stdout-output"),
    ProgressTestCase("verbose", False, "osbuild-stdout-output", "[|]"),
    ProgressTestCase("term", False, "[|]", "osbuild-stdout-output"),
])
@pytest.mark.skipif(os.getuid() != 0, reason="needs root")
def test_progress_smoke(tmp_path, build_fake_container, case: ProgressTestCase):
    output_dir = tmp_path / "output"
    output_dir.mkdir()

    podman_command = podman_run + [
        "-t" if case.pty else "-i",
        "-v", f"{output_dir}:/output",
        build_fake_container,
        "build",
        "qcow2",
        "--distro", "centos-9",
        "--output-dir=.",
        f"--progress={case.progress}",
    ]

    cast_filename = f"recording-{case.progress}.cast.json"
    asciinema_command = [
        "asciinema", "rec",
        "--quiet",
        "--overwrite",
        "--cols=80", "--rows=25",
        "--command", shlex.join(podman_command),
        cast_filename,
    ]

    if case.pty:
        result = subprocess.run(asciinema_command, text=True, check=False)
    else:
        result = subprocess.run(podman_command, text=True, check=False)
    assert result.returncode == 0, f"Podman with asciinema failed:\nSTDERR:\n{result.stderr}"

    assert os.path.exists(cast_filename)
    with open(cast_filename, "r", encoding="utf-8") as f:
        cast_text = f.read()

    assert case.needle in cast_text
    assert case.forbidden not in cast_text
