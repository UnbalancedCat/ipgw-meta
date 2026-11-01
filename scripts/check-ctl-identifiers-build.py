"""Check independent ctl-identifiers binding using self-tests and invalid fixtures only."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess

def main():
 p=argparse.ArgumentParser();p.add_argument('--directory',required=True);args=p.parse_args()
 directory=Path(args.directory).absolute();root=Path(__file__).resolve().parent.parent
 build=json.loads((directory/'build-receipt.json').read_text());digest=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
 for name,field in [('ctl-identifiers-launch.exe','launcher_sha256'),('payload/ctl-identifiers.exe','executable_sha256'),('payload/build-inputs.json','source_digest')]:assert digest(directory/name)==build[field]
 inputs=json.loads((directory/'payload/build-inputs.json').read_text())
 for item in inputs['sources']:
  assert digest(root/item['path'])==item['sha256']==digest(directory/'source'/item['path'])
  assert not item['path'].startswith(('internal/casacquire/','internal/cascollect/','internal/cmd/cas-acquire'))
 fixtures=directory/'audit-fixtures';fixtures.mkdir();checks=[]
 for case in ('valid','exe-tamper','inputs-tamper','missing','extra','wrong-identity'):
  fixture=fixtures/case;fixture.mkdir();shutil.copyfile(directory/'ctl-identifiers-launch.exe',fixture/'ctl-identifiers-launch.exe');shutil.copytree(directory/'payload',fixture/'payload')
  if case=='exe-tamper':
   with (fixture/'payload/ctl-identifiers.exe').open('ab') as f:f.write(b'synthetic-mutation')
  elif case=='inputs-tamper':(fixture/'payload/build-inputs.json').write_bytes(b'{}\n')
  elif case=='missing':(fixture/'payload/build-inputs.json').rename(fixture/'retained-inputs.json')
  elif case=='extra':(fixture/'payload/extra.json').write_bytes(b'{}\n')
  elif case=='wrong-identity':(fixture/'payload/ctl-identifiers.exe').rename(fixture/'payload/cas-acquire.exe')
  modes=('--self-test',) if case=='valid' else ('--self-test','--ctl-identifiers')
  for mode in modes:
   result=subprocess.run([str(fixture/'ctl-identifiers-launch.exe'),mode],stdin=subprocess.DEVNULL,capture_output=True,timeout=25)
   assert result.stderr==b'';receipt=json.loads(result.stdout)
   if case=='valid':
    assert result.returncode==0 and receipt['state']=='self_test_passed' and receipt['binding']=='verified'
    assert receipt['result']['mode']=='self-test' and all(v==0 for v in receipt['result']['operations'].values())
   else:assert result.returncode!=0 and receipt['state']=='binding_invalid' and receipt['result'] is None
   checks.append(dict(case=case,mode=mode,passed=True))
 for mode in ('--run-once','--anonymous-once-confirmed','--cache-check','--ctl-check'):
  result=subprocess.run([str(directory/'ctl-identifiers-launch.exe'),mode],stdin=subprocess.DEVNULL,capture_output=True,timeout=10)
  assert result.returncode!=0 and result.stderr==b'' and json.loads(result.stdout)['state']=='arguments_rejected'
  checks.append(dict(case='forbidden-mode',mode=mode,passed=True))
 report=dict(kind='ctl_identifiers_binding_checks',checks=checks,source_files=len(inputs['sources']),local_store_read=False,network_requests=0)
 (directory/'audit-receipt.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))
if __name__=='__main__':main()
