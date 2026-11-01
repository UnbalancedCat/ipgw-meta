package caskey

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"strings"
	"testing"
	"time"
)

func TestFiniteRelation(t *testing.T) {
	valid := castest.LoginScript("synthetic-key")
	cases := []struct {
		name, source string
		accept       bool
	}{
		{"valid", valid, true},
		{"comment_decoy", "// const publicKeyStr='decoy';\n" + valid, true},
		{"string_decoy", `var example="publicKeyStr='decoy'";` + valid, true},
		{"template_decoy", "var example=`publicKeyStr='decoy'`;" + valid, true},
		{"substring", `var unusedpublicKeyStr='decoy';` + valid, true},
		{"bare", `const publicKeyStr="synthetic-key";`, false},
		{"parameter", strings.Replace(valid, "syntheticLogin()", "syntheticLogin(RSAEncryptor)", 1), false},
		{"rest_parameter", strings.Replace(valid, "syntheticLogin()", "syntheticLogin(...RSAEncryptor)", 1), false},
		{"rest_jquery", strings.Replace(valid, "syntheticLogin()", "syntheticLogin(...$)", 1), false},
		{"global_replace", `RSAEncryptor=other;` + valid, false},
		{"global_binding", `const RSAEncryptor=other;` + valid, false},
		{"prototype_replace", `RSAEncryptor.prototype={};` + valid, false},
		{"global_jquery", valid + `window.$=other;`, false},
		{"bracket_method", strings.Replace(valid, "var u=", `$u["val"]=other;var u=`, 1), false},
		{"prototype_delete", `delete RSAEncryptor.prototype.encrypt;` + valid, false},
		{"prototype_update", `RSAEncryptor.prototype["encrypt"]++;` + valid, false},
		{"jquery_parameter", strings.Replace(valid, "syntheticLogin()", "syntheticLogin($)", 1), false},
		{"input_swap", strings.Replace(valid, "u+p", "p+u", 1), false},
		{"wrong_input", strings.Replace(valid, "#pd", "#other", 1), false},
		{"dynamic", strings.Replace(valid, `"synthetic-key"`, `part+other`, 1), false},
		{"reassigned", strings.Replace(valid, `$("#rsa").val(cipher);`, `$("#rsa").val(cipher);cipher=other;`, 1), false},
		{"shadow", strings.Replace(valid, "var u=", "if(false){var enc;}var u=", 1), false},
		{"member_write", strings.Replace(valid, "var u=", "enc.encrypt=other;var u=", 1), false},
		{"conditional", "if(flag){" + valid + "}", false},
		{"nested", "function outer($){" + valid + "}", false},
		{"other_write", valid + `;$("#rsa").val(other);`, false},
		{"alias_sink", valid + `;const dest=$("#rsa");dest.val(other);`, false},
		{"computed_sink", strings.Replace(valid, `$("#rsa").val`, `$("#rsa")["val"]`, 1), false},
		{"malformed", valid + "function(", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Collector
			err := c.Add(context.Background(), []byte(tc.source))
			got := err == nil && len(c.Keys) == 1
			if got != tc.accept {
				t.Fatalf("accepted=%v error=%v keys=%d", got, err, len(c.Keys))
			}
		})
	}
}

func TestCollectorBoundsAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var c Collector
	if c.Add(ctx, []byte("var a=1;")) != context.Canceled {
		t.Fatal("cancellation")
	}
	if c.Add(context.Background(), []byte(strings.Repeat(" ", 1<<20+1))) == nil {
		t.Fatal("source bound")
	}
	c = Collector{}
	for i := 0; i < 4; i++ {
		if c.Add(context.Background(), []byte(strings.Repeat(" ", 1<<20))) != nil {
			t.Fatal("budget within bound")
		}
	}
	if c.Add(context.Background(), []byte(" ")) == nil {
		t.Fatal("cumulative byte bound")
	}
	c = Collector{nodes: 100000}
	if c.Add(context.Background(), []byte("var a=1;")) == nil {
		t.Fatal("AST budget")
	}
	c = Collector{}
	if c.Add(context.Background(), []byte(strings.Repeat("(", 80)+"0"+strings.Repeat(")", 80))) == nil {
		t.Fatal("depth")
	}
	c = Collector{spent: 10 * time.Second}
	if c.Add(context.Background(), []byte("")) == nil {
		t.Fatal("time budget")
	}
	c = Collector{}
	chunk := []byte(strings.Repeat("a;", 20000))
	if c.Add(context.Background(), chunk) != nil {
		t.Fatal("token budget within limit")
	}
	if c.Add(context.Background(), chunk) != nil {
		t.Fatal("second source within budget")
	}
	if c.Add(context.Background(), chunk) == nil {
		t.Fatal("cumulative token budget")
	}
}
func FuzzFiniteRelation(f *testing.F) {
	f.Add(castest.LoginScript("synthetic-key"))
	f.Add("/* comment */")
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 65536 {
			return
		}
		var c Collector
		_ = c.Add(context.Background(), []byte(s))
	})
}
