package cascollect

import (
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// This import allowlist is a regression guard for the small audited replay
// package, not an assertion that all transitive standard-library code lacks I/O.
func TestReplayDependencyBoundary(t *testing.T) {
	allowed := map[string]bool{"bytes": true, "context": true, "encoding/json": true, "errors": true, "fmt": true, "io": true, "net/url": true, "regexp": true, "strconv": true, "strings": true, "unicode/utf8": true, "github.com/UnbalancedCat/ipgw-meta/internal/casprovenance": true, "golang.org/x/net/html": true}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range f.Imports {
			path, err := strconv.Unquote(im.Path.Value)
			if err != nil || !allowed[path] {
				t.Fatal("unreviewed replay dependency")
			}
		}
	}
}
