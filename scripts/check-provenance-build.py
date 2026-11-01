#!/usr/bin/env python3
"""Audit a newly created local development package; never execute its binaries."""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import subprocess
import sys
import zipfile


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    directory = Path(args.directory).resolve(strict=True)
    if not directory.is_relative_to((root / "build/dev-runs").resolve()):
        parser.error("Expected this repository's development output")
    receipt = json.loads((directory / "build-receipt.json").read_text(encoding="utf-8"))
    package = directory / "package"
    names = ["build-inputs.json", "cas-provenance.exe", "manifest.json", "runner.json"]
    if sorted(path.name for path in package.iterdir()) != names:
        raise RuntimeError("Package file set differs")
    before = {name: (package / name).read_bytes() for name in names}
    if digest(before["manifest.json"]) != receipt["manifest_sha256"]:
        raise RuntimeError("Manifest receipt differs")
    for filename, metadata in (("development-bundle.zip", receipt["archive"]), ("reviewed-launcher/cas-provenance-launch.exe", receipt["launcher"])):
        data = (directory / filename).read_bytes()
        if len(data) != metadata["size"] or digest(data) != metadata["sha256"]:
            raise RuntimeError("Build receipt differs")
    with zipfile.ZipFile(directory / "development-bundle.zip") as archive:
        if sorted(archive.namelist()) != names or len(archive.infolist()) != 4:
            raise RuntimeError("Archive member set differs")
        for member in archive.infolist():
            if member.file_size != len(before[member.filename]) or archive.read(member) != before[member.filename]:
                raise RuntimeError("Archive bytes differ")
    inputs = json.loads(before["build-inputs.json"])
    for source in inputs["sources"]:
        relative = PurePosixPath(source["name"])
        if relative.is_absolute() or ".." in relative.parts:
            raise RuntimeError("Source path rejected")
        path = directory / "source-snapshot" / relative
        data = path.read_bytes()
        if len(data) != source["size"] or digest(data) != source["sha256"]:
            raise RuntimeError("Source snapshot differs")
    # The repository packager must reject reuse before creating or overwriting
    # anything. No executable from the inspected directory is invoked here.
    retry = subprocess.run([sys.executable, str(root / "scripts/build-provenance-bundle.py"), "--output", str(directory)], capture_output=True, timeout=15)
    if retry.returncode != 2 or any((package / name).read_bytes() != data for name, data in before.items()):
        raise RuntimeError("Existing package was not preserved")
    print(json.dumps({"kind": "local_development_package_audit", "archive_members": 4, "archive_bytes_match": True, "source_files": len(inputs["sources"]), "source_bytes_match": True, "repeated_output_rejected": True, "existing_package_unchanged": True, "binaries_executed": 0}))


if __name__ == "__main__":
    main()
