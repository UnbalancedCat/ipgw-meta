#!/usr/bin/env python3
"""Snapshot and build an independent Windows QR development bundle. No live run."""
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
    if os.name != 'nt' or output.exists() or root/'build' not in output.parents or output.resolve() != output:
        parser.error('New plain workspace build directory required')
    go = Path(os.environ['GOROOT'])/'bin/go.exe'
    env = os.environ.copy()
    env.update(GOTOOLCHAIN='local', GOENV='off', GOWORK='off', GOPROXY='off', GOSUMDB='off', GOFLAGS='-mod=readonly', GOOS='windows', GOARCH='amd64', CGO_ENABLED='0')
    def command(*argv):
        return subprocess.run([str(go), *argv], cwd=root, env=env, capture_output=True, check=True, text=True, encoding='utf-8').stdout
    version = command('version').strip()
    if version != 'go version go1.25.0 windows/amd64': parser.error('Exact Windows Go 1.25.0 required')
    command('mod','verify')
    targets = ['./internal/cmd/qr-dev','./internal/cmd/qr-dev-launch','./cmd/ipgw-meta']
    listing = command('list','-deps','-json',*targets)
    files = {root/'go.mod',root/'go.sum',Path(__file__).resolve(),root/'scripts/check-qr-dev-build.py'}
    decoder = json.JSONDecoder()
    while listing.strip():
        package, end = decoder.raw_decode(listing.lstrip()); listing = listing.lstrip()[end:]
        if package.get('Module',{}).get('Main'):
            directory = Path(package['Dir'])
            if directory != root and root not in directory.parents: raise RuntimeError('Source root escaped')
            files.update(directory.glob('*.go'))
            files.update(directory/name for name in package.get('EmbedFiles',[]))
    head = subprocess.run(['git','rev-parse','HEAD'],cwd=root,capture_output=True,check=True,text=True).stdout.strip()
    output.mkdir(parents=True)
    snapshot, payload = output/'source',output/'payload'; payload.mkdir()
    records = []
    for source in sorted(files):
        if source.is_symlink() or source.is_junction() or any(p.is_symlink() or p.is_junction() for p in source.parents): raise RuntimeError('Source link')
        relative=source.relative_to(root);data=source.read_bytes();target=snapshot/relative
        target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
        records.append(dict(path=relative.as_posix(),size=len(data),sha256=digest(data)))
    flags=['-trimpath','-buildvcs=false']
    inputs=dict(schema_version=1,tool_kind='qr-dev',source_state='uncommitted_snapshot',head_background=head,go_version=version,goos='windows',goarch='amd64',cgo_enabled=False,flags=flags,product_ldflags='-X main.version=dev-qr-local',sources=records)
    raw=(json.dumps(inputs,sort_keys=True,separators=(',',':'))+'\n').encode()
    (payload/'build-inputs.json').write_bytes(raw);inputs_hash=digest(raw)
    def build(target,destination,ldflags=''):
        subprocess.run([str(go),'build',*flags,'-ldflags='+ldflags,'-o',str(destination),target],cwd=snapshot,env=env,check=True)
    build(targets[0],payload/'qr-dev.exe')
    build(targets[2],payload/'ipgw-meta.exe',inputs['product_ldflags'])
    worker_hash=digest((payload/'qr-dev.exe').read_bytes());product_hash=digest((payload/'ipgw-meta.exe').read_bytes())
    build(targets[1],output/'qr-dev-launch.exe',f'-X main.workerDigest={worker_hash} -X main.inputsDigest={inputs_hash} -X main.productDigest={product_hash}')
    if not all(digest((snapshot/r['path']).read_bytes())==r['sha256'] for r in records): raise RuntimeError('Snapshot changed')
    receipt=dict(schema_version=1,tool_kind='qr-dev',source_digest=inputs_hash,worker_sha256=worker_hash,product_sha256=product_hash,launcher_sha256=digest((output/'qr-dev-launch.exe').read_bytes()),source_files=len(records),snapshot_verified=True,live_executed=False,release_evidence=False,candidate=False)
    (output/'build-receipt.json').write_text(json.dumps(receipt,indent=2)+'\n',encoding='utf-8')
    text=f"""# Single QR development window (separate authorization required)
Source: {inputs_hash}. Not Candidate or release evidence.
Confirm wired exclusive, TUN OFF, Wi-Fi disconnected and target account OFFLINE before live work.
Zero-network preflight:

    & '{output / 'qr-dev-launch.exe'}' --preflight

Only after ready and fresh single-run authorization, run in a private native terminal:

    & '{output / 'qr-dev-launch.exe'}' --run-once bhk-dev-prompt

Type RUN-DEV-QR-ONCE. Scan once using the target account. Do not capture the QR or raw terminal output.
The tool checks initial offline, QR, target status, logout and final offline. No automatic retry.
Overall 10 minutes; QR command 300 seconds; other commands 120 seconds. Topology change stops subsequent requests including cleanup.
Cleanup requires established ownership and a valid window. System DNS/PKI may access network or update caches.
After process exit restore original Clash configuration if needed. Hard interruption means account state unknown.
Return only the final sanitized JSON and exit code. Restored connectivity does not prove logout.
"""
    (output/'qr-window.md').write_text(text,encoding='utf-8')
    print(json.dumps(receipt,indent=2))

if __name__=='__main__': main()
