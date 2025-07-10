package http

import (
	"net/url"
	"regexp"
)

var urlSchemaPattern = regexp.MustCompile(`^([a-z0-9+\-.]+://)|mailto:|news:`)

func ParseUrl(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	if (u.Scheme == "" || u.Opaque == "") && u.Hostname() == "" && !urlSchemaPattern.MatchString(u.Path) {
		u, err = url.Parse("https://" + raw)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}
