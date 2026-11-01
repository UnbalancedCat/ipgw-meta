package cascollect

import (
	"net/url"
	"regexp"
	"strings"
)

var discoveryKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

func normalizedDiscoveryKey(key string) string {
	return strings.ReplaceAll(strings.ToLower(key), "_", "")
}

// discoveryQueryGuard inspects decoded names only. Unknown values are neither
// interpreted nor forwarded. Shape errors take precedence across the whole map.
func discoveryQueryGuard(query url.Values) string {
	for key := range query {
		if key == "ac_id" {
			continue
		}
		if !discoveryKeyPattern.MatchString(key) || normalizedDiscoveryKey(key) == "acid" {
			return "discovery_query_key_rejected"
		}
	}
	for key := range query {
		switch normalizedDiscoveryKey(key) {
		case "ticket", "tgt", "lt", "execution", "eventid", "code", "state", "session", "sessionid", "sid", "jsessionid", "token", "accesstoken", "idtoken", "refreshtoken", "samlrequest", "samlresponse", "relaystate", "assertion", "cookie", "authorization", "password", "passwd", "pwd", "credential", "credentials", "otp", "smscode":
			return "discovery_query_auth_marker"
		}
	}
	return ""
}
