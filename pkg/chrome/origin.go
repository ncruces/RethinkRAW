package chrome

import (
	"net/url"
	"strings"
)

func origin(u string) string {
	url, err := url.Parse(u)
	if err != nil {
		return "null"
	}
	switch url.Scheme {
	case "blob":
		return origin(url.Opaque)
	case "http", "ws":
		return url.Scheme + "://" + strings.TrimSuffix(url.Host, ":80")
	case "https", "wss":
		return url.Scheme + "://" + strings.TrimSuffix(url.Host, ":443")
	default:
		return "null"
	}
}
