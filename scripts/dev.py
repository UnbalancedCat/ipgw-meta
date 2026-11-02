#!/usr/bin/env python3
"""Run bounded, synthetic development checks with a pinned local toolchain."""
import argparse
import datetime
import json
import os
from pathlib import Path
import shutil
import subprocess
import uuid


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("stage", choices=("qr-dev", "wifi-dev", "p4", "logout-structure", "key-relation", "focus", "install", "test", "race", "vet", "docs", "quality", "build", "cross", "cas-observe", "cas-provenance", "cas-binding", "cas-install", "cas-collect", "cas-acquire", "trust-audit", "ctl-check", "ctl-identifiers", "ctl-member", "auth-dev", "auth-context", "auth-static-context"))
    parser.add_argument("--run", help="Go test name expression; only valid with the install stage")
    parser.add_argument("--test-temp-root", type=Path, help="Existing absolute test-only directory outside the repository; no system TEMP changes")
    parser.add_argument("--context", choices=("unspecified", "desktop-native", "tool"), default="unspecified", help="Caller-declared execution context, not an isolation guarantee")
    args = parser.parse_args()
    if args.run and args.stage != "install":
        parser.error("--run requires the install stage")
    root = Path(__file__).resolve().parent.parent
    windows = os.name == "nt"
    home = Path(os.environ["LOCALAPPDATA"]) / "ipgw-meta-dev" if windows else Path.home() / ".local/share/ipgw-meta-dev"
    go_root = home / "toolchains/go1.25.0/go"
    go = go_root / ("bin/go.exe" if windows else "bin/go")
    if not go.is_file():
        parser.error("Install the verified Go 1.25.0 archive in the documented development directory first")
    if windows and args.stage in ("race", "quality", "cross"):
        parser.error("This stage runs in Ubuntu WSL; Windows runs native ordinary tests")
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%dT%H%M%SZ") + "-" + uuid.uuid4().hex[:8]
    report_dir = root / "build/dev-runs" / ("windows" if windows else "linux") / (stamp + "-" + args.stage)
    report_dir.mkdir(parents=True)
    isolated = home / "runs" / stamp
    for name in ("home", "appdata", "localappdata", "tmp", "config", "data", "state"):
        (isolated / name).mkdir(parents=True)
    # Installation fixtures must stay outside the repository; also avoid deep
    # paths because Windows executable/junction consumers have differing limits.
    temporary = home / "t" / uuid.uuid4().hex[:8] if windows else isolated / "tmp"
    if windows and args.stage in ("qr-dev", "wifi-dev", "logout-structure", "key-relation", "cas-binding", "cas-install", "cas-acquire", "trust-audit", "ctl-check", "ctl-identifiers", "ctl-member", "auth-dev", "auth-context", "auth-static-context"):
        # These are development bundle fixtures, not product installer/profile tests.
        # Keep them on the shared workspace volume so canonical handle paths
        # are testable even when the desktop tool virtualizes LOCALAPPDATA.
        temporary = report_dir / "t"
    if args.test_temp_root:
        test_root = args.test_temp_root
        if not test_root.is_absolute() or not test_root.is_dir() or test_root.is_symlink():
            parser.error("test temp root must be an existing absolute plain directory")
        resolved = test_root.resolve()
        if resolved == Path(resolved.anchor) or resolved == Path.home().resolve() or resolved.is_relative_to(root) or root.is_relative_to(resolved):
            parser.error("test temp root must not overlap repository, user or volume roots")
        temporary = resolved / ("run-" + uuid.uuid4().hex)
    temporary.mkdir(parents=True, exist_ok=True)
    cache = home / "cache"
    for name in ("build", "mod", "gopath"):
        (cache / name).mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    for key in list(env):
        if key.upper().startswith(("IPGW_", "GO", "GIT_CONFIG")) or key.upper() in (
            "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "CURL_CA_BUNDLE", "SSL_CERT_FILE", "SSL_CERT_DIR", "REQUESTS_CA_BUNDLE",
        ):
            env.pop(key, None)
    env.update({
        "PATH": os.pathsep.join((str(go_root / "bin"), str(home / "bin"), os.environ.get("PATH", ""))),
        "GOROOT": str(go_root), "GOTOOLCHAIN": "local", "GOENV": "off", "GOWORK": "off", "GOPROXY": "off", "GOSUMDB": "off", "GOFLAGS": "-mod=readonly",
        "GOCACHE": str(cache / "build"), "GOMODCACHE": str(cache / "mod"), "GOPATH": str(cache / "gopath"),
        "HOME": str(isolated / "home"), "USERPROFILE": str(isolated / "home"),
        "APPDATA": str(isolated / "appdata"), "LOCALAPPDATA": str(isolated / "localappdata"),
        "XDG_CONFIG_HOME": str(isolated / "config"), "XDG_DATA_HOME": str(isolated / "data"), "XDG_STATE_HOME": str(isolated / "state"),
        "TEMP": str(temporary), "TMP": str(temporary), "TMPDIR": str(temporary),
        "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.devnull, "GIT_TERMINAL_PROMPT": "0",
        "HTTP_PROXY": "http://127.0.0.1:9", "HTTPS_PROXY": "http://127.0.0.1:9", "ALL_PROXY": "http://127.0.0.1:9",
        "NO_PROXY": "localhost,127.0.0.1,::1",
    })
    # These process settings prevent module downloads and isolate user state.
    # They are NOT an OS firewall: reviewed tests must use fake transports or loopback.
    version = subprocess.run([str(go), "version"], env=env, cwd=root, capture_output=True, text=True, check=True).stdout.strip()
    if version != "go version go1.25.0 " + ("windows/amd64" if windows else "linux/amd64"):
        parser.error("Unexpected local Go version or architecture: " + version)
    result = {"stage": args.stage, "go_version": version, "report_directory": str(report_dir), "synthetic_only": True, "execution_context_declared": args.context, "go_executable": str(go), "steps": []}

    def run(label, command, timeout=1200, overrides=None):
        child_env = env.copy()
        child_env.update(overrides or {})
        log = report_dir / (label + ".log")
        started = datetime.datetime.now(datetime.timezone.utc).isoformat()
        print("Running " + label, flush=True)
        try:
            with log.open("wb") as out:
                process = subprocess.run(command, cwd=root, env=child_env, stdout=out, stderr=subprocess.STDOUT, timeout=timeout)
            code = process.returncode
        except subprocess.TimeoutExpired:
            code = 124
        except OSError as error:
            log.write_text(type(error).__name__ + ": executable unavailable\n", encoding="utf-8")
            code = 127
        result["steps"].append({"name": label, "command": command, "started_at_utc": started, "exit_code": code, "log": log.name})
        result["passed"] = all(step["exit_code"] == 0 for step in result["steps"])
        (report_dir / "result.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
        print(label + ": exit " + str(code), flush=True)
        if code:
            raise SystemExit(code)

    if args.stage == "qr-dev":
        packages = [".", "./internal/srun", "./internal/qrdev", "./internal/livegate", "./internal/labpath", "./internal/casbundle", "./internal/cmd/qr-dev", "./internal/cmd/qr-dev-launch"]
        run("qr-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("qr-vet", [str(go), "vet"] + packages)
        run("qr-self-test", [str(go), "run", "./internal/cmd/qr-dev", "--self-test"])
        if not windows:
            run("qr-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("qr-fuzz", [str(go), "test", "./internal/srun", "-run=^$", "-fuzz=^FuzzQRContinuation$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        else:
            bundle = report_dir / "ready"
            run("qr-build", [shutil.which("python"), "scripts/build-qr-dev.py", "--output", str(bundle)])
            run("qr-chain", [str(bundle / "qr-dev-launch.exe"), "--self-test"], timeout=55)
            run("qr-build-audit", [shutil.which("python"), "scripts/check-qr-dev-build.py", "--directory", str(bundle)], timeout=180)
    elif args.stage == "wifi-dev":
        packages = ["./internal/wifidev", "./internal/cli", "./internal/labpath", "./internal/cmd/wifi-dev", "./internal/cmd/wifi-dev-launch"]
        run("wifi-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("wifi-vet", [str(go), "vet"] + packages)
        run("wifi-self-test", [str(go), "run", "./internal/cmd/wifi-dev", "--self-test"])
        if not windows:
            run("wifi-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("wifi-fuzz", [str(go), "test", "./internal/labpath", "-run=^$", "-fuzz=^FuzzWiFiConnection$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        else:
            bundle = report_dir / "ready"
            run("wifi-build", [shutil.which("python"), "scripts/build-wifi-dev.py", "--output", str(bundle)])
            run("wifi-chain", [str(bundle / "wifi-dev-launch.exe"), "--self-test"], timeout=55)
            run("wifi-build-audit", [shutil.which("python"), "scripts/check-wifi-dev-build.py", "--directory", str(bundle)], timeout=180)
    elif args.stage == "p4":
        packages = ["./internal/livegate", "./internal/candidate", "./internal/auxgate", "./internal/labpath", "./internal/cmd/ipgw-live-gate"]
        run("p4-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("p4-vet", [str(go), "vet"] + packages)
        if not windows:
            run("p4-race", [str(go), "test", "-race", "-count=1"] + packages)
            for package, target in (("livegate", "FuzzAuxiliaryDecoder"), ("labpath", "FuzzTopology")):
                run("p4-" + target, [str(go), "test", "./internal/" + package, "-run=^$", "-fuzz=^" + target + "$", "-fuzztime=5s", "-parallel=2"], timeout=60)
    elif args.stage == "auth-static-context":
        packages = ["./internal/jsparser", ".", "./internal/cas", "./internal/authdev", "./internal/authcontext", "./internal/casprovenance", "./internal/cascollect", "./internal/casacquire", "./internal/casbundle", "./internal/cmd/auth-static-context", "./internal/cmd/auth-static-context-launch", "./internal/cmd/auth-dev", "./internal/cmd/auth-dev-launch"]
        run("context-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("context-vet", [str(go), "vet"] + packages)
        if not windows:
            run("context-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("context-fuzz", [str(go), "test", "./internal/authcontext", "-run=^$", "-fuzz=^FuzzStaticEnvelope$", "-fuzztime=5s", "-parallel=2"], timeout=60)
            run("parser-match-fuzz", [str(go), "test", "./internal/casprovenance", "-run=^$", "-fuzz=^FuzzStaticAnalysis$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("context-self-test", [str(go), "run", "./internal/cmd/auth-static-context", "--self-test"])
        if windows:
            bundle = report_dir / "ready"
            run("context-build", [shutil.which("python"), "scripts/build-auth-static-context.py", "--output", str(bundle)])
            run("context-chain", [str(bundle / "auth-static-context-launch.exe"), "--self-test"], timeout=45)
            run("context-build-audit", [shutil.which("python"), "scripts/check-auth-static-context-build.py", "--directory", str(bundle)], timeout=180)
    elif args.stage == "auth-context":
        packages = [".", "./internal/cas", "./internal/authdev", "./internal/authcontext", "./internal/casprovenance", "./internal/cascollect", "./internal/casacquire", "./internal/casbundle", "./internal/cmd/auth-context", "./internal/cmd/auth-context-launch", "./internal/cmd/auth-dev", "./internal/cmd/auth-dev-launch"]
        run("context-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("context-vet", [str(go), "vet"] + packages)
        if not windows:
            run("context-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("context-fuzz", [str(go), "test", "./internal/authcontext", "-run=^$", "-fuzz=^FuzzContextReport$", "-fuzztime=5s", "-parallel=2"], timeout=60)
            run("parser-match-fuzz", [str(go), "test", "./internal/casprovenance", "-run=^$", "-fuzz=^FuzzParserMatch$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("context-self-test", [str(go), "run", "./internal/cmd/auth-context", "--self-test"])
        if windows:
            bundle = report_dir / "ready"
            run("context-build", [shutil.which("python"), "scripts/build-auth-context.py", "--output", str(bundle)])
            run("context-chain", [str(bundle / "auth-context-launch.exe"), "--self-test"], timeout=45)
            run("context-build-audit", [shutil.which("python"), "scripts/check-auth-context-build.py", "--directory", str(bundle)], timeout=180)
    elif args.stage == "logout-structure":
        packages = ["./internal/authobserve", "./internal/cmd/auth-dev-launch", "./internal/cmd/auth-observed"]
        run("logout-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("logout-vet", [str(go), "vet"] + packages)
        if not windows:
            run("logout-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("logout-fuzz", [str(go), "test", "./internal/authobserve", "-run=^$", "-fuzz=^FuzzLogoutStructure$", "-fuzztime=10s", "-parallel=2"], timeout=60)
    elif args.stage == "auth-dev":
        packages = [".", "./internal/app", "./internal/cli", "./internal/authdev", "./internal/authobserve", "./internal/livegate", "./internal/casbundle", "./internal/cmd/auth-dev", "./internal/cmd/auth-dev-launch", "./internal/cmd/auth-observed"]
        # Bundle handle fixtures need workspace paths in this tool context;
        # ACL publisher fixtures use the ordinary user-temp location, as in
        # the existing full test stage. Do not alter ACLs to accommodate tests.
        ordinary_packages = [p for p in packages if p != "./internal/livegate"]
        run("auth-tests", [str(go), "test", "-json", "-count=1"] + ordinary_packages)
        live_env = None
        if windows:
            live_temp = home / "t" / uuid.uuid4().hex[:8]
            live_temp.mkdir(parents=True)
            live_env = {"TEMP": str(live_temp), "TMP": str(live_temp), "TMPDIR": str(live_temp)}
        run("auth-livegate-tests", [str(go), "test", "-json", "-count=1", "./internal/livegate"], overrides=live_env)
        run("auth-vet", [str(go), "vet"] + packages)
        if not windows:
            run("auth-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("auth-fuzz", [str(go), "test", "./internal/authdev", "-run=^$", "-fuzz=^FuzzReport$", "-fuzztime=5s", "-parallel=2"], timeout=60)
            run("logout-structure-fuzz", [str(go), "test", "./internal/authobserve", "-run=^$", "-fuzz=^FuzzLogoutStructure$", "-fuzztime=10s", "-parallel=2"], timeout=60)
            run("observed-fuzz", [str(go), "test", "./internal/authobserve", "-run=^$", "-fuzz=^FuzzChildProtocol$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("auth-self-test", [str(go), "run", "./internal/cmd/auth-dev", "--self-test"])
        if windows:
            bundle = report_dir / "ready"
            run("auth-build", [shutil.which("python"), "scripts/build-auth-dev.py", "--output", str(bundle)])
            run("auth-chain", [str(bundle / "auth-dev-launch.exe"), "--self-test"], timeout=45)
            run("auth-build-audit", [shutil.which("python"), "scripts/check-auth-dev-build.py", "--directory", str(bundle)], timeout=120)
    elif args.stage == "key-relation":
        packages = [".", "./internal/cas", "./internal/caskey", "./internal/authdev", "./internal/authcontext", "./internal/authobserve", "./internal/cmd/auth-dev", "./internal/cmd/auth-context", "./internal/cmd/auth-static-context", "./internal/cmd/auth-context-launch", "./internal/cmd/auth-static-context-launch", "./internal/cmd/cas-acquire-launch"]
        run("key-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("key-vet", [str(go), "vet"] + packages)
        if not windows:
            run("key-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("key-fuzz", [str(go), "test", "./internal/caskey", "-run=^$", "-fuzz=^FuzzFiniteRelation$", "-fuzztime=10s", "-parallel=2"], timeout=60)
    elif args.stage == "focus":
        run("cas", [str(go), "test", "-count=1", "./internal/cas", "./internal/caskey"])
        run("sdk", [str(go), "test", "-count=1", "."])
    elif args.stage == "install":
        command = [str(go), "test", "-json", "-count=1"]
        if args.run:
            command += ["-run", args.run]
        run("install", command + ["./internal/installtest"])
    elif args.stage == "test":
        # Ignored build/ contains historical scan copies without go.mod. Go's
        # ./... walks those copies despite .gitignore; they are not current
        # source packages. Enumerate every authoritative source package.
        patterns = [".", "./cmd/...", "./internal/..."]
        if windows:
            listed = subprocess.run([str(go), "list", *patterns], cwd=root, env=env, capture_output=True, check=True, text=True, encoding="utf-8").stdout.splitlines()
            diagnostic = [p for p in listed if p.endswith("/internal/casbundle") or "/internal/cmd/" in p and any(p.rsplit("/", 1)[-1].startswith(x) for x in ("cas-", "ctl-", "trust-", "auth-"))]
            installer = [p for p in listed if p.endswith("/internal/installtest")]
            ordinary = [p for p in listed if p not in diagnostic + installer]
            run("tests", [str(go), "test", "-json", "-count=1"] + ordinary)
            bundle_temp = report_dir / "bundle-fixtures"
            bundle_temp.mkdir()
            run("diagnostic-tests", [str(go), "test", "-json", "-count=1"] + diagnostic, overrides={"TEMP": str(bundle_temp), "TMP": str(bundle_temp), "TMPDIR": str(bundle_temp)})
            run("installer-tests", [str(go), "test", "-json", "-count=1"] + installer)
        else:
            run("tests", [str(go), "test", "-json", "-count=1"] + patterns)
    elif args.stage == "race":
        run("race", [str(go), "test", "-json", "-race", "-count=1", ".", "./cmd/...", "./internal/..."])
    elif args.stage == "vet":
        run("vet", [str(go), "vet", ".", "./cmd/...", "./internal/..."])
    elif args.stage == "docs":
        run("doccheck", [str(go), "run", "./cmd/doccheck", "--check"])
    elif args.stage == "cas-observe":
        run("cas-observe", [str(go), "run", "./internal/cmd/cas-observe"])
    elif args.stage == "cas-provenance":
        packages = ["./internal/casprovenance", "./internal/cmd/cas-provenance"]
        run("provenance-tests", [str(go), "test", "-json", "-count=1"] + packages)
        if not windows:
            run("provenance-race", [str(go), "test", "-race", "-count=1"] + packages)
            for target in ("FuzzBoundedAnalysis", "FuzzReportDecoder"):
                run(target, [str(go), "test", "./internal/casprovenance", "-run=^$", "-fuzz=^" + target + "$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("provenance-vet", [str(go), "vet"] + packages)
        run("provenance-self-test", [str(go), "run", "./internal/cmd/cas-provenance", "--self-test"])
    elif args.stage == "trust-audit":
        packages = ["./internal/trustaudit", "./internal/casbundle", "./internal/casprovenance", "./internal/cmd/trust-audit", "./internal/cmd/trust-audit-launch", "./internal/cmd/cas-acquire-launch"]
        run("trust-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("trust-vet", [str(go), "vet"] + packages)
        if not windows:
            run("trust-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("trust-fuzz", [str(go), "test", "./internal/trustaudit", "-run=^$", "-fuzz=^FuzzTrustAuditReport$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("trust-self-test", [str(go), "run", "./internal/cmd/trust-audit", "--self-test"])
        if windows:
            bundle = report_dir / "ready"
            run("trust-build", [shutil.which("python"), "scripts/build-trust-audit.py", "--output", str(bundle)])
            run("trust-chain", [str(bundle / "trust-audit-launch.exe"), "--self-test"], timeout=30)
            run("trust-build-audit", [shutil.which("python"), "scripts/check-trust-audit-build.py", "--directory", str(bundle)], timeout=120)
    elif args.stage == "ctl-member":
        packages = ["./internal/ctlcheck", "./internal/cmd/ctl-member"]
        run("member-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("member-vet", [str(go), "vet"] + packages)
        if not windows:
            run("member-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("member-key-fuzz", [str(go), "test", "./internal/ctlcheck", "-run=^$", "-fuzz=^FuzzMemberKey$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("member-self-test", [str(go), "run", "./internal/cmd/ctl-member", "--self-test"], timeout=30)
        if windows:
            run("member-build", [shutil.which("python"), "scripts/build-ctl-member.py", "--output", str(report_dir / "ready")])
            run("member-packaged-self-test", [str(report_dir / "ready/ctl-member.exe"), "--self-test"], timeout=15)
            run("member-build-audit", [shutil.which("python"), "scripts/check-ctl-member-build.py", "--directory", str(report_dir / "ready")], timeout=120)
    elif args.stage == "ctl-check":
        packages = ["./internal/ctlcheck", "./internal/casbundle", "./internal/casprovenance", "./internal/cmd/ctl-check", "./internal/cmd/ctl-check-launch", "./internal/cmd/cas-acquire-launch", "./internal/cmd/ctl-identifiers", "./internal/cmd/ctl-identifiers-launch"]
        run("ctl-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("ctl-vet", [str(go), "vet"] + packages)
        if not windows:
            run("ctl-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("ctl-fuzz", [str(go), "test", "./internal/ctlcheck", "-run=^$", "-fuzz=^FuzzCTLReport$", "-fuzztime=5s", "-parallel=2"], timeout=60)
            run("ctl-der-fuzz", [str(go), "test", "./internal/ctlcheck", "-run=^$", "-fuzz=^FuzzCTLDER$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("ctl-self-test", [str(go), "run", "./internal/cmd/ctl-check", "--self-test"])
        if windows:
            bundle = report_dir / "ready"
            run("ctl-build", [shutil.which("python"), "scripts/build-ctl-check.py", "--output", str(bundle)])
            run("ctl-chain", [str(bundle / "ctl-check-launch.exe"), "--self-test"], timeout=30)
            run("ctl-build-audit", [shutil.which("python"), "scripts/check-ctl-check-build.py", "--directory", str(bundle)], timeout=120)
    elif args.stage == "ctl-identifiers":
        packages = ["./internal/ctlcheck", "./internal/casbundle", "./internal/casprovenance", "./internal/cmd/ctl-identifiers", "./internal/cmd/ctl-identifiers-launch", "./internal/cmd/cas-acquire-launch", "./internal/cmd/ctl-check", "./internal/cmd/ctl-check-launch"]
        run("ctl-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("ctl-vet", [str(go), "vet"] + packages)
        if not windows:
            run("ctl-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("ctl-fuzz", [str(go), "test", "./internal/ctlcheck", "-run=^$", "-fuzz=^FuzzIdentifierReport$", "-fuzztime=5s", "-parallel=2"], timeout=60)
            run("ctl-der-fuzz", [str(go), "test", "./internal/ctlcheck", "-run=^$", "-fuzz=^FuzzIdentifierLocator$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("ctl-self-test", [str(go), "run", "./internal/cmd/ctl-identifiers", "--self-test"])
        if windows:
            bundle = report_dir / "ready"
            run("ctl-build", [shutil.which("python"), "scripts/build-ctl-identifiers.py", "--output", str(bundle)])
            run("ctl-chain", [str(bundle / "ctl-identifiers-launch.exe"), "--self-test"], timeout=30)
            run("ctl-build-audit", [shutil.which("python"), "scripts/check-ctl-identifiers-build.py", "--directory", str(bundle)], timeout=120)
    elif args.stage == "cas-acquire":
        packages = ["./internal/ctlcheck", "./internal/casacquire", "./internal/cascollect", "./internal/casbundle", "./internal/casprovenance", "./internal/cmd/cas-acquire", "./internal/cmd/cas-acquire-launch"]
        run("acquisition-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("acquisition-vet", [str(go), "vet"] + packages)
        if not windows:
            run("acquisition-race", [str(go), "test", "-race", "-count=1"] + packages)
            for target in ("FuzzDNSPacket", "FuzzAcquisitionReport", "FuzzSourceDifferenceReport", "FuzzMembershipReport"):
                run(target, [str(go), "test", "./internal/casacquire", "-run=^$", "-fuzz=^" + target + "$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("acquisition-self-test", [str(go), "run", "./internal/cmd/cas-acquire", "--self-test"])
        if windows:
            bundle = report_dir / "ready"
            run("acquisition-build", [shutil.which("python"), "scripts/build-acquisition.py", "--output", str(bundle)])
            run("acquisition-chain", [str(bundle / "cas-acquire-launch.exe"), "--self-test"], timeout=60)
            run("acquisition-build-audit", [shutil.which("python"), "scripts/check-acquisition-build.py", "--directory", str(bundle)], timeout=120)
    elif args.stage == "cas-collect":
        packages = ["./internal/cascollect", "./internal/cmd/cas-collect", "./internal/casprovenance"]
        run("collection-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("collection-vet", [str(go), "vet"] + packages)
        if not windows:
            run("collection-race", [str(go), "test", "-race", "-count=1"] + packages)
            for target in ("FuzzReplayPage", "FuzzCollectionReport"):
                run(target, [str(go), "test", "./internal/cascollect", "-run=^$", "-fuzz=^" + target + "$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        run("collection-self-test", [str(go), "run", "./internal/cmd/cas-collect", "--self-test"])
    elif args.stage in ("cas-binding", "cas-install"):
        packages = ["./internal/casbundle", "./internal/cmd/cas-provenance-launch", "./internal/casprovenance", "./internal/cmd/cas-provenance"]
        run("binding-tests", [str(go), "test", "-json", "-count=1"] + packages)
        run("binding-vet", [str(go), "vet"] + packages)
        if not windows:
            run("binding-race", [str(go), "test", "-race", "-count=1"] + packages)
            run("binding-fuzz", [str(go), "test", "./internal/casbundle", "-run=^$", "-fuzz=^FuzzMetadata$", "-fuzztime=5s", "-parallel=2"], timeout=60)
            if args.stage == "cas-install":
                run("archive-fuzz", [str(go), "test", "./internal/casbundle", "-run=^$", "-fuzz=^FuzzArchive$", "-fuzztime=5s", "-parallel=2"], timeout=60)
        else:
            bundle = report_dir / "binding-artifacts"
            run("binding-build", [shutil.which("python"), "scripts/build-provenance-bundle.py", "--output", str(bundle)])
            receipt = json.loads((bundle / "build-receipt.json").read_text(encoding="utf-8"))
            run("binding-build-audit", [shutil.which("python"), "scripts/check-provenance-build.py", "--directory", str(bundle)])
            # This anchor is from this invocation's reviewed local build, not an
            # arbitrary package's own checksum file. It grants no remote trust.
            run("binding-chain", [str(bundle / "reviewed-launcher/cas-provenance-launch.exe"), "--bundle", str(bundle / "package"), "--manifest-sha256", receipt["manifest_sha256"]], timeout=60)
            if args.stage == "cas-install":
                run("synthetic-install", [shutil.which("python"), "scripts/check-provenance-install.py", "--directory", str(bundle)], timeout=90)
    elif args.stage in ("build", "cross"):
        platforms = [("windows" if windows else "linux", "amd64")] if args.stage == "build" else [(system, arch) for system in ("darwin", "linux", "windows") for arch in ("amd64", "arm64")]
        for system, arch in platforms:
            directory = report_dir / "artifacts" / (system + "-" + arch)
            directory.mkdir(parents=True)
            for binary in ("ipgw", "ipgw-meta", "ipgw-legacy"):
                destination = directory / (binary + (".exe" if system == "windows" else ""))
                label = system + "-" + arch + "-" + binary
                run(label + "-deps", [str(go), "list", "-deps", "./cmd/" + binary], overrides={"CGO_ENABLED": "0", "GOOS": system, "GOARCH": arch})
                dependencies = (report_dir / (label + "-deps.log")).read_text().splitlines()
                forbidden = ("/internal/qrdev", "/internal/wifidev", "/internal/auxgate", "/internal/livegate", "/internal/labpath", "/internal/authdev", "/internal/authobserve", "/internal/authcontext", "/internal/caslegacy", "/internal/castest", "/internal/ctlcheck", "/internal/casacquire", "/internal/trustaudit", "/internal/casbundle", "/internal/cascollect", "/internal/casprovenance", "/internal/cmd/")
                if any(any(marker in dependency for marker in forbidden) or (dependency.startswith("github.com/dop251/goja") and dependency not in ("github.com/dop251/goja/ast", "github.com/dop251/goja/file", "github.com/dop251/goja/token", "github.com/dop251/goja/unistring")) for dependency in dependencies):
                    result["passed"] = False
                    result["product_dependency_isolation"] = False
                    (report_dir / "result.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
                    raise SystemExit("Product depends on diagnostic code")
                result["product_dependency_isolation"] = True
                run(system + "-" + arch + "-" + binary, [str(go), "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -X main.version=dev-local", "-o", str(destination), "./cmd/" + binary], overrides={"CGO_ENABLED": "0", "GOOS": system, "GOARCH": arch})
    elif args.stage == "quality":
        run("promotion", ["python3", "-W", "error::ResourceWarning", "scripts/test-promotion.py"])
        for name in ("milestone-gate", "candidate-checks-jq", "gitleaks"):
            run(name, ["bash", "scripts/test-" + name + ".sh"])
        run("actionlint", [str(home / "bin/actionlint")])
        run("history-secrets", [str(home / "bin/gitleaks"), "git", "--config=.gitleaks.toml", "--redact", "--no-banner", "--log-opts=--all --full-history --reflog", "."])
        # Only versioned/nonignored source, never private ignored build/evidence material.
        files = subprocess.run(["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"], cwd=root, env=env, capture_output=True, check=True).stdout.split(b"\0")
        snapshot = isolated / "source-snapshot"
        snapshot.mkdir()
        for entry in set(files):
            if not entry:
                continue
            relative = Path(os.fsdecode(entry))
            source = root / relative
            if source.is_symlink() or relative.is_absolute() or ".." in relative.parts:
                raise SystemExit("Unsafe source snapshot entry")
            if not source.exists():
                continue  # A tracked deletion has no worktree bytes to scan.
            target = snapshot / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, target)
        run("worktree-secrets", [str(home / "bin/gitleaks"), "dir", "--config=" + str(root / ".gitleaks.toml"), "--redact", "--no-banner", str(snapshot)])
    print("Report: " + str(report_dir / "result.json"), flush=True)


if __name__ == "__main__":
    main()
