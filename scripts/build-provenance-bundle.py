#!/usr/bin/env python3
"""Build a new synthetic Windows development bundle from an explicit source snapshot."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import zipfile


def digest(data):
    return hashlib.sha256(data).hexdigest()


def canonical(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode("utf-8")


def record(name, data):
    return {"name": name, "size": len(data), "sha256": digest(data)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, help="New directory; existing destinations are refused")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    destination = Path(args.output).absolute()
    # This is a development packager, not an installer or arbitrary tree copier.
    allowed = (root / "build/dev-runs").resolve()
    if not destination.parent.resolve().is_relative_to(allowed) or destination.exists() or destination.is_symlink():
        parser.error("Use a fresh directory below build/dev-runs")
    go = shutil.which("go")
    if not go:
        parser.error("Pinned Go unavailable")
    env = os.environ.copy()
    retained = {key: env[key] for key in ("GOROOT", "GOCACHE", "GOMODCACHE", "GOPATH") if key in env}
    for key in list(env):
        if key.upper().startswith(("GO", "IPGW_")):
            env.pop(key)
    env.update(retained)
    env.update(GOTOOLCHAIN="local", GOENV="off", GOWORK="off", GOPROXY="off", GOSUMDB="off", GOFLAGS="-mod=readonly", CGO_ENABLED="0", GOOS="windows", GOARCH="amd64", GOAMD64="v1")
    version = subprocess.check_output([go, "version"], env=env, text=True).strip()
    if version not in ("go version go1.25.0 windows/amd64", "go version go1.25.0 linux/amd64"):
        parser.error("Exact local Go 1.25.0 required")
    selected = [root / "go.mod", root / "go.sum", Path(__file__).resolve()]
    for directory in ("internal/casprovenance", "internal/casbundle", "internal/cmd/cas-provenance", "internal/cmd/cas-provenance-launch"):
        selected.extend(path for path in (root / directory).glob("*.go") if not path.name.endswith("_test.go"))
    selected.sort(key=lambda path: path.relative_to(root).as_posix())
    source_bytes = []
    for path in selected:
        if path.is_symlink() or path.stat().st_size > 8 << 20 or getattr(path.lstat(), "st_file_attributes", 0) & 0x400:
            parser.error("Invalid source snapshot member")
        source_bytes.append((path.relative_to(root).as_posix(), path.read_bytes()))
    head = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
    dirty = bool(subprocess.check_output(["git", "status", "--porcelain", "--untracked-files=normal"], cwd=root))
    destination.mkdir()  # Exclusive creation: never replace an earlier bundle.
    snapshot = destination / "source-snapshot"
    snapshot.mkdir()
    for name, data in source_bytes:
        target = snapshot / name
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(data)
    files = [record(name, data) for name, data in source_bytes]
    inputs = {"schema_version": 1, "tool_kind": "cas-provenance-dev", "go_version": "go1.25.0", "platform": "windows/amd64", "cgo_enabled": "0", "head": head, "worktree": "dirty" if dirty else "clean", "source_digest": digest(canonical(files)), "sources": files, "build_flags": ["-trimpath", "-buildvcs=false", "-ldflags=-s -w"]}
    package = destination / "package"
    package.mkdir()
    trusted = destination / "reviewed-launcher"
    trusted.mkdir()
    commands = [[go, "mod", "verify"]]
    for binary, output in (("cas-provenance", package / "cas-provenance.exe"), ("cas-provenance-launch", trusted / "cas-provenance-launch.exe")):
        commands.append([go, "build", *inputs["build_flags"], "-o", str(output), "./internal/cmd/" + binary])
    for index, command in enumerate(commands):
        with (destination / ("build-" + str(index) + ".log")).open("wb") as log:
            subprocess.run(command, cwd=snapshot, env=env, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=180)
    # Detect edits to the staging snapshot itself during compilation.
    for name, data in source_bytes:
        if (snapshot / name).read_bytes() != data:
            raise RuntimeError("Source snapshot changed")
    input_data = canonical(inputs) + b"\n"
    (package / "build-inputs.json").write_bytes(input_data)
    exe = record("cas-provenance.exe", (package / "cas-provenance.exe").read_bytes())
    runner = {"schema_version": 1, "tool_kind": "cas-provenance-dev", "mode": "synthetic-only", "executable": exe["name"], "size": exe["size"], "sha256": exe["sha256"], "args": ["--self-test"], "timeout_seconds": 30, "max_output_bytes": 65536}
    (package / "runner.json").write_bytes(canonical(runner) + b"\n")
    manifest = {"schema_version": 1, "tool_kind": "cas-provenance-dev", "bundle_id": "dev-" + digest(input_data), "platform": "windows/amd64", "files": [record(name, (package / name).read_bytes()) for name in ("build-inputs.json", "cas-provenance.exe", "runner.json")]}
    manifest_data = canonical(manifest) + b"\n"
    (package / "manifest.json").write_bytes(manifest_data)
    archive = destination / "development-bundle.zip"
    with zipfile.ZipFile(archive, "x", compression=zipfile.ZIP_DEFLATED) as stream:
        for path in sorted(package.iterdir()):
            stream.write(path, path.name)
    receipt = {"kind": "local_development_build", "bundle_id": manifest["bundle_id"], "manifest_sha256": digest(manifest_data), "archive": record(archive.name, archive.read_bytes()), "launcher": record("cas-provenance-launch.exe", (trusted / "cas-provenance-launch.exe").read_bytes()), "source_digest": inputs["source_digest"], "go_version": version, "worktree": inputs["worktree"], "installation": "not_performed", "trust": "local_review_required_not_attestation"}
    (destination / "build-receipt.json").write_bytes(canonical(receipt) + b"\n")
    print(json.dumps(receipt, indent=2))


if __name__ == "__main__":
    main()
