package cli

import (
	"fmt"
	"strings"
)

func normalizeCrewRateType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func parseCrewRateResourceType(value string) (string, error) {
	normalized := normalizeCrewRateType(value)
	switch normalized {
	case "laborer", "laborers":
		return "laborers", nil
	case "equipment":
		return "equipment", nil
	default:
		return "", fmt.Errorf("invalid resource type %q (expected Laborer or Equipment)", value)
	}
}

func parseCrewRateResourceClassificationType(value string) (string, error) {
	normalized := normalizeCrewRateType(value)
	switch normalized {
	case "laborclassification", "laborclassifications":
		return "labor-classifications", nil
	case "equipmentclassification", "equipmentclassifications":
		return "equipment-classifications", nil
	default:
		return "", fmt.Errorf("invalid resource classification type %q (expected LaborClassification or EquipmentClassification)", value)
	}
}

func parseCrewRateBool(value string, flagName string) (bool, error) {
	if value == "true" {
		return true, nil
	}
	if value == "false" {
		return false, nil
	}
	return false, fmt.Errorf("--%s must be true or false", flagName)
}
