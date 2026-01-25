package cli

import (
	"fmt"
	"strings"
)

func parseJobScheduleShiftBool(value string, flagName string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("--%s must be true or false", flagName)
	}
}
