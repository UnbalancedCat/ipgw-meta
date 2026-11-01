#!/usr/bin/env python3
"""Audit a locally reviewed acquisition build and run only synthetic checks."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--directory', required=True)
    args = parser.parse_args()
    directory = Path(args.directory).absolute()
    root = Path(__file__).resolve().parent.parent
    receipt = json.loads((directory / 'build-receipt.json').read_text(encoding='utf-8'))
    payload = directory / 'payload'
    assert sorted(p.name for p in payload.iterdir()) == ['build-inputs.json', 'cas-acquire.exe']
    assert digest(payload / 'cas-acquire.exe') == receipt['executable_sha256']
    assert digest(payload / 'build-inputs.json') == receipt['source_digest']
    assert digest(directory / 'cas-acquire-launch.exe') == receipt['launcher_sha256']
    inputs = json.loads((payload / 'build-inputs.json').read_text(encoding='utf-8'))
    assert len(inputs['sources']) == receipt['source_files']
    for source in inputs['sources']:
        assert digest(directory / 'source' / source['path']) == source['sha256']
        assert digest(root / source['path']) == source['sha256']
        assert (directory / 'source' / source['path']).stat().st_size == source['size']
    checks = []
    fixtures = directory / 'audit-fixtures'
    fixtures.mkdir()  # Never overwrite or delete any prior artifact.
    for case in ('valid', 'executable-tamper', 'inputs-tamper', 'missing', 'extra'):
        fixture = fixtures / case
        fixture.mkdir()
        shutil.copyfile(directory / 'cas-acquire-launch.exe', fixture / 'cas-acquire-launch.exe')
        shutil.copytree(payload, fixture / 'payload')
        if case == 'executable-tamper':
            with (fixture / 'payload/cas-acquire.exe').open('ab') as target:
                target.write(b'synthetic-mutation')
        elif case == 'inputs-tamper':
            (fixture / 'payload/build-inputs.json').write_bytes(b'{}\n')
        elif case == 'missing':
            # Move only this exact synthetic fixture member, preserving its bytes.
            (fixture / 'payload/build-inputs.json').rename(fixture / 'retained-inputs.json')
        elif case == 'extra':
            (fixture / 'payload/runner.json').write_bytes(b'{}\n')
        process = subprocess.run([str(fixture / 'cas-acquire-launch.exe'), '--self-test'], stdin=subprocess.DEVNULL, capture_output=True, timeout=60)
        assert process.stderr == b''
        result = json.loads(process.stdout)
        assert result['live_started'] is False
        if case == 'valid':
            assert process.returncode == 0 and result['state'] == 'self_test_passed'
            assert all(value == 0 for value in result['result']['attempts'].values())
            assert all(value == 0 for value in result['result']['operations'].values())
        else:
            assert process.returncode != 0 and result['state'] == 'binding_invalid'
            assert result['result'] is None
        checks.append(dict(case=case, passed=True, live_started=False))
        if case != 'valid':
            # Invalid bundles must reject the new local entry before any CA read.
            local = subprocess.run([str(fixture / 'cas-acquire-launch.exe'), '--cache-check'], stdin=subprocess.DEVNULL, capture_output=True, timeout=30)
            assert local.returncode != 0 and local.stderr == b''
            rejected = json.loads(local.stdout)
            assert rejected['state'] == 'binding_invalid' and rejected['result'] is None
            assert not rejected['live_started'] and not rejected['live_attempted']
            checks.append(dict(case=case + '-cache-check', passed=True, live_started=False))
    for member, expected in [('cas-acquire.exe', receipt['executable_sha256']), ('build-inputs.json', receipt['source_digest'])]:
        assert digest(payload / member) == expected
    report = dict(kind='acquisition_build_audit', source_files=receipt['source_files'], source_snapshot_match=True, original_payload_unchanged=True, checks=checks, live_executed=False)
    (directory / 'audit-receipt.json').write_text(json.dumps(report, indent=2) + '\n', encoding='utf-8')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()
