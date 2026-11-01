#!/usr/bin/env python3
"""Build a separately bound Windows development observer and product; no live run."""
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
    def command(*argv):
        return subprocess.run([str(go), *argv], cwd=root, env=env, capture_output=True, check=True, text=True, encoding='utf-8').stdout
    version = command('version').strip()
    if version != 'go version go1.25.0 windows/amd64':
        parser.error('Exact local Windows Go 1.25.0 required')
    command('mod', 'verify')
    targets = ['./internal/cmd/auth-dev', './internal/cmd/auth-dev-launch', './cmd/ipgw-meta', './internal/cmd/auth-observed']
    # Resolve the actual local dependency closure, not a hand-maintained copy of
    # authentication packages. Include each package's tests in snapshot identity.
    listing = command('list', '-deps', '-json', *targets)
    decoder = json.JSONDecoder()
    files = {root / 'go.mod', root / 'go.sum', Path(__file__).resolve(), root / 'scripts/check-auth-dev-build.py'}
    while listing.strip():
        package, end = decoder.raw_decode(listing.lstrip())
        listing = listing.lstrip()[end:]
        if package.get('Module', {}).get('Main'):
            directory = Path(package['Dir'])
            if directory != root and root not in directory.parents:
                raise RuntimeError('Local package escaped source root')
            files.update(directory.glob('*.go'))
            for name in package.get('EmbedFiles', []):
                files.add(directory / name)
    head = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=root, capture_output=True, check=True, text=True).stdout.strip()
    output.mkdir(parents=True)
    snapshot, payload = output / 'source', output / 'payload'
    payload.mkdir()
    records = []
    for source in sorted(files):
        if source.is_symlink() or source.is_junction() or any(p.is_symlink() or p.is_junction() for p in source.parents):
            raise RuntimeError('Source link rejected')
        relative = source.relative_to(root)
        data = source.read_bytes()
        target = snapshot / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(data)
        records.append(dict(path=relative.as_posix(), size=len(data), sha256=digest(data)))
    flags = ['-trimpath', '-buildvcs=false']
    inputs = dict(schema_version=3, tool_kind='auth-dev-observed', source_state='uncommitted_snapshot', head_background=head,
                  go_version=version, goos='windows', goarch='amd64', cgo_enabled=False, flags=flags,
                  product_ldflags='-X main.version=dev-auth-local', sources=records)
    data = (json.dumps(inputs, sort_keys=True, separators=(',', ':')) + '\n').encode()
    (payload / 'build-inputs.json').write_bytes(data)
    inputs_hash = digest(data)
    def build(target, destination, ldflags=''):
        subprocess.run([str(go), 'build', *flags, '-ldflags='+ldflags, '-o', str(destination), target], cwd=snapshot, env=env, check=True)
    build(targets[0], payload / 'auth-dev.exe')
    build(targets[2], payload / 'ipgw-meta.exe', inputs['product_ldflags'])
    build(targets[3], payload / 'auth-observed.exe')
    observer_hash = digest((payload / 'auth-dev.exe').read_bytes())
    product_hash = digest((payload / 'ipgw-meta.exe').read_bytes())
    observed_hash = digest((payload / 'auth-observed.exe').read_bytes())
    build(targets[1], output / 'auth-dev-launch.exe', '-X main.observerDigest='+observer_hash+' -X main.inputsDigest='+inputs_hash+' -X main.productDigest='+product_hash+' -X main.observedDigest='+observed_hash)
    verified = all(digest((snapshot / r['path']).read_bytes()) == r['sha256'] for r in records)
    if not verified:
        raise RuntimeError('Snapshot changed while building')
    receipt = dict(schema_version=3, tool_kind='auth-dev-observed', source_digest=inputs_hash, observed_cli_sha256=observed_hash,
                   observer_sha256=observer_hash, product_sha256=product_hash,
                   launcher_sha256=digest((output / 'auth-dev-launch.exe').read_bytes()),
                   source_files=len(records), snapshot_verified=verified, live_executed=False, candidate=False)
    (output / 'build-receipt.json').write_text(json.dumps(receipt, indent=2)+'\n', encoding='utf-8')
    print(json.dumps(receipt, indent=2))


if __name__ == '__main__':
    main()
