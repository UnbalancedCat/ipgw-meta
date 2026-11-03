package doccheck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testFrontMatter = "---\nplan_id: " + PlanID + "\nrevision: " + Revision + "\n---\n\n"

func TestR2GovernanceInputs(t *testing.T) {
	if Revision != "2026-08-28-r2" {
		t.Fatalf("Revision = %q", Revision)
	}
	want := map[string]bool{
		"docs/architecture/decisions/ADR-0007-immutable-candidate-promotion.md":   false,
		"docs/architecture/decisions/ADR-0008-offline-transactional-installer.md": false,
		"docs/architecture/decisions/ADR-0009-separated-live-test-plane.md":       false,
		"docs/architecture/decisions/ADR-0011-nonblocking-m0-governance.md":       false,
		"docs/operations/offline-install.md":                                      false,
		"docs/operations/live-validation.md":                                      false,
		"docs/runbooks/campus-lab.md":                                             false,
	}
	for _, path := range requiredPaths {
		if _, ok := want[path]; ok {
			want[path] = true
		}
	}
	for path, found := range want {
		if !found {
			t.Errorf("requiredPaths is missing %s", path)
		}
	}
}

func TestGenerateAndCheckDeterministicIndex(t *testing.T) {
	root := seedRepository(t)
	changed, violations, err := Generate(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("Generate() violations:\n%s", formatViolations(violations))
	}
	if !changed {
		t.Fatal("first Generate() reported no change")
	}

	indexFile := filepath.Join(root, filepath.FromSlash(IndexPath))
	first, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"| `ADR-0001` | 产品边界 |",
		"architecture/decisions/ADR-0001-product-boundaries.md#adr-0001产品边界",
		"reference/go-sdk.md#sdk-test-001中文-标题",
	} {
		if !bytes.Contains(first, []byte(want)) {
			t.Errorf("generated index does not contain %q:\n%s", want, first)
		}
	}

	changed, violations, err = Generate(root)
	if err != nil || len(violations) != 0 || changed {
		t.Fatalf("second Generate() = changed %v, violations %v, err %v", changed, violations, err)
	}
	if violations, err = Check(root); err != nil || len(violations) != 0 {
		t.Fatalf("Check() = violations %v, err %v", violations, err)
	}

	crlf := bytes.ReplaceAll(first, []byte("\n"), []byte("\r\n"))
	if err := os.WriteFile(indexFile, crlf, 0o644); err != nil {
		t.Fatal(err)
	}
	if violations, err = Check(root); err != nil || len(violations) != 0 {
		t.Fatalf("Check(CRLF) = violations %v, err %v", violations, err)
	}
	if err := os.WriteFile(indexFile, append(crlf, []byte("drift\r\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	violations, err = Check(root)
	if err != nil {
		t.Fatal(err)
	}
	requireViolation(t, violations, "generated stable-ID index is stale")
}

func TestEvidenceInstanceReferencesAreNotDocumentationDeclarations(t *testing.T) {
	for _, tc := range []struct {
		id    string
		valid bool
	}{
		{"EVID-20260904-001", true},
		{"EVID-20240229-999", true},
		{"EVID-20260229-001", false},
		{"EVID-20260904-000", false},
		{"EVID-UNDECLARED-001", false},
	} {
		t.Run(tc.id, func(t *testing.T) {
			root := seedRepository(t)
			appendFile(t, root, "agent/handoff.md", "\nEvidence instance: `"+tc.id+"`\n")
			_, violations, err := Generate(root)
			if err != nil {
				t.Fatal(err)
			}
			if tc.valid {
				if len(violations) != 0 {
					t.Fatalf("valid evidence instance was treated as a docs ID: %v", violations)
				}
			} else {
				requireViolation(t, violations, "references undeclared docs ID "+tc.id)
			}
		})
	}
}

func TestPromotionReleaseNotes(t *testing.T) {
	manual, err := os.ReadFile("../../docs/operations/release.md")
	if err != nil {
		t.Fatal(err)
	}
	// The promotion verifier's own regression compares this same specification
	// block with the bytes it requires from the release artifact.
	const title = "# IPGW-Meta v1.0.0\n"
	_, rest, ok := strings.Cut(string(manual), "```text\n"+title)
	if !ok {
		t.Fatal("missing canonical release notes in specification")
	}
	body, _, ok := strings.Cut(rest, "\n```")
	if !ok {
		t.Fatal("unterminated release notes in specification")
	}
	exact := title + body + "\n"
	const notes = "docs/evidence/releases/v1.0.0/release-notes.md"
	for _, tc := range []struct{ name, path, body, want string }{
		{"canonical", notes, exact, ""},
		{"header", notes, testFrontMatter + exact, "v1.0.0 canonical release notes mismatch"},
		{"extra_text", notes, exact + "extra\n", "v1.0.0 canonical release notes mismatch"},
		{"missing_lf", notes, strings.TrimSuffix(exact, "\n"), "v1.0.0 canonical release notes mismatch"},
		{"crlf", notes, strings.ReplaceAll(exact, "\n", "\r\n"), "v1.0.0 canonical release notes mismatch"},
		{"neighbor", "docs/evidence/releases/v1.0.0/other.md", exact, "front matter plan_id must be " + PlanID},
		{"other_version", "docs/evidence/releases/v1.0.1/release-notes.md", exact, "front matter plan_id must be " + PlanID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := seedRepository(t)
			if _, v, err := Generate(root); err != nil || len(v) > 0 {
				t.Fatalf("fixture: %v %v", v, err)
			}
			writeFile(t, root, tc.path, tc.body)
			v, err := Check(root)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != "" {
				requireViolation(t, v, tc.want)
			} else if len(v) > 0 {
				t.Fatal(v)
			}
		})
	}
}

func TestValidationFailures(t *testing.T) {
	tests := []struct {
		name string
		want string
		edit func(*testing.T, string)
	}{
		{
			name: "duplicate declaration",
			want: "stable ID SDK-TEST-001 is already declared",
			edit: func(t *testing.T, root string) {
				appendFile(t, root, "docs/architecture/overview.md", "\n## SDK-TEST-001：重复声明\n")
			},
		},
		{
			name: "unknown reference",
			want: "references undeclared docs ID UNKNOWN-CASE-999",
			edit: func(t *testing.T, root string) {
				appendFile(t, root, "docs/README.md", "\n`UNKNOWN-CASE-999`\n")
			},
		},
		{
			name: "agent plain reference",
			want: "agent reference SDK-TEST-001 must be a link",
			edit: func(t *testing.T, root string) {
				writeFile(t, root, "agent/handoff.md", testFrontMatter+"# 交接\n\n`SDK-TEST-001`\n")
			},
		},
		{
			name: "agent wrong owner",
			want: "stable ID SDK-TEST-001 link must target its owner",
			edit: func(t *testing.T, root string) {
				writeFile(t, root, "agent/handoff.md", testFrontMatter+"# 交接\n\n[`SDK-TEST-001`](../docs/architecture/security.md)\n")
			},
		},
		{
			name: "bad fragment",
			want: "relative link fragment #不存在 does not exist",
			edit: func(t *testing.T, root string) {
				appendFile(t, root, "docs/README.md", "\n[坏锚点](reference/go-sdk.md#不存在)\n")
			},
		},
		{
			name: "missing path",
			want: "relative link target does not exist",
			edit: func(t *testing.T, root string) {
				appendFile(t, root, "docs/README.md", "\n[缺失](missing.md)\n")
			},
		},
		{
			name: "repository escape",
			want: "relative link escapes repository",
			edit: func(t *testing.T, root string) {
				outside := filepath.Join(filepath.Dir(root), "doccheck-outside.md")
				if err := os.WriteFile(outside, []byte("# outside\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Remove(outside) })
				appendFile(t, root, "docs/README.md", "\n[逃逸](../../doccheck-outside.md)\n")
			},
		},
		{
			name: "revision mismatch",
			want: "front matter revision must be " + Revision,
			edit: func(t *testing.T, root string) {
				writeFile(t, root, "docs/upgrade/status.md",
					"---\nplan_id: "+PlanID+"\nrevision: old\n---\n\n# 状态\n")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := seedRepository(t)
			test.edit(t, root)
			changed, violations, err := Generate(root)
			if err != nil {
				t.Fatal(err)
			}
			if changed {
				t.Fatal("Generate() wrote an index despite source violations")
			}
			requireViolation(t, violations, test.want)
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(IndexPath))); !os.IsNotExist(err) {
				t.Fatalf("index exists after failed Generate(): %v", err)
			}
		})
	}
}

func seedRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range requiredPaths {
		body := testFrontMatter + "# " + filepath.Base(name) + "\n"
		if name == "AGENTS.md" {
			body = "# Agents\n"
		} else if base := filepath.Base(name); strings.HasPrefix(base, "ADR-") {
			body = testFrontMatter + "# " + base[:8] + "：测试决策\n"
		}
		writeFile(t, root, name, body)
	}
	writeFile(t, root, "docs/README.md", testFrontMatter+
		"# 文档中心\n\n"+
		"[同页](#文档中心)\n"+
		"[中文](reference/go-sdk.md?view=1#sdk-test-001中文-标题)\n"+
		"[重复](reference/go-sdk.md#重复-1)\n")
	writeFile(t, root, "docs/reference/go-sdk.md", testFrontMatter+
		"# SDK\n\n"+
		"## SDK-TEST-001：中文 标题\n\n"+
		"## 重复\n\n"+
		"## 重复\n")
	writeFile(t, root, "docs/architecture/decisions/ADR-0001-product-boundaries.md", testFrontMatter+
		"# ADR-0001：产品边界\n")
	writeFile(t, root, "agent/handoff.md", testFrontMatter+
		"# 交接\n\n[`SDK-TEST-001`](../docs/reference/go-sdk.md)\n")
	writeFile(t, root, "AGENTS.md", "# Agents\n\n[`SDK-TEST-001`](docs/reference/go-sdk.md)\n")
	return root
}

func writeFile(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendFile(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(current, []byte(body)...), 0o644); err != nil {
		t.Fatal(err)
	}
}

func requireViolation(t *testing.T, violations []Violation, want string) {
	t.Helper()
	for _, violation := range violations {
		if strings.Contains(violation.String(), want) {
			return
		}
	}
	t.Fatalf("missing violation %q:\n%s", want, formatViolations(violations))
}

func formatViolations(violations []Violation) string {
	var lines []string
	for _, violation := range violations {
		lines = append(lines, violation.String())
	}
	return strings.Join(lines, "\n")
}
