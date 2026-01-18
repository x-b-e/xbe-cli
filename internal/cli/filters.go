package cli

import (
	"net/url"
	"strings"
)

func setFilterIfPresent(query url.Values, key, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		query.Set(key, value)
	}
}
