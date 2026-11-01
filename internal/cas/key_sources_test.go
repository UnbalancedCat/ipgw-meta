package cas

import (
	"net/url"
	"testing"
)

func TestKeySourceCoverage(t *testing.T) {
	base, _ := url.Parse("https://pass.example.test/tpass/current")
	for _, tc := range []struct {
		name, html string
		accept     bool
		count      int
	}{
		{"inline", `<script>var a=1;</script>`, true, 1},
		{"html_inert", `<div>var key='x'</div><!-- <script>bad</script> --><template><script>bad</script></template><script type="application/json">{"key":"x"}</script>`, true, 0},
		{"module", `<script type="module">bad</script>`, false, 0},
		{"nomodule", `<script nomodule>bad</script>`, false, 0},
		{"base_local", `<base href="/other/"><script src="login.js"></script>`, false, 0},
		{"base_external", `<base href="https://other.example.test/"><script src="login.js"></script>`, false, 0},
		{"base_multiple", `<base href="/a/"><base href="/b/">`, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in, _, e := KeySources(base, []byte(tc.html))
			if (e == nil) != tc.accept || tc.accept && len(in) != tc.count {
				t.Fatal("source coverage mismatch")
			}
		})
	}
	_, out, e := KeySources(base, []byte(`<script src="login.js"></script>`))
	if e != nil || len(out) != 1 || out[0].Path != "/tpass/login.js" {
		t.Fatal("response URL resolution")
	}
}
