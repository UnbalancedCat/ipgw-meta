package authobserve

import (
	"net/netip"
	"regexp"
)

var profileName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// Command accepts only the exact private ProcessExecutor forms. It does not
// allow config overrides, profile writes, QR, switches or arbitrary flags.
func Command(args []string) string {
	if len(args) < 6 || args[0] != "--json" || args[1] != "--profile" || !profileName.MatchString(args[2]) || args[3] != "--bind-ip" {
		return ""
	}
	ip, err := netip.ParseAddr(args[4])
	if err != nil || !ip.Is4() || ip.IsUnspecified() || ip.IsMulticast() {
		return ""
	}
	rest := args[5:]
	if len(rest) == 1 && (rest[0] == "status" || rest[0] == "logout") {
		return rest[0]
	}
	if len(rest) == 2 && rest[0] == "profile" && rest[1] == "show" {
		return "profile_show"
	}
	if len(rest) == 3 && rest[0] == "login" && rest[1] == "--method" && rest[2] == "password" {
		return "login"
	}
	return ""
}
