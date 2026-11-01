package auxgate

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/candidate"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
)

var errInstall = errors.New("auxgate: install verification incomplete")

func collectInstall(ctx context.Context, flags map[string]string, identity candidate.Result, a *livegate.Auxiliary) (*livegate.InstallVerifier, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, errInstall
	}
	root := filepath.Join(flags["--work-root"], "ipgw-native-aux-"+hex.EncodeToString(nonce[:]))
	// Only this newly allocated directory is tightened. Keep the private fixture
	// on failure for local review; raw test output is never saved.
	if livegate.PrivateWorkDirectory(root) != nil {
		return nil, errInstall
	}
	dirs := map[string]string{}
	for _, name := range []string{"source", "tmp", "home", "appdata", "localappdata", "gocache", "gopath"} {
		p := filepath.Join(root, name)
		if os.Mkdir(p, 0700) != nil {
			return nil, errInstall
		}
		dirs[name] = p
	}
	source, digest, err := candidate.SnapshotGitSource(ctx, flags["--repository"], identity.SourceCommit, dirs["source"])
	if err != nil || source.Tree != identity.SourceTree || source.BuildInputSHA256 != identity.BuildInputSHA256 {
		return nil, errInstall
	}
	goexe := flags["--go"]
	if !strings.EqualFold(filepath.Base(goexe), "go.exe") {
		return nil, errInstall
	}
	gi, err := os.Lstat(goexe)
	if err != nil || !gi.Mode().IsRegular() || gi.Mode()&os.ModeSymlink != 0 {
		return nil, errInstall
	}
	goHash, err := hashFile(goexe, 128<<20)
	if err != nil {
		return nil, errInstall
	}
	before, err := userPathDigest()
	if err != nil {
		return nil, errInstall
	}
	env := installEnvironment(dirs, goexe, flags["--module-cache"], identity)
	run := func(args ...string) ([]byte, error) {
		return boundedCommand(ctx, goexe, args, dirs["source"], env, 8<<20)
	}
	version, err := run("version")
	if err != nil || strings.TrimSpace(string(version)) != "go version go1.25.0 windows/amd64" {
		return nil, errInstall
	}
	if _, err = run("mod", "verify"); err != nil {
		return nil, errInstall
	}
	testBinary := filepath.Join(root, "install.test.exe")
	if _, err = run("test", "-c", "-o", testBinary, "./internal/installtest"); err != nil {
		return nil, errInstall
	}
	binaryHash, err := hashFile(testBinary, 128<<20)
	if err != nil {
		return nil, errInstall
	}
	verifier := &livegate.InstallVerifier{SourceTree: source.Tree, SnapshotSHA256: digest, GoVersion: "go1.25.0 windows/amd64", TestBinarySHA256: binaryHash}
	// test2json runs this compiled verifier. A forced required Candidate path
	// prevents the native test's historical artifact fallback and skip branch.
	raw, err := run("tool", "test2json", "-t", "-p", "github.com/UnbalancedCat/ipgw-meta/internal/installtest", testBinary, "-test.v=test2json", "-test.run=^TestNativeReleaseAssetSmoke$", "-test.timeout=10m")
	if err != nil || validateInstallEvents(raw) != nil {
		return verifier, errInstall
	}
	for _, name := range livegate.InstallChecks[1:8] {
		pass(a, name)
	}
	after, err := userPathDigest()
	if err != nil || before != after {
		fail(a, "real_user_path_unchanged")
		return verifier, errInstall
	}
	pass(a, "real_user_path_unchanged")
	currentGo, e := hashFile(goexe, 128<<20)
	currentTest, e2 := hashFile(testBinary, 128<<20)
	if e != nil || e2 != nil || currentGo != goHash || currentTest != binaryHash {
		return verifier, errInstall
	}
	return verifier, nil
}

