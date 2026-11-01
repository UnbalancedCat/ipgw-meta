#!/usr/bin/env python3
"""Build an independent local-only Windows CTL identifier snapshot."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    output = Path(args.output).absolute()
    if os.name != 'nt' or output.exists() or root / 'build' not in output.parents or output.resolve() != output:
        parser.error('A new workspace build directory on Windows is required')
    go = Path(os.environ['GOROOT']) / 'bin/go.exe'
    env = os.environ.copy()
    env.update(GOTOOLCHAIN='local', GOENV='off', GOPROXY='off', GOSUMDB='off', GOFLAGS='-mod=readonly', GOOS='windows', GOARCH='amd64', CGO_ENABLED='0')
    version = subprocess.run([str(go), 'version'], env=env, capture_output=True, check=True, text=True).stdout.strip()
    if version != 'go version go1.25.0 windows/amd64':
        parser.error('Exact local Windows Go 1.25.0 required')
    subprocess.run([str(go), 'mod', 'verify'], cwd=root, env=env, check=True)
    head = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=root, capture_output=True, check=True, text=True).stdout.strip()
    files = [root / 'go.mod', root / 'go.sum', Path(__file__).resolve()]
    for package in ('ctlcheck', 'casprovenance', 'casbundle', 'cmd/ctl-identifiers', 'cmd/ctl-identifiers-launch'):
        files += [p for p in (root / 'internal' / package).glob('*.go') if not p.name.endswith('_test.go')]
    output.mkdir(parents=True)
    snapshot = output / 'source'
    payload = output / 'payload'
    payload.mkdir()
    records = []
    for source in sorted(files):
        relative = source.relative_to(root)
        if source.is_symlink() or source.is_junction() or any(parent.is_symlink() or parent.is_junction() for parent in source.parents):
            raise RuntimeError('Source link rejected')
        data = source.read_bytes()
        target = snapshot / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(data)
        records.append(dict(path=relative.as_posix(), size=len(data), sha256=digest(data)))
    flags = ['-trimpath', '-buildvcs=false']
    inputs = dict(schema_version=1, tool_kind='ctl-identifiers-dev', head_background=head, source_state='uncommitted_snapshot', go_version=version, goos='windows', goarch='amd64', cgo_enabled=False, flags=flags, sources=records)
    data = (json.dumps(inputs, sort_keys=True, separators=(',', ':')) + '\n').encode()
    (payload / 'build-inputs.json').write_bytes(data)
    inputs_hash = digest(data)
    subprocess.run([str(go), 'build', *flags, '-o', str(payload / 'ctl-identifiers.exe'), './internal/cmd/ctl-identifiers'], cwd=snapshot, env=env, check=True)
    exe_hash = digest((payload / 'ctl-identifiers.exe').read_bytes())
    ldflags = '-X main.executableDigest=' + exe_hash + ' -X main.inputsDigest=' + inputs_hash
    subprocess.run([str(go), 'build', *flags, '-ldflags=' + ldflags, '-o', str(output / 'ctl-identifiers-launch.exe'), './internal/cmd/ctl-identifiers-launch'], cwd=snapshot, env=env, check=True)
    receipt = dict(schema_version=1, tool_kind='ctl-identifiers-dev', source_digest=inputs_hash, executable_sha256=exe_hash, launcher_sha256=digest((output / 'ctl-identifiers-launch.exe').read_bytes()), source_files=len(records), snapshot_verified=all(digest((snapshot / r['path']).read_bytes()) == r['sha256'] for r in records), modules_verified=True, live_executed=False, candidate=False)
    (output / 'build-receipt.json').write_text(json.dumps(receipt, indent=2) + '\n', encoding='utf-8')
    print(json.dumps(receipt, indent=2))


if __name__ == '__main__':
    main()
