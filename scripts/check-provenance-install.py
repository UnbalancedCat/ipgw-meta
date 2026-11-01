#!/usr/bin/env python3
"""Exercise only a fresh locally built bundle in its synthetic report directory."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    directory = Path(args.directory).resolve()
    if os.name != "nt" or not directory.is_relative_to(root / "build/dev-runs/windows"):
        parser.error("Windows synthetic development report directory required")
    receipt = json.loads((directory / "build-receipt.json").read_text(encoding="utf-8"))
    archive = directory / "development-bundle.zip"
    launcher = directory / "reviewed-launcher/cas-provenance-launch.exe"
    destination = directory / "synthetic-install"
    if destination.exists() or destination.is_symlink():
        parser.error("Fresh synthetic destination required")
    if digest(archive) != receipt["archive"]["sha256"] or digest(launcher) != receipt["launcher"]["sha256"]:
        raise RuntimeError("Reviewed local build changed")
    command = [str(launcher), "--install-archive", str(archive), "--archive-sha256", receipt["archive"]["sha256"], "--manifest-sha256", receipt["manifest_sha256"], "--destination", str(destination)]

    def invoke(label, argv):
        process = subprocess.run(argv, capture_output=True, timeout=45)
        if process.stderr or len(process.stdout) > 131072:
            raise RuntimeError("Unexpected installer output")
        result = json.loads(process.stdout)
        (directory / (label + ".json")).write_bytes(process.stdout)
        if any(result["operations"].values()):
            raise RuntimeError("Unexpected operations")
        return process.returncode, result

    wrong = command.copy()
    wrong[4] = "0" * 64
    code, rejected = invoke("install-wrong-anchor", wrong)
    if code != 1 or rejected["state"] != "rejected" or rejected["directory_created"] or rejected["self_test"] is not None or destination.exists():
        raise RuntimeError("Wrong archive anchor crossed creation boundary")
    code, installed = invoke("install-positive", command)
    if code != 0 or installed["state"] != "installed" or installed["files_written"] != 4 or installed["bundle_id"] != receipt["bundle_id"]:
        raise RuntimeError("Fresh installation failed")
    chain = installed["self_test"]
    if chain["runner"] != "passed" or chain["exe_attempts"] != 1 or chain["exe_starts"] != 1 or chain["result"]["binding"] != "not_packaged":
        raise RuntimeError("Installed self-test contract failed")
    before = {p.name: digest(p) for p in destination.iterdir()}
    expected = {p.name: digest(p) for p in (directory / "package").iterdir()}
    if before != expected or len(before) != 4:
        raise RuntimeError("Installed bytes differ")
    code, rejected = invoke("install-existing-destination", command)
    after = {p.name: digest(p) for p in destination.iterdir()}
    if code != 1 or rejected["reason"] != "destination_invalid" or rejected["directory_created"] or rejected["self_test"] is not None or before != after:
        raise RuntimeError("Existing directory was not preserved")
    if digest(archive) != receipt["archive"]["sha256"]:
        raise RuntimeError("Source archive changed")
    print(json.dumps({"kind": "synthetic_development_install_check", "status": "passed", "checks": ["wrong_anchor_no_creation", "fresh_install_full_chain", "four_files_identical", "existing_destination_unchanged", "archive_unchanged"], "user_installation": "not_performed", "bundle_id": receipt["bundle_id"]}, indent=2))


if __name__ == "__main__":
    main()
