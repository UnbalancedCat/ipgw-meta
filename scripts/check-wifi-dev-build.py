#!/usr/bin/env python3
"""Check snapshot, executable binding and fail-closed synthetic copies. No campus access."""
import argparse
import hashlib
import json
import shutil
import subprocess
from pathlib import Path

def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--directory',required=True)
    bundle=Path(parser.parse_args().directory).absolute()
    receipt=json.loads((bundle/'build-receipt.json').read_text())
    inputs=json.loads((bundle/'payload/build-inputs.json').read_text())
    assert receipt['schema_version']==inputs['schema_version']==1
    assert receipt['tool_kind']==inputs['tool_kind']=='wifi-dev'
    expected={'wifi-dev-launch.exe':'launcher_sha256','payload/wifi-dev.exe':'worker_sha256','payload/ipgw-meta.exe':'product_sha256','payload/build-inputs.json':'source_digest'}
    for name,field in expected.items(): assert sha(bundle/name)==receipt[field]
    for record in inputs['sources']:
        relative=Path(record['path']);assert not relative.is_absolute() and '..' not in relative.parts
        path=bundle/'source'/relative
        assert path.stat().st_size==record['size'] and sha(path)==record['sha256']
    version=subprocess.run([str(bundle/'payload/ipgw-meta.exe'),'--version'],capture_output=True,check=True,timeout=10)
    assert b'dev-wifi-local' in version.stdout
    work=bundle.parent/'binding-audit';work.mkdir()
    checks=[]
    for case in ('valid','worker-tamper','product-tamper','inputs-tamper','missing','extra'):
        target=work/case;target.mkdir()
        shutil.copy2(bundle/'wifi-dev-launch.exe',target/'wifi-dev-launch.exe');shutil.copytree(bundle/'payload',target/'payload')
        name={'worker-tamper':'wifi-dev.exe','product-tamper':'ipgw-meta.exe','inputs-tamper':'build-inputs.json'}.get(case)
        if name:
            with (target/'payload'/name).open('ab') as stream: stream.write(b'synthetic-tamper')
        elif case=='missing': (target/'payload/wifi-dev.exe').unlink()
        elif case=='extra': (target/'payload/extra').write_bytes(b'synthetic')
        run=subprocess.run([str(target/'wifi-dev-launch.exe'),'--self-test'],capture_output=True,timeout=55)
        r=json.loads(run.stdout)
        assert r['tool_kind']=='wifi-dev' and not r['release_evidence'] and not r['live_attempted']
        assert r['password'] is None and r['auxiliary_checks']==[] and r['preflight'] is None
        if case=='valid': assert run.returncode==0 and r['binding']=='verified' and r['state']=='pass'
        else: assert run.returncode!=0 and r['binding']=='invalid' and r['reason']=='binding_rejected'
        checks.append(dict(case=case,passed=True))
    report=dict(tool_kind='wifi-dev',source_digest=receipt['source_digest'],snapshot_verified=True,checks=checks,live_executed=False)
    (bundle.parent/'binding-audit.json').write_text(json.dumps(report,indent=2)+'\n',encoding='utf-8')
    print(json.dumps(report,indent=2))

if __name__=='__main__':main()
