package ipgw

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"net/http"
	"strings"
	"testing"
)

func TestPasswordKeyUseRelationBeforeCredentials(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	other, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	encode := func(k *rsa.PublicKey) string {
		der, e := x509.MarshalPKIXPublicKey(k)
		if e != nil {
			t.Fatal(e)
		}
		return base64.StdEncoding.EncodeToString(der)
	}
	active, decoy := encode(&key.PublicKey), encode(&other.PublicKey)
	valid := castest.LoginScript(active)
	wrong := castest.LoginScript(decoy)
	type scenario struct {
		name, inline string
		scripts      []string
		fail         int
		accept       bool
		code         ErrorCode
	}
	cases := []scenario{
		{name: "comment_first", scripts: []string{`//const publicKeyStr="` + decoy + `";`, valid}, accept: true},
		{name: "comment_last", scripts: []string{valid, `//const publicKeyStr="` + decoy + `";`}, accept: true},
		{name: "inline_and_external_same", inline: valid, scripts: []string{valid}, accept: true},
		{name: "same_key_other_encoding", scripts: []string{valid, castest.LoginScript(base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey)))}, accept: true},
		{name: "different", scripts: []string{valid, wrong}},
		{name: "different_swapped", scripts: []string{wrong, valid}},
		{name: "inline_conflict", inline: valid, scripts: []string{wrong}},
		{name: "bare", scripts: []string{`const publicKeyStr="` + active + `";`}},
		{name: "read_failure_after_key", scripts: []string{valid, ""}, fail: 2, code: CodeNetwork},
		{name: "read_failure_before_key", scripts: []string{"", valid}, fail: 1, code: CodeNetwork},
		{name: "parse_failure_after_key", scripts: []string{valid, "function("}},
		{name: "invalid_rsa_before_provider", scripts: []string{castest.LoginScript("invalid")}},
		{name: "too_many", inline: valid, scripts: make([]string, 9)},
		{name: "module", inline: `</script><script type="module">` + valid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gets, credentials, posts := 0, 0, 0
			tr := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodPost {
					posts++
					if err := req.ParseForm(); err != nil {
						t.Fatal(err)
					}
					cipher, e := base64.StdEncoding.DecodeString(req.PostForm.Get("rsa"))
					if e != nil {
						t.Fatal(e)
					}
					plain, e := rsa.DecryptPKCS1v15(rand.Reader, key, cipher)
					if e != nil || string(plain) != "synthetic-usersynthetic-password" {
						t.Fatal("wrong key or plaintext")
					}
					if len(req.PostForm) != 6 {
						t.Fatal("submission field set")
					}
					return testResponse(req, 500, "synthetic-stop", ""), nil
				}
				switch req.URL.Path {
				case "/cgi-bin/rad_user_info":
					return testResponse(req, 200, "not_online_error", ""), nil
				case "/":
					return testResponse(req, 200, "ac_id=1", ""), nil
				case "/tpass/login":
					body := `<form id="loginForm" action="/different/submit"><input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution"></form><script>` + tc.inline + `</script>`
					for i := range tc.scripts {
						body += fmt.Sprintf(`<script src="login-%d.js"></script>`, i)
					}
					return testResponse(req, 200, body, ""), nil
				default:
					if !strings.HasPrefix(req.URL.Path, "/tpass/login-") {
						t.Fatal("script resolved against form action instead of page")
					}
					gets++
					if gets == tc.fail {
						return nil, errors.New("synthetic transport failure")
					}
					if gets > len(tc.scripts) {
						t.Fatal("unexpected source")
					}
					return testResponse(req, 200, tc.scripts[gets-1], ""), nil
				}
			})
			client, e := NewClient(WithRoundTripper(tr))
			if e != nil {
				t.Fatal(e)
			}
			_, e = client.Login(context.Background(), LoginRequest{Method: AuthMethodPassword, ExpectedUsername: "synthetic-user", Credentials: CredentialProviderFunc(func(context.Context, CredentialRequest) (Credential, error) {
				credentials++
				if gets != len(tc.scripts) {
					t.Fatal("credential read before complete collection")
				}
				return Credential{Password: "synthetic-password"}, nil
			})})
			if tc.accept {
				if credentials != 1 || posts != 1 || !IsCode(e, CodeNetwork) {
					t.Fatalf("accepted scenario: credentials=%d posts=%d code=%s", credentials, posts, CodeOf(e))
				}
			} else {
				code := tc.code
				if code == "" {
					code = CodeProtocolChanged
				}
				if credentials != 0 || posts != 0 || !IsCode(e, code) {
					t.Fatalf("rejected scenario: credentials=%d posts=%d code=%s", credentials, posts, CodeOf(e))
				}
			}
		})
	}
}
