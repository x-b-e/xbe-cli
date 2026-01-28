package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
)

var irregularSingulars = map[string]string{
	"people":   "person",
	"children": "child",
	"men":      "man",
	"women":    "woman",
	"feet":     "foot",
	"teeth":    "tooth",
	"geese":    "goose",
	"mice":     "mouse",
	"oxen":     "ox",
}

var uncountableSingulars = map[string]bool{
	"equipment": true,
	"news":      true,
	"series":    true,
	"species":   true,
	"data":      true,
}

func versionChangesRequested(cmd *cobra.Command) bool {
	flag := cmd.Flags().Lookup("version-changes")
	if flag == nil {
		return false
	}
	value, err := cmd.Flags().GetBool("version-changes")
	if err != nil {
		return false
	}
	return value
}

func applyVersionChangesContext(cmd *cobra.Command) error {
	if !versionChangesRequested(cmd) {
		return nil
	}
	resource, ok := resourceForSparseFields(cmd)
	if !ok {
		return fmt.Errorf("--version-changes is only supported on view <resource> list/show commands")
	}
	spec, ok := versionChangesResourceSpec(resource)
	if !ok {
		return fmt.Errorf("resource %q is not configured for version changes yet", resource)
	}
	if !spec.VersionChanges {
		return fmt.Errorf("version changes are not available for %q (use 'xbe knowledge resources --version-changes' to list supported resources)", resource)
	}

	metaKey := versionChangesMetaKey(resource, spec)
	if metaKey == "" {
		return fmt.Errorf("failed to resolve version changes meta key for %q", resource)
	}

	params := map[string]string{
		fmt.Sprintf("meta[%s]", metaKey): "version_changes",
	}
	if len(spec.VersionChangesOptionalFeatures) > 0 {
		params["meta[optional-feature]"] = strings.Join(spec.VersionChangesOptionalFeatures, ",")
	}

	cmd.SetContext(api.WithMetaOverrides(cmd.Context(), params))
	return nil
}

func versionChangesResourceSpec(resource string) (resourceSpec, bool) {
	resourceMap, err := loadResourceMap()
	if err != nil {
		return resourceSpec{}, false
	}
	spec, ok := resourceMap.Resources[resource]
	return spec, ok
}

func versionChangesMetaKey(resource string, spec resourceSpec) string {
	resourceType := resource
	if len(spec.ServerTypes) > 0 {
		resourceType = spec.ServerTypes[0]
	}
	if resourceType == "" {
		return ""
	}
	parts := strings.Split(resourceType, "-")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	last = singularizeWord(last)
	parts[len(parts)-1] = last
	return strings.Join(parts, "_")
}

func singularizeWord(word string) string {
	if word == "" {
		return word
	}
	if uncountableSingulars[word] {
		return word
	}
	if singular, ok := irregularSingulars[word]; ok {
		return singular
	}
	if strings.HasSuffix(word, "ies") && len(word) > 3 {
		return word[:len(word)-3] + "y"
	}
	if strings.HasSuffix(word, "ses") || strings.HasSuffix(word, "xes") ||
		strings.HasSuffix(word, "zes") || strings.HasSuffix(word, "ches") ||
		strings.HasSuffix(word, "shes") {
		return word[:len(word)-2]
	}
	if strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") && len(word) > 1 {
		return word[:len(word)-1]
	}
	return word
}
