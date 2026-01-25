package cli

import (
	"fmt"
	"strings"
)

func normalizeResourceUnavailabilityType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func parseResourceUnavailabilityType(value string) (string, error) {
	normalized := normalizeResourceUnavailabilityType(value)
	switch normalized {
	case "user", "users":
		return "users", nil
	case "equipment":
		return "equipment", nil
	case "trailer", "trailers":
		return "trailers", nil
	case "tractor", "tractors":
		return "tractors", nil
	default:
		return "", fmt.Errorf("invalid resource type %q (expected User, Equipment, Trailer, or Tractor)", value)
	}
}
