package uploads

import (
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const defaultMetaCallback = "https://publisher.meruw.dev/auth/meta/"
const callbackFile = "meta-callback.txt"

var dnsLabel = regexp.MustCompile("^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$")
var pathSegment = regexp.MustCompile("^[a-zA-Z0-9._~-]+$")

// Accept a canonical HTTPS directory URI. Browsers and Meta must see exactly
// the URI the desktop app uses, including the port, path and trailing slash.
func validateMetaCallback(raw string) (string, error) {
	invalid := errors.New("enter an HTTPS callback URL on your domain with a trailing slash, without a query or fragment")
	if raw == "" || len(raw) > 2048 {
		return "", invalid
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Host == "" || u.RawQuery != "" || u.ForceQuery || strings.ContainsAny(raw, "#%\\") || !strings.HasSuffix(u.Path, "/") {
		return "", invalid
	}
	host := strings.ToLower(u.Hostname())
	labels := strings.Split(host, ".")
	if len(host) > 253 || len(labels) < 2 || net.ParseIP(host) != nil || labels[len(labels)-1] == "localhost" {
		return "", invalid
	}
	for _, label := range labels {
		if !dnsLabel.MatchString(label) {
			return "", invalid
		}
	}
	// Numeric last labels can be interpreted as IP addresses by a browser.
	tld := labels[len(labels)-1]
	if tld[0] < 'a' || tld[0] > 'z' {
		return "", invalid
	}
	for _, part := range strings.Split(strings.Trim(u.Path, "/"), "/") {
		if u.Path != "/" && (!pathSegment.MatchString(part) || part == "." || part == "..") {
			return "", invalid
		}
	}
	if strings.Contains(u.Path, "//") || strings.HasSuffix(u.Host, ":") {
		return "", invalid
	}
	port := u.Port()
	if port != "" {
		number, e := strconv.Atoi(port)
		if e != nil || number < 1 || number > 65535 {
			return "", invalid
		}
		if number != 443 {
			host = net.JoinHostPort(host, strconv.Itoa(number))
		}
	}
	return "https://" + host + u.Path, nil
}

func MetaCallbackURL() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "CATsoup Publisher", callbackFile))
	if errors.Is(err, os.ErrNotExist) {
		return defaultMetaCallback, nil
	}
	if err != nil {
		return "", errors.New("could not read the Meta callback setting")
	}
	return validateMetaCallback(string(raw))
}
func SetMetaCallbackURL(raw string) (string, error) {
	value, err := validateMetaCallback(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	if err = saveConfiguration(callbackFile, value); err != nil {
		return "", err
	}
	return value, nil
}