func installEnvironment(d map[string]string, goexe, modcache string, id candidate.Result) []string {
	var env []string
	for _, name := range []string{"SystemRoot", "WINDIR", "COMSPEC", "OS", "PATHEXT", "PROCESSOR_ARCHITECTURE", "PROGRAMDATA", "PROGRAMFILES", "PROGRAMFILES(X86)", "PROGRAMW6432", "SYSTEMDRIVE"} {
		if v := os.Getenv(name); v != "" {
			env = append(env, name+"="+v)
		}
	}
	values := map[string]string{"PATH": filepath.Dir(goexe) + string(os.PathListSeparator) + filepath.Join(os.Getenv("SystemRoot"), "System32") + string(os.PathListSeparator) + filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0"), "HOME": d["home"], "USERPROFILE": d["home"], "APPDATA": d["appdata"], "LOCALAPPDATA": d["localappdata"], "TEMP": d["tmp"], "TMP": d["tmp"], "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off", "GOENV": "off", "GOFLAGS": "-mod=readonly", "GOCACHE": d["gocache"], "GOPATH": d["gopath"], "GOMODCACHE": modcache, "CGO_ENABLED": "0", "IPGW_NATIVE_INSTALL_REQUIRED": "1", "IPGW_NATIVE_INSTALL_CANDIDATE_ONLY": "1", "IPGW_NATIVE_INSTALL_ARTIFACT_ROOT": id.Root, "IPGW_NATIVE_INSTALL_VERSION": candidate.Version, "IPGW_NATIVE_INSTALL_SOURCE_COMMIT": id.SourceCommit, "IPGW_NATIVE_INSTALL_SOURCE_TREE": id.SourceTree}
	for k, v := range values {
		env = append(env, k+"="+v)
	}
	return env
}

func boundedCommand(ctx context.Context, exe string, args []string, dir string, env []string, limit int) ([]byte, error) {
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.WaitDelay = 2 * time.Second
	out := &limitBuffer{limit: limit}
	stderr := &limitBuffer{limit: 64 << 10}
	cmd.Stdout = out
	cmd.Stderr = stderr
	if cmd.Run() != nil || ctx.Err() != nil || out.failed || stderr.failed {
		return nil, errInstall
	}
	return out.Bytes(), nil
}

type limitBuffer struct {
	buffer bytes.Buffer
	limit  int
	failed bool
}

func (b *limitBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		b.failed = true
		return 0, io.ErrShortWrite
	}
	return b.buffer.Write(p)
}
func hashFile(path string, max int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, max+1))
	if err != nil || n > max {
		return "", errInstall
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func validateInstallEvents(raw []byte) error {
	names := map[string]bool{"TestNativeReleaseAssetSmoke": false, "TestNativeReleaseAssetSmoke/fresh_install_and_launcher_default": false, "TestNativeReleaseAssetSmoke/upgrade_preserves_launcher": false, "TestNativeReleaseAssetSmoke/basic_rollback_restores_baseline": false}
	started := map[string]bool{}
	packagePass := false
	packageStarted := false
	s := bufio.NewScanner(bytes.NewReader(raw))
	s.Buffer(make([]byte, 4096), 1<<20)
	for s.Scan() {
		if packagePass {
			return errInstall
		}
		if livegate.CheckJSONStructure(s.Bytes(), 1<<20) != nil {
			return errInstall
		}
		var e struct {
			Time    time.Time
			Action  string
			Package string
			Test    string
			Elapsed float64
			Output  string
		}
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		if d.Decode(&e) != nil {
			return errInstall
		}
		if !packageStarted && e.Action != "start" {
			return errInstall
		}
		if e.Package != "github.com/UnbalancedCat/ipgw-meta/internal/installtest" {
			return errInstall
		}
		if e.Test != "" {
			if _, ok := names[e.Test]; !ok {
				return errInstall
			}
		}
		switch e.Action {
		case "start":
			if e.Test != "" || packageStarted {
				return errInstall
			}
			packageStarted = true
		case "run":
			if e.Test == "" || started[e.Test] {
				return errInstall
			}
			started[e.Test] = true
		case "pass":
			if e.Test == "" {
				if packagePass {
					return errInstall
				}
				packagePass = true
			} else {
				if !started[e.Test] || names[e.Test] {
					return errInstall
				}
				names[e.Test] = true
			}
		case "output":
		case "fail", "skip":
			return errInstall
		default:
			return errInstall
		}
	}
	if s.Err() != nil || !packagePass {
		return errInstall
	}
	for name, passed := range names {
		if !passed || !started[name] {
			return errInstall
		}
	}
	return nil
}

func (b *limitBuffer) Bytes() []byte { return b.buffer.Bytes() }
func (b *limitBuffer) Len() int      { return b.buffer.Len() }
