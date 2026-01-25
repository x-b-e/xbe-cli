package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
)

func initSparseFieldFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArray(
		"fields",
		nil,
		"Sparse fieldset for the primary or related resources (e.g. --fields name,status or --fields users=name,email-address)",
	)
	cmd.PersistentFlags().String(
		"include",
		"",
		"Include direct relationships (comma-separated, no nested paths)",
	)
}

func applySparseFieldOverrides(cmd *cobra.Command) error {
	fieldsValues, err := cmd.Flags().GetStringArray("fields")
	if err != nil {
		return err
	}
	includeValue, err := cmd.Flags().GetString("include")
	if err != nil {
		return err
	}

	fieldsSet := cmd.Flags().Changed("fields")
	includeSet := cmd.Flags().Changed("include")
	if !fieldsSet && !includeSet {
		return nil
	}

	overrides := api.SparseFieldOverrides{
		FieldsSet:  fieldsSet,
		IncludeSet: includeSet,
	}

	if fieldsSet {
		primary, typed, parseErr := parseFieldsValues(fieldsValues)
		if parseErr != nil {
			return parseErr
		}
		overrides.Primary = primary
		overrides.Typed = typed
	}

	if includeSet {
		include := parseCSV(includeValue)
		for _, rel := range include {
			if strings.Contains(rel, ".") {
				return fmt.Errorf("invalid --include %q: only direct relationships are supported", rel)
			}
		}
		overrides.Include = include
	}

	cmd.SetContext(api.WithSparseFieldOverrides(cmd.Context(), overrides))
	return nil
}

func parseFieldsValues(values []string) ([]string, map[string][]string, error) {
	primary := []string{}
	typed := make(map[string][]string)

	for _, raw := range values {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "=") {
			parts := strings.SplitN(entry, "=", 2)
			resource := strings.TrimSpace(parts[0])
			fields := strings.TrimSpace(parts[1])
			if resource == "" || fields == "" {
				return nil, nil, fmt.Errorf("invalid --fields %q: expected resource=field1,field2", entry)
			}
			typed[resource] = appendUnique(typed[resource], parseCSV(fields)...)
			continue
		}
		primary = appendUnique(primary, parseCSV(entry)...)
	}

	return primary, typed, nil
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		cleaned = append(cleaned, item)
	}
	return cleaned
}

func appendUnique(items []string, additions ...string) []string {
	if len(additions) == 0 {
		return items
	}
	set := make(map[string]struct{}, len(items)+len(additions))
	for _, item := range items {
		set[item] = struct{}{}
	}
	for _, item := range additions {
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for item := range set {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}
