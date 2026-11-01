package candidate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceSnapshotUsesCommittedBlobsNotDirtyCheckout(t *testing.T) {
	repo := t.TempDir()
	runTestCommand(t, repo, nil, "git", "init", "-q")
	runTestCommand(t, repo, nil, "git", "config", "user.name", "Snapshot Test")
	runTestCommand(t, repo, nil, "git", "config", "user.email", "snapshot@example.invalid")
	runTestCommand(t, repo, nil, "git", "config", "commit.gpgsign", "false")
	writeTestFile(t, repo, "input.txt", []byte("committed\n"), 0600)
	runTestCommand(t, repo, nil, "git", "add", ".")
	commitTestTree(t, repo, "snapshot")
	commit := strings.TrimSpace(runTestCommand(t, repo, nil, "git", "rev-parse", "HEAD"))
	writeTestFile(t, repo, "input.txt", []byte("dirty\n"), 0600)
	writeTestFile(t, repo, "extra.txt", []byte("untracked\n"), 0600)
	dest := t.TempDir()
	source, digest, err := SnapshotGitSource(context.Background(), repo, commit, dest)
	if err != nil || source.Commit != commit || len(digest) != 64 {
		t.Fatal("snapshot", err)
	}
	raw, err := os.ReadFile(filepath.Join(dest, "input.txt"))
	if err != nil || string(raw) != "committed\n" {
		t.Fatal("dirty included")
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 1 {
		t.Fatal("untracked included")
	}
	if _, _, err := SnapshotGitSource(context.Background(), repo, commit, dest); err == nil {
		t.Fatal("overwrote destination")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := SnapshotGitSource(ctx, repo, commit, t.TempDir()); err == nil {
		t.Fatal("ignored cancel")
	}
	if _, _, err := SnapshotGitSource(context.Background(), repo, strings.Repeat("a", 40), t.TempDir()); err == nil {
		t.Fatal("missing commit")
	}
}
