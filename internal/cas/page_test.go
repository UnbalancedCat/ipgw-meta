package cas

import (
	"net/url"
	"strings"
	"testing"
)

func TestParsePageDiscoversDynamicValues(t *testing.T) {
	base, err := url.Parse("https://pass.example.test/tpass/login?service=synthetic")
	if err != nil {
		t.Fatal(err)
	}
	publicKey := strings.Repeat("A", 160)
	html := []byte(`<html>
		<form id="loginForm" action="/tpass/auth">
			<input type="hidden" name="lt" value="synthetic-form-state">
			<input type="hidden" name="execution" value="e1s1">
		</form>
		<script>var publicKeyStr='` + publicKey + `'; var uuid='synthetic-qr-id';</script>
		<script src="/tpass/comm/js/login_neu.js"></script>
		<script src="/tpass/comm/js/login-qrcode.js"></script>
	</html>`)

	page, err := ParsePage(base, html)
	if err != nil {
		t.Fatalf("ParsePage() error = %v", err)
	}
	if page.Action == nil || page.Action.String() != "https://pass.example.test/tpass/auth" {
		t.Fatalf("unexpected dynamic action")
	}
	if page.Hidden.Get("lt") != "synthetic-form-state" || page.Hidden.Get("execution") != "e1s1" {
		t.Fatalf("required hidden fields were not discovered")
	}
	inline, external, err := KeySources(base, html)
	if err != nil || len(inline) != 1 || len(external) != 2 {
		t.Fatal("script coverage")
	}
	if page.QRUUID != "synthetic-qr-id" || !page.QRSupported || len(page.Scripts) != 2 {
		t.Fatalf("QR metadata or scripts were not discovered")
	}
	if page.Challenge != ChallengeNone {
		t.Fatalf("ordinary login page challenge = %q", page.Challenge)
	}
}

func TestParsePageRejectsMissingBaseAndOversizedKey(t *testing.T) {
	if _, err := ParsePage(nil, []byte(`<html></html>`)); err != ErrUnrecognized {
		t.Fatalf("ParsePage(nil) error = %v", err)
	}

}

func TestParsePagePreservesDuplicateNamedControlsForCallerValidation(t *testing.T) {
	base, err := url.Parse("https://pass.example.test/tpass/login?service=synthetic")
	if err != nil {
		t.Fatal(err)
	}
	page, err := ParsePage(base, []byte(`<form id="loginForm" action="/tpass/login">`+
		`<input name="lt" value="first"><input name="lt" value="second">`+
		`<input name="execution" value="e1s1"></form>`))
	if err != nil {
		t.Fatal(err)
	}
	if got := page.Hidden["lt"]; len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("duplicate lt values = %#v", got)
	}
}
