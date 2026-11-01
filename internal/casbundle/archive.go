package casbundle

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
)

const MaxArchive = 40 << 20

type verifiedArchive struct {
	files    map[string][]byte
	manifest Manifest
}

// verifyArchive never extracts paths and never executes a member.
func verifyArchive(ctx context.Context, data []byte, archiveDigest, manifestDigest string) (*verifiedArchive, error) {
	if ctx == nil || ctx.Err() != nil || !digestPattern.MatchString(archiveDigest) || !digestPattern.MatchString(manifestDigest) || len(data) > MaxArchive || Digest(data) != archiveDigest {
		return nil, ErrInvalid
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(z.File) != len(names) {
		return nil, ErrInvalid
	}
	a := &verifiedArchive{files: map[string][]byte{}}
	for _, entry := range z.File {
		if ctx.Err() != nil {
			return nil, ErrInvalid
		}
		limit := int64(0)
		for _, name := range names {
			if entry.Name == name {
				limit = MaxMetadata
				if name == "cas-provenance.exe" {
					limit = MaxExecutable
				}
				break
			}
		}
		if limit == 0 || a.files[entry.Name] != nil || !entry.Mode().IsRegular() || entry.Flags&1 != 0 || (entry.Method != zip.Store && entry.Method != zip.Deflate) || entry.UncompressedSize64 == 0 || entry.UncompressedSize64 > uint64(limit) || entry.CompressedSize64 > uint64(len(data)) {
			return nil, ErrInvalid
		}
		stream, err := entry.Open()
		if err != nil {
			return nil, ErrInvalid
		}
		content, readErr := io.ReadAll(io.LimitReader(&contextReader{ctx: ctx, reader: stream}, limit+1))
		closeErr := stream.Close()
		if readErr != nil || closeErr != nil || int64(len(content)) > limit || uint64(len(content)) != entry.UncompressedSize64 {
			return nil, ErrInvalid
		}
		a.files[entry.Name] = content
	}
	m, r, err := ValidateMetadata(a.files["manifest.json"], a.files["runner.json"], a.files["build-inputs.json"], manifestDigest)
	if err != nil || int64(len(a.files["cas-provenance.exe"])) != r.Size || Digest(a.files["cas-provenance.exe"]) != r.SHA256 || ctx.Err() != nil {
		return nil, ErrInvalid
	}
	a.manifest = m
	return a, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
