package livegate

import (
	"os"
	"path/filepath"
)

// VerifiedHelper binds an auxiliary invocation without requiring an extracted
// product. The caller must independently validate the complete Candidate set.
type VerifiedHelper struct {
	Binding                        CandidateBinding
	executable, manifest           string
	fileSnapshot, manifestSnapshot privateFileSnapshot
}

func VerifyHelperBefore(manifestPath string, platform Platform) (VerifiedHelper, error) {
	if !validCandidatePath(manifestPath) {
		return VerifiedHelper{}, ErrCandidateSecurity
	}
	f, ms, err := openPrivateRegularFile(manifestPath)
	if err != nil {
		return VerifiedHelper{}, ErrCandidateSecurity
	}
	defer f.Close()
	b, err := DecodeCandidateManifest(f, platform)
	if err != nil || !snapshotStillNamesOpenFile(f, ms) {
		return VerifiedHelper{}, ErrCandidateSecurity
	}
	exe, err := os.Executable()
	if err != nil {
		return VerifiedHelper{}, ErrCandidateSecurity
	}
	h, hs, err := openPrivateExecutableFile(exe)
	if err != nil {
		return VerifiedHelper{}, ErrCandidateSecurity
	}
	defer h.Close()
	hash, err := hashOpenCandidate(h, hs.info.Size())
	if err != nil || filepath.Base(exe) != filepath.Base(b.RunnerName) || hs.info.Size() != b.RunnerSize || hash != b.RunnerSHA256 || !snapshotStillNamesOpenFile(h, hs) {
		return VerifiedHelper{}, ErrCandidateSecurity
	}
	return VerifiedHelper{Binding: b, executable: exe, manifest: manifestPath, fileSnapshot: hs, manifestSnapshot: ms}, nil
}

func (v VerifiedHelper) VerifyAfter() error {
	next, err := VerifyHelperBefore(v.manifest, v.Binding.Platform)
	if err != nil || next.Binding != v.Binding || next.executable != v.executable || !samePrivateFileSnapshot(next.fileSnapshot, v.fileSnapshot) || !samePrivateFileSnapshot(next.manifestSnapshot, v.manifestSnapshot) {
		return ErrCandidateDrift
	}
	return nil
}

// VerifyRunnerBefore is used by the formal entry point, not development
// executors that deliberately run a separately bound observed executable.
func VerifyRunnerBefore(candidatePath, manifestPath, explicitSHA string, platform Platform) (VerifiedCandidate, error) {
	v, err := VerifyCandidateBefore(candidatePath, manifestPath, explicitSHA, platform)
	if err != nil {
		return VerifiedCandidate{}, ErrCandidateSecurity
	}
	executable, err := os.Executable()
	if err != nil {
		return VerifiedCandidate{}, ErrCandidateSecurity
	}
	return bindRunner(v, executable)
}

func bindRunner(v VerifiedCandidate, path string) (VerifiedCandidate, error) {
	if !validCandidatePath(path) || filepath.Base(path) != filepath.Base(v.Binding.RunnerName) {
		return VerifiedCandidate{}, ErrCandidateSecurity
	}
	f, snapshot, err := openPrivateExecutableFile(path)
	if err != nil {
		return VerifiedCandidate{}, ErrCandidateSecurity
	}
	defer f.Close()
	if snapshot.info.Size() != v.Binding.RunnerSize || os.SameFile(snapshot.info, v.candidateSnapshot.info) || os.SameFile(snapshot.info, v.manifestSnapshot.info) {
		return VerifiedCandidate{}, ErrCandidateSecurity
	}
	hash, err := hashOpenCandidate(f, snapshot.info.Size())
	if err != nil || hash != v.Binding.RunnerSHA256 || !snapshotStillNamesOpenFile(f, snapshot) {
		return VerifiedCandidate{}, ErrCandidateSecurity
	}
	v.runnerPath, v.runnerSnapshot = path, snapshot
	return v, nil
}

func (v VerifiedCandidate) verifyRunnerAfter() error {
	f, snapshot, err := openPrivateExecutableFile(v.runnerPath)
	if err != nil {
		return ErrCandidateDrift
	}
	defer f.Close()
	if !samePrivateFileSnapshot(snapshot, v.runnerSnapshot) || snapshot.info.Size() != v.Binding.RunnerSize {
		return ErrCandidateDrift
	}
	hash, err := hashOpenCandidate(f, snapshot.info.Size())
	if err != nil || hash != v.Binding.RunnerSHA256 || !snapshotStillNamesOpenFile(f, snapshot) {
		return ErrCandidateDrift
	}
	return nil
}
