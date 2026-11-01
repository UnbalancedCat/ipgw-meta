package cascollect

import (
	"net/url"
	"regexp"
	"strings"
)

var hostPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)*\.invalid$`)
var acidPattern = regexp.MustCompile(`^[0-9]{1,10}$`)
var acidBodyPattern = regexp.MustCompile(`(?:[?&]|\b)ac_id=([0-9]+)`)

func pathOK(p string, absolute bool) bool {
	if len(p) > 256 || p == "" || strings.ContainsAny(p, `\%`) || strings.Contains(p, "//") || absolute && !strings.HasPrefix(p, "/") {
		return false
	}
	for _, c := range p {
		if c <= 32 || c >= 127 {
			return false
		}
	}
	for _, s := range strings.Split(p, "/") {
		if s == "." || s == ".." {
			return false
		}
	}
	return true
}
func parseURL(raw string, absolute, query bool) (*url.URL, bool) {
	if len(raw) == 0 || len(raw) > 2048 || strings.ContainsAny(raw, "\\#\r\n\t ") {
		return nil, false
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Opaque != "" || u.Fragment != "" || u.RawFragment != "" || u.RawPath != "" || !query && (u.RawQuery != "" || u.ForceQuery) {
		return nil, false
	}
	if absolute || u.IsAbs() || u.Host != "" {
		host := strings.ToLower(u.Hostname())
		if u.Scheme != "https" && u.Scheme != "http" || !(hostPattern.MatchString(host) || host == "ipgw.neu.edu.cn" || host == "pass.neu.edu.cn") || strings.Contains(u.Host, "[") {
			return nil, false
		}
		p := u.Port()
		if strings.HasSuffix(u.Host, ":") || p != "" && !(u.Scheme == "http" && p == "80") && !(u.Scheme == "https" && p == "443") {
			return nil, false
		}
		u.Host = strings.ToLower(u.Hostname())
	} else if strings.HasPrefix(raw, "//") {
		return nil, false
	}
	if !pathOK(u.Path, absolute || u.IsAbs()) {
		return nil, false
	}
	return u, true
}
func parseScope(s Scope) (*url.URL, *url.URL, *url.URL, bool) {
	c, a := parseURL(s.Captive, true, false)
	v, b := parseURL(s.Service, true, false)
	l, d := parseURL(s.CAS, true, false)
	if !a || !b || !d || l.Scheme != "https" || v.Host != c.Host {
		return nil, nil, nil, false
	}
	return c, v, l, true
}
func scriptURL(raw string, base *url.URL) (string, bool) {
	u, ok := parseURL(raw, false, false)
	if !ok {
		return "", false
	}
	u = base.ResolveReference(u)
	if u.Scheme != "https" || u.Host != base.Host || !pathOK(u.Path, true) {
		return "", false
	}
	return u.String(), true
}
func loginTarget(service, cas *url.URL, acid string) string {
	s, c := *service, *cas
	q := url.Values{}
	q.Set("ac_id", acid)
	s.RawQuery = q.Encode()
	q = url.Values{}
	q.Set("service", s.String())
	c.RawQuery = q.Encode()
	return c.String()
}
