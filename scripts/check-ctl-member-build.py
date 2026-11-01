#!/usr/bin/env python3
"""Audit a locally built synthetic-only package; never invokes a real reader."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--directory', required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    directory = Path(args.directory).absolute()
    if os.name != 'nt' or root / 'build' not in directory.parents or directory.resolve() != directory:
        parser.error('Workspace Windows build required')
    read = lambda p: json.loads(p.read_text(encoding='utf-8'))
    sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
    receipt = read(directory / 'build-receipt.json')
    inputs = read(directory / 'build-inputs.json')
    assert receipt['tool_kind'] == inputs['tool_kind'] == 'ctl-member-synthetic'
    assert receipt['synthetic_only'] and not receipt['candidate']
    assert sha(directory / 'build-inputs.json') == receipt['source_digest']
    executable = directory / 'ctl-member.exe'
    assert sha(executable) == receipt['executable_sha256']
    for item in inputs['sources']:
        relative = Path(item['path'])
        assert not relative.is_absolute() and '..' not in relative.parts
        assert sha(directory / 'source' / relative) == sha(root / relative) == item['sha256']
        assert (directory / 'source' / relative).stat().st_size == item['size']
    rejected = [[], ['--ctl-check'], ['--ctl-identifiers'], ['--run-once'], ['--file', 'synthetic'], ['--self-test', 'extra']]
    for args in rejected:
        result = subprocess.run([str(executable), *args], capture_output=True, timeout=15)
        assert result.returncode == 2 and result.stdout == result.stderr == b''
    # The runner has already executed --self-test once; inspect that receipt,
    # rather than repeating the same process with identical inputs.
    result = read(directory.parent / 'member-packaged-self-test.log')
    assert set(result) == {'schema_version', 'tool_kind', 'mode', 'state', 'authentication', 'trust', 'checks', 'operations'}
    assert result['schema_version'] == 1 and result['tool_kind'] == 'ctl-member-synthetic'
    assert result['mode'] == 'self-test' and result['state'] == 'complete'
    assert result['authentication'] == 'blocked' and result['trust'] == 'not_established'
    assert result['checks'] == ['signed_member_match', 'signed_member_absent', 'opaque_policy_preserved', 'duplicate_rejected', 'tamper_rejected', 'both_encapsulations']
    assert set(result['operations']) == {'network_requests', 'dns', 'connects', 'profile_reads', 'credential_reads', 'certificate_writes', 'registry_writes'}
    assert all(type(v) is int and v == 0 for v in result['operations'].values())
    report = dict(kind='ctl_member_build_audit', source_digest=receipt['source_digest'], executable_sha256=receipt['executable_sha256'], snapshot_files=len(inputs['sources']), rejected_modes=len(rejected), self_test_complete=True, synthetic_only=True, passed=True)
    (directory / 'audit-receipt.json').write_text(json.dumps(report, indent=2) + '\n', encoding='utf-8')
    print(json.dumps(report))


if __name__ == '__main__':
    main()
