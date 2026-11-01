#!/usr/bin/env python3
"""Audit a new auth-dev bundle with self-tests and synthetic tamper copies only."""
import argparse
import hashlib
import json
import os
import shutil
import subprocess
from pathlib import Path


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--directory', required=True)
    args = parser.parse_args()
    bundle = Path(args.directory).absolute()
    receipt = json.loads((bundle / 'build-receipt.json').read_text())
    inputs = json.loads((bundle / 'payload/build-inputs.json').read_text())
    assert receipt['schema_version'] == inputs['schema_version'] == 3
    assert receipt['tool_kind'] == inputs['tool_kind'] == 'auth-dev-observed'
    expected = {'auth-dev-launch.exe': 'launcher_sha256', 'payload/auth-dev.exe': 'observer_sha256',
                'payload/ipgw-meta.exe': 'product_sha256', 'payload/build-inputs.json': 'source_digest', 'payload/auth-observed.exe': 'observed_cli_sha256'}
    for name, field in expected.items():
        assert sha(bundle / name) == receipt[field], 'Bundle hash mismatch'
    for record in inputs['sources']:
        path = bundle / 'source' / record['path']
        assert path.stat().st_size == record['size'] and sha(path) == record['sha256'], 'Snapshot mismatch'
    version = subprocess.run([str(bundle / 'payload/ipgw-meta.exe'), '--version'], capture_output=True, timeout=10, check=True)
    assert b'dev-auth-local' in version.stdout, 'Product lacks development identity'
    results = []
    work = bundle.parent / 'binding-audit'
    work.mkdir()
    # Real compiled observed entry, but a synthetic local profile only. No
    # login/status/topology command or native transport is invoked here.
    synthetic = work / 'synthetic-appdata'
    env = os.environ.copy()
    env['APPDATA'] = str(synthetic)
    config = synthetic / 'ipgw-meta/config.yaml'
    subprocess.run([str(bundle / 'payload/ipgw-meta.exe'), '--config', str(config), '--json', 'profile', 'add', 'synthetic', '--username', 'synthetic-account', '--credential-provider', 'prompt', '--switch', 'refuse'], env=env, capture_output=True, check=True, timeout=15)
    probe = subprocess.run([str(bundle / 'payload/auth-observed.exe'), '--internal-observed', '--json', '--profile', 'synthetic', '--bind-ip', '192.0.2.9', 'profile', 'show'], env=env, capture_output=True, check=True, timeout=15)
    child = json.loads(probe.stdout)
    assert child['tool_kind'] == 'auth-observed-cli' and child['exit_code'] == 0 and not probe.stderr
    assert child['observation']['requests'] == [] and child['observation']['sdk_category'] == 'not_called'
    results.append(dict(case='compiled-observed-synthetic-profile', passed=True))
    for case in ('valid', 'observer-tamper', 'product-tamper', 'observed-tamper', 'inputs-tamper', 'missing', 'extra'):
        copy = work / case
        copy.mkdir()
        shutil.copy2(bundle / 'auth-dev-launch.exe', copy / 'auth-dev-launch.exe')
        shutil.copytree(bundle / 'payload', copy / 'payload')
        target = {'observer-tamper': 'auth-dev.exe', 'observed-tamper': 'auth-observed.exe', 'product-tamper': 'ipgw-meta.exe', 'inputs-tamper': 'build-inputs.json'}.get(case)
        if target:
            with (copy / 'payload' / target).open('ab') as f:
                f.write(b'synthetic-tamper')
        elif case == 'missing':
            (copy / 'payload/auth-dev.exe').unlink()
        elif case == 'extra':
            (copy / 'payload/extra').write_bytes(b'synthetic')
        run = subprocess.run([str(copy / 'auth-dev-launch.exe'), '--self-test'], capture_output=True, timeout=40)
        report = json.loads(run.stdout)
        assert report['schema_version'] == 3 and report['tool_kind'] == 'auth-dev-observed'
        assert report['observations'] == dict(state='not_started', reason='none', commands=[])
        assert not report['live_attempted'], 'Unexpected live attempt'
        if case == 'valid':
            assert run.returncode == 0 and report['binding'] == 'verified' and report['result']['state'] == 'pass'
            assert not report['result']['release_evidence'] and report['result']['application_requests'] == 0
        else:
            assert run.returncode != 0 and report['binding'] == 'invalid' and report['result'] is None
        results.append({'case': case, 'passed': True})
    out = dict(tool_kind='auth-dev-observed', source_digest=receipt['source_digest'], snapshot_verified=True,
               product_development_identity=True, checks=results, live_executed=False)
    (bundle.parent / 'binding-audit.json').write_text(json.dumps(out, indent=2)+'\n', encoding='utf-8')
    print(json.dumps(out, indent=2))


if __name__ == '__main__':
    main()
