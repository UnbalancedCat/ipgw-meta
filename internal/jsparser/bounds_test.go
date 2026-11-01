package parser_test

import (
	"context"
	bounded "github.com/UnbalancedCat/ipgw-meta/internal/jsparser"
	original "github.com/dop251/goja/parser"
	"reflect"
	"strings"
	"testing"
)

func TestStaticAgreement(t *testing.T) {
	cases := []string{
		`var ratio=6/2; publicKeyStr='synthetic';`,
		`var pattern=/[a-z/]+/g; publicKeyStr='synthetic';`,
		`function f(){if(true){publicKeyStr='synthetic';}}`,
		`var café='x';publicKeyStr='synthetic';`,
		`var ca\u0066e='x';publicKeyStr='synthetic';`,
		"var text=`${name}`;publicKeyStr='synthetic';",
		"// comment\u2028publicKeyStr='synthetic';",
		`var ratio=6/2; /* publicKeyStr='synthetic'; */`,
		`var ratio=6/2;var text="publicKeyStr='synthetic';";`,
		"var f=(x)=>`${x / 2}:${/a/.test(x)}`;",
		`while(false){publicKeyStr='synthetic';}`,
		`/* x */ var f=(a,b)=>a/b; // tail`,
	}
	for i, input := range cases {
		want, e := original.ParseFile(nil, "synthetic", input, 0, original.WithDisableSourceMaps)
		got, _, reason := bounded.ParseLimited([]byte(input), &bounded.Guard{Context: context.Background()})
		if e != nil || reason != "none" || !reflect.DeepEqual(want, got) {
			t.Fatalf("case %d AST disagreement; reason %s", i, reason)
		}
	}
}
func TestLimits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name, input, want string
		guard             *bounded.Guard
	}{
		{"cancel", "var x='a';", "canceled", &bounded.Guard{Context: ctx}},
		{"token", strings.Repeat(";", 101), "token_limit", &bounded.Guard{Context: context.Background(), TokenLimit: 100}},
		{"depth", strings.Repeat("(", 65) + "1" + strings.Repeat(")", 65), "depth_limit", &bounded.Guard{Context: context.Background()}},
		{"syntax", "var x='", "syntax_rejected", &bounded.Guard{Context: context.Background()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, c, r := bounded.ParseLimited([]byte(tc.input), tc.guard)
			if a != nil || c != nil || r != tc.want {
				t.Fatalf("wrong closed rejection: %s", r)
			}
		})
	}
}

type cancelAfter struct {
	context.Context
	calls int
}

func (c *cancelAfter) Err() error {
	c.calls++
	if c.calls > 5 {
		return context.Canceled
	}
	return nil
}
func TestCancelInsideLongSingleToken(t *testing.T) {
	c := &cancelAfter{Context: context.Background()}
	_, _, r := bounded.ParseLimited([]byte("var x='"+strings.Repeat("a", 500000)+"';"), &bounded.Guard{Context: c})
	if r != "canceled" || c.calls < 6 {
		t.Fatal("long token not cancelable")
	}
}
func TestDisabledSources(t *testing.T) {
	input := "var x=1;\n//# sourceMappingURL=https://example.invalid/never.map"
	calls := 0
	_, err := original.ParseFile(nil, "synthetic", input, 0, original.WithDisableSourceMaps, original.WithSourceMapLoader(func(string) ([]byte, error) { calls++; return nil, nil }))
	_, _, r := bounded.ParseLimited([]byte(input), &bounded.Guard{Context: context.Background()})
	if err != nil || calls != 0 || r != "none" {
		t.Fatal("source-map boundary failed")
	}
}

func TestTokenBudgetBoundaryAndCumulativeSources(t *testing.T) {
	// Scan invocations include EOF and lookahead, intentionally conservative.
	g := &bounded.Guard{Context: context.Background(), TokenLimit: 100}
	if _, _, r := bounded.ParseLimited([]byte(strings.Repeat(";", 99)), g); r != "none" || g.Tokens != 100 {
		t.Fatal("token boundary rejected")
	}
	if a, c, r := bounded.ParseLimited([]byte(";"), g); a != nil || c != nil || r != "token_limit" {
		t.Fatal("budget reset across sources")
	}
	g = &bounded.Guard{Context: context.Background(), Bytes: (4 << 20) - 1}
	if _, _, r := bounded.ParseLimited([]byte(";"), g); r != "none" {
		t.Fatal("byte boundary")
	}
	if a, c, r := bounded.ParseLimited([]byte(";"), g); a != nil || c != nil || r != "input_limit" {
		t.Fatal("cumulative byte cap")
	}
}
