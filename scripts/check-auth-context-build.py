#!/usr/bin/env python3
"""Check an anonymous-context bundle using fake self-tests and tamper copies."""
import argparse
import hashlib
import json
import shutil
import subprocess
from pathlib import Path

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--directory',required=True)
    bundle=Path(parser.parse_args().directory).absolute()
    receipt=json.loads((bundle/'build-receipt.json').read_text())
    inputs=json.loads((bundle/'payload/build-inputs.json').read_text())
    assert receipt['schema_version']==inputs['schema_version']==1
    assert receipt['tool_kind']==inputs['tool_kind']=='auth-context-dev'
    expected={'auth-context-launch.exe':'launcher_sha256','payload/auth-context.exe':'observer_sha256','payload/ipgw-meta.exe':'product_sha256','payload/build-inputs.json':'source_digest'}
    for name,field in expected.items(): assert sha(bundle/name)==receipt[field]
    for item in inputs['sources']:
        p=bundle/'source'/item['path']
        assert p.stat().st_size==item['size'] and sha(p)==item['sha256']
    work=bundle.parent/'binding-audit'
    work.mkdir()
    results=[]
    for case in ('valid','observer-tamper','product-tamper','inputs-tamper','missing','extra','old-name','no-confirmation','password-rejected'):
        copy=work/case
        copy.mkdir()
        shutil.copy2(bundle/'auth-context-launch.exe',copy/'auth-context-launch.exe')
        shutil.copytree(bundle/'payload',copy/'payload')
        target={'observer-tamper':'auth-context.exe','product-tamper':'ipgw-meta.exe','inputs-tamper':'build-inputs.json'}.get(case)
        if target:
            with (copy/'payload'/target).open('ab') as f: f.write(b'synthetic-tamper')
        elif case=='missing': (copy/'payload/auth-context.exe').unlink()
        elif case=='extra': (copy/'payload/extra').write_bytes(b'synthetic')
        elif case=='old-name': (copy/'payload/auth-context.exe').rename(copy/'payload/auth-dev.exe')
        args=['--self-test']
        if case=='no-confirmation': args=['--anonymous-once']
        if case=='password-rejected': args=['--password-core','synthetic']
        run=subprocess.run([str(copy/'auth-context-launch.exe'),*args],input=b'',capture_output=True,timeout=40)
        r=json.loads(run.stdout)
        assert r['schema_version']==1 and r['tool_kind']=='auth-context-dev' and not r['live_attempted'] and not run.stderr
        if case=='valid':
            assert run.returncode==0 and r['binding']=='verified' and r['result']['state']=='pass' and r['result']['sdk']['application_requests']==0
        elif case=='no-confirmation': assert r['state']=='confirmation_required' and r['result'] is None and run.returncode!=0
        elif case=='password-rejected': assert r['state']=='arguments_rejected' and r['result'] is None and run.returncode!=0
        else: assert r['binding']=='invalid' and r['result'] is None and run.returncode!=0
        results.append(dict(case=case,passed=True))
    result=dict(tool_kind='auth-context-dev',source_digest=receipt['source_digest'],snapshot_verified=True,checks=results,live_executed=False)
    (bundle.parent/'binding-audit.json').write_text(json.dumps(result,indent=2)+'\n',encoding='utf-8')
    print(json.dumps(result,indent=2))
if __name__=='__main__': main()
