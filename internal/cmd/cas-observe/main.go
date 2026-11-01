// cas-observe reports synthetic behavior without asserting a key-selection policy.
// It has no live mode, input files, persisted keys, or socket-capable transport.
package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/caslegacy"
	"io"
	"net/http"
	"os"
	"strings"

	ipgw "github.com/UnbalancedCat/ipgw-meta"
)

type fakeTransport func(*http.Request) (*http.Response, error)

func (f fakeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type row struct {
	Case           string   `json:"case"`
	ParserSelected string   `json:"parser_selected"`
	SDKSelected    string   `json:"sdk_selected"`
	ScriptRequests []string `json:"script_requests"`
	Requests       int      `json:"synthetic_requests"`
	Credentials    int      `json:"synthetic_credential_calls"`
	Posts          int      `json:"synthetic_posts"`
	Unexpected     int      `json:"unexpected_requests"`
	Code           string   `json:"sdk_error_code"`
}

func main() {
	if len(os.Args) != 1 {
		os.Exit(2)
	}
	if err := observe(); err != nil {
		_, _ = os.Stderr.WriteString("synthetic observation failed\n")
		os.Exit(1)
	}
}
func observe() error {
	keys := map[string]*rsa.PrivateKey{}
	encoded := map[string]string{}
	for _, label := range []string{"A", "B"} {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return err
		}
		keys[label], encoded[label] = key, base64.StdEncoding.EncodeToString(der)
	}
	decl := func(label string) string { return "var publicKeyStr='" + encoded[label] + "';" }
	cases := []struct {
		name, inline string
		scripts      []string
	}{
		{"single", decl("A"), nil},
		{"duplicate_same", decl("A") + decl("A"), nil},
		{"distinct_ab", decl("A") + decl("B"), nil},
		{"distinct_ba", decl("B") + decl("A"), nil},
		{"identifier_substring", "var notpublicKeyStr='" + encoded["A"] + "';", nil},
		{"page_a_script_b", decl("A"), []string{"B"}},
		{"scripts_ab", "", []string{"A", "B"}},
		{"scripts_ba", "", []string{"B", "A"}},
		{"identifier_suffix", "var publicKeyStrSuffix='" + encoded["A"] + "';", nil},
		{"case_variant", "var PUBLICKEYSTR='" + encoded["A"] + "';", nil},
		{"line_comment", "// " + decl("A") + "\n", nil},
		{"block_comment", "/* " + decl("A") + " */", nil},
		{"string_literal", "var text=\"" + decl("A") + "\";", nil},
		{"template_literal", "var text=`" + decl("A") + "`;", nil},
		{"uninvoked_function", "function unused(){" + decl("A") + "}", nil},
		{"property_assignment", "window.publicKeyStr='" + encoded["A"] + "';", nil},
		{"escaped_identifier", `var public\u004beyStr='` + encoded["A"] + "';", nil},
		{"generic_key_first", "var publicKey='" + encoded["A"] + "';" + decl("B"), nil},
		{"specific_key_first", decl("B") + "var publicKey='" + encoded["A"] + "';", nil},
	}
	rows := []row{}
	for _, scenario := range cases {
		out := row{Case: scenario.name, ParserSelected: "none", SDKSelected: "none", ScriptRequests: []string{}}
		parsed := caslegacy.ExtractPublicKey([]byte(scenario.inline))
		for label, value := range encoded {
			if parsed == value {
				out.ParserSelected = label
			}
		}
		transport := fakeTransport(func(req *http.Request) (*http.Response, error) {
			out.Requests++
			response := func(code int, body, location string) (*http.Response, error) {
				headers := make(http.Header)
				if location != "" {
					headers.Set("Location", location)
				}
				return &http.Response{StatusCode: code, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			}
			switch {
			case req.URL.Host == "ipgw.neu.edu.cn" && req.URL.Path == "/cgi-bin/rad_user_info":
				return response(200, "not_online_error", "")
			case req.URL.Host == "ipgw.neu.edu.cn" && req.URL.Path == "/":
				return response(302, "", "http://ipgw.neu.edu.cn/srun_portal_pc?ac_id=15")
			case req.URL.Host == "pass.neu.edu.cn" && req.URL.Path == "/tpass/login" && req.Method == http.MethodGet:
				html := `<form id="loginForm" action="/tpass/auth"><input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution"></form><script>` + scenario.inline + `</script>`
				for _, label := range scenario.scripts {
					html += `<script src="/tpass/login-` + label + `.js"></script>`
				}
				return response(200, html, "")
			case req.URL.Host == "pass.neu.edu.cn" && strings.HasPrefix(req.URL.Path, "/tpass/login-") && req.Method == http.MethodGet:
				label := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/tpass/login-"), ".js")
				if _, ok := keys[label]; !ok {
					out.Unexpected++
					return nil, errors.New("unexpected synthetic script")
				}
				out.ScriptRequests = append(out.ScriptRequests, label)
				return response(200, decl(label), "")
			case req.URL.Host == "pass.neu.edu.cn" && req.URL.Path == "/tpass/auth" && req.Method == http.MethodPost:
				out.Posts++
				if err := req.ParseForm(); err != nil {
					return nil, err
				}
				ciphertext, err := base64.StdEncoding.DecodeString(req.PostForm.Get("rsa"))
				if err != nil {
					return nil, err
				}
				for label, key := range keys {
					plain, err := rsa.DecryptPKCS1v15(rand.Reader, key, ciphertext)
					if err == nil && string(plain) == "synthetic-usersynthetic-password" {
						out.SDKSelected = label
					}
				}
				// Stop in the fake transport; no ticket, activation, or success path.
				return response(500, "", "")
			default:
				out.Unexpected++
				return nil, errors.New("unexpected synthetic request")
			}
		})
		client, err := ipgw.NewClient(ipgw.WithRoundTripper(transport))
		if err != nil {
			return err
		}
		_, err = client.Login(context.Background(), ipgw.LoginRequest{
			Method: ipgw.AuthMethodPassword, ExpectedUsername: "synthetic-user",
			Credentials: ipgw.CredentialProviderFunc(func(context.Context, ipgw.CredentialRequest) (ipgw.Credential, error) {
				out.Credentials++
				return ipgw.Credential{Password: "synthetic-password"}, nil
			}),
		})
		if err == nil || out.Unexpected != 0 {
			return errors.New("unexpected synthetic observation path")
		}
		out.Code = string(ipgw.CodeOf(err))
		rows = append(rows, out)
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Kind            string `json:"evidence_kind"`
		Network         int    `json:"network_requests"`
		RealCredentials int    `json:"real_credential_reads"`
		Rows            []row  `json:"observations"`
	}{Kind: "synthetic_behavior_observation_not_key_selection_contract", Rows: rows})
}
