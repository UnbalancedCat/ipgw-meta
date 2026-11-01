package casbundle

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"strings"
	"testing"
)

type zipMember struct {
	name string
	data []byte
	mode os.FileMode
}

func archiveFixture(exe []byte) []zipMember {
	m, r, b := fixture(exe)
	return []zipMember{{"build-inputs.json", encode(b), 0600}, {"cas-provenance.exe", exe, 0600}, {"manifest.json", encode(m), 0600}, {"runner.json", encode(r), 0600}}
}
func zipBytes(t testing.TB, members []zipMember, method uint16) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, m := range members {
		h := &zip.FileHeader{Name: m.name, Method: method}
		h.SetMode(m.mode)
		f, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write(m.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
func TestArchiveValidation(t *testing.T) {
	members := archiveFixture([]byte("synthetic executable"))
	anchor := Digest(members[2].data)
	for _, method := range []uint16{zip.Store, zip.Deflate} {
		data := zipBytes(t, members, method)
		if _, err := verifyArchive(context.Background(), data, Digest(data), anchor); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"missing", "extra", "duplicate", "case", "traversal", "absolute", "backslash", "symlink", "directory", "oversize", "tamper", "pairing"} {
		t.Run(name, func(t *testing.T) {
			m := archiveFixture([]byte("synthetic executable"))
			switch name {
			case "missing":
				m = m[:3]
			case "extra":
				m = append(m, zipMember{"extra", []byte("x"), 0600})
			case "duplicate":
				m[0] = m[1]
			case "case":
				m[0].name = "BUILD-INPUTS.json"
			case "traversal":
				m[0].name = "../build-inputs.json"
			case "absolute":
				m[0].name = "/build-inputs.json"
			case "backslash":
				m[0].name = `dir\build-inputs.json`
			case "symlink":
				m[0].mode = os.ModeSymlink | 0600
			case "directory":
				m[0].mode = os.ModeDir | 0700
			case "oversize":
				m[0].data = bytes.Repeat([]byte("x"), MaxMetadata+1)
			case "tamper":
				m[1].data = []byte("replacement")
			case "pairing":
				manifest, r, b := fixture(m[1].data)
				r.SHA256 = Digest([]byte("old executable"))
				m[3].data = encode(r)
				manifest.Files[2] = rec("runner.json", m[3].data)
				m[2].data = encode(manifest)
				m[0].data = encode(b)
			}
			data := zipBytes(t, m, zip.Deflate)
			a := anchor
			if name == "pairing" {
				a = Digest(m[2].data)
			}
			if _, err := verifyArchive(context.Background(), data, Digest(data), a); err != ErrInvalid {
				t.Fatal("invalid archive accepted")
			}
		})
	}
	data := zipBytes(t, members, zip.Store)
	for _, name := range []string{"archive-anchor", "manifest-anchor", "truncated", "crc", "encryption", "method", "size", "archive-limit"} {
		t.Run(name, func(t *testing.T) {
			b := append([]byte{}, data...)
			a := anchor
			switch name {
			case "manifest-anchor":
				a = strings.Repeat("0", 64)
			case "truncated":
				b = b[:len(b)-8]
			case "crc":
				i := bytes.Index(b, []byte("synthetic executable"))
				b[i] ^= 1
			case "encryption", "method", "size":
				i := bytes.Index(b, []byte{'P', 'K', 1, 2})
				switch name {
				case "encryption":
					b[i+8] |= 1
				case "method":
					binary.LittleEndian.PutUint16(b[i+10:], 99)
				case "size":
					binary.LittleEndian.PutUint32(b[i+24:], MaxMetadata+1)
				}
			case "archive-limit":
				b = make([]byte, MaxArchive+1)
			}
			digest := Digest(b)
			if name == "archive-anchor" {
				digest = strings.Repeat("0", 64)
			}
			if _, err := verifyArchive(context.Background(), b, digest, a); err != ErrInvalid {
				t.Fatal("invalid archive accepted")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := verifyArchive(ctx, data, Digest(data), anchor); err != ErrInvalid {
		t.Fatal("canceled accepted")
	}
}
func TestInstallReceiptClosed(t *testing.T) {
	r := InstallReceipt{SchemaVersion: 1, ToolKind: ToolKind, State: "rejected", Reason: "archive_invalid"}
	data, err := EncodeInstallReceipt(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeInstallReceipt(data); err != nil {
		t.Fatal(err)
	}
	for _, b := range [][]byte{bytes.Replace(data, []byte(`"state":"rejected"`), []byte(`"state":"installed"`), 1), bytes.Replace(data, []byte(`"files_written":0`), []byte(`"files_written":1`), 1), bytes.Replace(data, []byte(`"schema_version":1,`), nil, 1), bytes.Replace(data, []byte(`"state":`), []byte(`"state":"rejected","state":`), 1), append(data, []byte(`{}`)...)} {
		if _, err = DecodeInstallReceipt(b); err != ErrInvalid {
			t.Fatal("invalid receipt accepted")
		}
	}
}
func FuzzArchive(f *testing.F) {
	m := archiveFixture([]byte("synthetic"))
	data := zipBytes(f, m, zip.Deflate)
	anchor := Digest(m[2].data)
	f.Add(data)
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > MaxMetadata {
			return
		}
		_, err := verifyArchive(context.Background(), b, Digest(b), anchor)
		if err != nil && err != ErrInvalid {
			t.Fatal("nonconstant error")
		}
		_, _ = DecodeInstallReceipt(b)
	})
}

func TestInstallReceiptStateConsistency(t *testing.T) {
	id := "dev-" + strings.Repeat("a", 64)
	other := "dev-" + strings.Repeat("b", 64)
	base := InstallReceipt{SchemaVersion: 1, ToolKind: ToolKind, State: "incomplete", Reason: "canceled", ArchiveVerified: true, BundleID: &id, DirectoryCreated: true, FilesWritten: 1}
	if _, err := EncodeInstallReceipt(base); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*InstallReceipt){
		func(r *InstallReceipt) { r.Reason = "destination_invalid" },
		func(r *InstallReceipt) { r.Reason = "copy_failed"; r.FilesWritten = 4 },
		func(r *InstallReceipt) { r.ArchiveVerified = false },
		func(r *InstallReceipt) { r.BundleID = nil },
		func(r *InstallReceipt) {
			r.FilesWritten = 4
			r.Reason = "self_test_failed"
			r.SelfTest = &Receipt{SchemaVersion: 1, ToolKind: ToolKind, Binding: "verified", BundleID: &other, Runner: "failed", Reason: "execution_failed", EXEAttempts: 1}
		},
	} {
		r := base
		mutate(&r)
		if _, err := DecodeInstallReceipt(encode(r)); err != ErrInvalid {
			t.Fatal("contradictory install receipt accepted")
		}
	}
}
