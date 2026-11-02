#!/usr/bin/env python3
"""Snapshot and build an independent Windows Wi-Fi development bundle. No live run."""
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
    targets = ['./internal/cmd/wifi-dev','./internal/cmd/wifi-dev-launch','./cmd/ipgw-meta']
    listing = command('list','-deps','-json',*targets)
    files = {root/'go.mod',root/'go.sum',Path(__file__).resolve(),root/'scripts/check-wifi-dev-build.py'}
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
    inputs=dict(schema_version=1,tool_kind='wifi-dev',source_state='uncommitted_snapshot',head_background=head,go_version=version,goos='windows',goarch='amd64',cgo_enabled=False,flags=flags,product_ldflags='-X main.version=dev-wifi-local',sources=records)
    raw=(json.dumps(inputs,sort_keys=True,separators=(',',':'))+'\n').encode()
    (payload/'build-inputs.json').write_bytes(raw);inputs_hash=digest(raw)
    def build(target,destination,ldflags=''):
        subprocess.run([str(go),'build',*flags,'-ldflags='+ldflags,'-o',str(destination),target],cwd=snapshot,env=env,check=True)
    build(targets[0],payload/'wifi-dev.exe')
    build(targets[2],payload/'ipgw-meta.exe',inputs['product_ldflags'])
    worker_hash=digest((payload/'wifi-dev.exe').read_bytes());product_hash=digest((payload/'ipgw-meta.exe').read_bytes())
    build(targets[1],output/'wifi-dev-launch.exe',f'-X main.workerDigest={worker_hash} -X main.inputsDigest={inputs_hash} -X main.productDigest={product_hash}')
    if not all(digest((snapshot/r['path']).read_bytes())==r['sha256'] for r in records): raise RuntimeError('Snapshot changed')
    receipt=dict(schema_version=1,tool_kind='wifi-dev',source_digest=inputs_hash,worker_sha256=worker_hash,product_sha256=product_hash,launcher_sha256=digest((output/'wifi-dev-launch.exe').read_bytes()),source_files=len(records),snapshot_verified=True,live_executed=False,release_evidence=False,candidate=False)
    (output/'build-receipt.json').write_text(json.dumps(receipt,indent=2)+'\n',encoding='utf-8')
    for label,ssid in [('NEU','neu'),('NEU-2.4G','neu_24g')]:
        text=f'''# {label} 独立开发窗口（尚须单次授权）

包源码身份：{inputs_hash}。非 Candidate，不生成发布 evidence。
先在有线管理连接可用时确认文件完整、私有 prompt profile 已就绪；不要在远程连接依赖该有线时直接断网。
获得本 SSID 单次授权后，维护者断开有线、关闭 TUN，连接 {label}，确认账号 offline 并独占网络，保持拓扑不变。
在原生私有 PowerShell 运行一次零网络预检：

    & '{output / 'wifi-dev-launch.exe'}' --preflight {ssid}

只有 state=ready 才运行一次辅助检查＋密码套件：

    & '{output / 'wifi-dev-launch.exe'}' --run-once {ssid} bhk-dev-prompt

输入显示的 RUN-WIFI-{ssid.upper()}-ONCE 确认；辅助全 pass 后才由普通产品提示密码。
密码仅留在终端。任一步 blocked/fail、权限错误、超时或未知结果即停止，不重试、不切换 SSID、不自行改参数或额外注销。
两项结束或停止并确认进程退出后恢复有线管理连接；需要时恢复原 Clash TUN，保留核心/代理/节点。
回传启动器最后一行脱敏 JSON 和退出码，不回传密码提示、二维码或原始产品输出。硬中断后账号状态未知，恢复联网不证明注销。
'''
        (output/f'{ssid}-window.md').write_text(text,encoding='utf-8')
    print(json.dumps(receipt,indent=2))

if __name__=='__main__': main()
