package candidate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SnapshotGitSource materializes regular blobs from one verified commit in an
// existing empty directory. It never copies the dirty checkout or invokes Git
// checkout filters. The digest covers all materialized names, modes and bytes.
func SnapshotGitSource(ctx context.Context, repository, commit, destination string) (GitSource, string, error) {
	source, err := InspectGitSource(ctx, repository, commit)
	if err != nil {
		return GitSource{}, "", err
	}
	info, err := os.Lstat(destination)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !filepath.IsAbs(destination) {
		return GitSource{}, "", ErrInvalidSource
	}
	children, err := os.ReadDir(destination)
	if err != nil || len(children) != 0 {
		return GitSource{}, "", ErrInvalidSource
	}
	raw, err := runGitBounded(ctx, repository, 4<<20, "ls-tree", "-rz", "--full-tree", "-r", commit)
	if err != nil {
		return GitSource{}, "", err
	}
	entries, err := parseGitTree(raw)
	if err != nil || len(entries) > 10000 {
		return GitSource{}, "", ErrInvalidSource
	}
	seen := map[string]bool{}
	var total int
	hash := sha256.New()
	for _, entry := range entries {
		if ctx.Err() != nil {
			return GitSource{}, "", ctx.Err()
		}
		name := string(entry.path)
		clean := filepath.FromSlash(name)
		if !filepath.IsLocal(clean) || strings.ContainsAny(name, "\\:") || strings.HasPrefix(name, "/") {
			return GitSource{}, "", ErrInvalidSource
		}
		for _, part := range strings.Split(name, "/") {
			if part == "" || part == "." || part == ".." || strings.TrimRight(part, " .") != part {
				return GitSource{}, "", ErrInvalidSource
			}
		}
		folded := strings.ToLower(name)
		if seen[folded] {
			return GitSource{}, "", ErrInvalidSource
		}
		seen[folded] = true
		content, err := runGitBounded(ctx, repository, 16<<20, "cat-file", "blob", entry.oid)
		if err != nil {
			return GitSource{}, "", err
		}
		total += len(content)
		if total > 256<<20 {
			return GitSource{}, "", ErrInvalidSource
		}
		target := filepath.Join(destination, clean)
		if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return GitSource{}, "", ErrInvalidSource
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return GitSource{}, "", ErrInvalidSource
		}
		_, we := f.Write(content)
		ce := f.Close()
		if we != nil || ce != nil {
			return GitSource{}, "", ErrInvalidSource
		}
		sum := sha256.Sum256(content)
		fmt.Fprintf(hash, "%s\x00%s\x00%d\x00%x\n", name, entry.mode, len(content), sum)
	}
	return source, hex.EncodeToString(hash.Sum(nil)), nil
}
