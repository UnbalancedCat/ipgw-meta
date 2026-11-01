package casbundle

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func encode(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return append(b, '\n')
}
func rec(name string, b []byte) FileRecord { return FileRecord{name, int64(len(b)), Digest(b)} }
func fixture(exe []byte) (Manifest, Runner, BuildInputs) {
	sources := []FileRecord{rec("go.mod", []byte("synthetic"))}
	data, _ := json.Marshal(sources)
	b := BuildInputs{1, ToolKind, "go1.25.0", "windows/amd64", "0", strings.Repeat("a", 40), "dirty", Digest(data), sources, append([]string{}, flags...)}
	r := Runner{1, ToolKind, "synthetic-only", "cas-provenance.exe", int64(len(exe)), Digest(exe), []string{"--self-test"}, 30, 65536}
	m := Manifest{1, ToolKind, "dev-" + Digest(encode(b)), "windows/amd64", []FileRecord{rec("build-inputs.json", encode(b)), rec("cas-provenance.exe", exe), rec("runner.json", encode(r))}}
	return m, r, b
}
func TestMetadataTrustAndPairing(t *testing.T) {
	m, r, b := fixture([]byte("synthetic executable"))
	mb, rb, bb := encode(m), encode(r), encode(b)
	if _, _, err := ValidateMetadata(mb, rb, bb, Digest(mb)); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"", strings.Repeat("0", 64), "../untrusted"} {
		if _, _, err := ValidateMetadata(mb, rb, bb, expected); err == nil {
			t.Fatal("missing/wrong anchor accepted")
		}
	}
	// Recompute manifest to keep it internally self-consistent, but runner binds
	// another EXE: this is the historical R2 defect, with wholly synthetic bytes.
	r.SHA256 = Digest([]byte("older synthetic executable"))
	rb = encode(r)
	m.Files[2] = rec("runner.json", rb)
	mb = encode(m)
	if _, _, err := ValidateMetadata(mb, rb, bb, Digest(mb)); err == nil {
		t.Fatal("self-consistent runner mismatch accepted")
	}
}
func TestMetadataClosedFieldsAndSemantics(t *testing.T) {
	for _, mutate := range []func(*Manifest, *Runner, *BuildInputs){
		func(m *Manifest, r *Runner, b *BuildInputs) { r.Args = []string{"--live"} },
		func(m *Manifest, r *Runner, b *BuildInputs) { r.Executable = "../escape.exe" },
		func(m *Manifest, r *Runner, b *BuildInputs) { r.TimeoutSeconds = 31 },
		func(m *Manifest, r *Runner, b *BuildInputs) { r.MaxOutputBytes = 0 },
		func(m *Manifest, r *Runner, b *BuildInputs) { b.GoVersion = "go1.23.4" },
		func(m *Manifest, r *Runner, b *BuildInputs) { b.Sources[0].Name = "../outside" },
		func(m *Manifest, r *Runner, b *BuildInputs) { b.SourceDigest = strings.Repeat("0", 64) },
		func(m *Manifest, r *Runner, b *BuildInputs) { b.Worktree = "verified" },
	} {
		m, r, b := fixture([]byte("synthetic"))
		mutate(&m, &r, &b)
		m.BundleID = "dev-" + Digest(encode(b))
		m.Files[0] = rec("build-inputs.json", encode(b))
		m.Files[2] = rec("runner.json", encode(r))
		mb := encode(m)
		if _, _, err := ValidateMetadata(mb, encode(r), encode(b), Digest(mb)); err == nil {
			t.Fatal("invalid metadata accepted")
		}
	}
	m, r, b := fixture([]byte("synthetic"))
	mb := encode(m)
	for _, data := range [][]byte{bytes.Replace(mb, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1), bytes.Replace(mb, []byte(`"schema_version":1,`), nil, 1), append(mb, []byte(`{}`)...), bytes.Replace(mb, []byte(`"files":`), []byte(`"extra":0,"files":`), 1)} {
		if _, _, err := ValidateMetadata(data, encode(r), encode(b), Digest(data)); err == nil {
			t.Fatal("non-closed manifest accepted")
		}
	}
}
func TestMissingAnchorDoesNotStart(t *testing.T) {
	r := Run(context.Background(), "untrusted", "")
	if r.EXEStarts != 0 || r.Binding != "invalid" || r.Result != nil {
		t.Fatal("untrusted launch")
	}
}
func FuzzMetadata(f *testing.F) {
	m, r, b := fixture([]byte("synthetic"))
	rb, bb := encode(r), encode(b)
	f.Add(string(encode(m)))
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > MaxMetadata+1 {
			t.Skip()
		}
		_, _, err := ValidateMetadata([]byte(input), rb, bb, Digest([]byte(input)))
		if err != nil && err != ErrInvalid {
			t.Fatal("nonconstant metadata error")
		}
	})
}
