package cli

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
)

type resourceSpec struct {
	ServerTypes []string `json:"server_types"`
	LabelFields []string `json:"label_fields"`
	Attributes  []string `json:"attributes"`
}

type relationshipSpec struct {
	Resources []string `json:"resources"`
}

type resourceMap struct {
	Resources     map[string]resourceSpec                `json:"resources"`
	Relationships map[string]map[string]relationshipSpec `json:"relationships"`
}

type sparseSelection struct {
	Fields           []string
	Primary          []string
	Typed            map[string][]string
	Include          []string
	RelationLabels   map[string]map[string][]string
	RelationIDFields map[string]string
}

//go:embed resource_map.json
var resourceMapJSON []byte

var (
	loadedResourceMap resourceMap
	resourceMapErr    error
	resourceMapOnce   sync.Once
)

func loadResourceMap() (resourceMap, error) {
	resourceMapOnce.Do(func() {
		if err := json.Unmarshal(resourceMapJSON, &loadedResourceMap); err != nil {
			resourceMapErr = err
			return
		}
		if loadedResourceMap.Resources == nil {
			loadedResourceMap.Resources = map[string]resourceSpec{}
		}
		if loadedResourceMap.Relationships == nil {
			loadedResourceMap.Relationships = map[string]map[string]relationshipSpec{}
		}
	})
	return loadedResourceMap, resourceMapErr
}

func fieldsHelpForResource(resource string) string {
	resourceMap, err := loadResourceMap()
	if err != nil {
		return ""
	}
	spec, ok := resourceMap.Resources[resource]
	if !ok {
		return ""
	}
	lines := []string{
		"Available fields:",
	}
	for _, field := range spec.Attributes {
		lines = append(lines, "  - "+field)
	}
	relations := resourceMap.Relationships[resource]
	if len(relations) == 0 {
		lines = append(lines,
			"",
			"Fields usage:",
			"  --fields name,broker",
			"  List default: label fields (or ID only). Show default: all fields.",
		)
		return strings.Join(lines, "\n")
	}

	relationNames := make([]string, 0, len(relations))
	for rel := range relations {
		relationNames = append(relationNames, rel)
	}
	sort.Strings(relationNames)
	lines = append(lines, "", "Related resources:")
	for _, rel := range relationNames {
		relSpec := relations[rel]
		related := []string{}
		for _, targetResource := range relSpec.Resources {
			targetSpec, ok := resourceMap.Resources[targetResource]
			labelSuffix := ""
			if ok && len(targetSpec.LabelFields) > 0 {
				labelFields := make([]string, len(targetSpec.LabelFields))
				copy(labelFields, targetSpec.LabelFields)
				sort.Strings(labelFields)
				labelSuffix = " (label fields: " + strings.Join(labelFields, ", ") + ")"
			}
			related = append(related, targetResource+labelSuffix)
		}
		sort.Strings(related)
		lines = append(lines, "  "+rel+" -> "+strings.Join(related, "; ")+"; adds "+rel+"-id")
	}
	lines = append(lines,
		"",
		"Fields usage:",
		"  --fields name,broker",
		"  List default: label fields (or ID only). Show default: all fields.",
		"  Relationships add <rel>-id automatically.",
	)
	return strings.Join(lines, "\n")
}

func fieldsHelpForCommand(cmd *cobra.Command) string {
	resource, ok := resourceForSparseFields(cmd)
	if !ok {
		return ""
	}
	return fieldsHelpForResource(resource)
}

func resourceForSparseList(cmd *cobra.Command) (string, bool) {
	parts := strings.Fields(cmd.CommandPath())
	if len(parts) < 4 {
		return "", false
	}
	if parts[1] != "view" {
		return "", false
	}
	if parts[3] != "list" {
		return "", false
	}
	return parts[2], true
}

func initSparseFieldFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArray(
		"fields",
		nil,
		"Sparse fieldset for list/show (e.g. --fields company-name,broker)",
	)
}

func applySparseFieldOverrides(cmd *cobra.Command) error {
	fieldsValues, err := cmd.Flags().GetStringArray("fields")
	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("fields") {
		resource, ok := resourceForSparseList(cmd)
		if !ok {
			return nil
		}
		selection, ok, err := defaultListSelection(resource)
		if err != nil || !ok {
			return err
		}
		if len(selection.Primary) == 0 {
			return nil
		}
		overrides := api.SparseFieldOverrides{
			FieldsSet: true,
			Primary:   selection.Primary,
		}
		cmd.SetContext(api.WithSparseFieldOverrides(cmd.Context(), overrides))
		return nil
	}

	selection, err := selectionForCommand(cmd, fieldsValues)
	if err != nil {
		return err
	}

	overrides := api.SparseFieldOverrides{
		FieldsSet:  true,
		IncludeSet: true,
		Primary:    selection.Primary,
		Typed:      selection.Typed,
		Include:    selection.Include,
	}

	cmd.SetContext(api.WithSparseFieldOverrides(cmd.Context(), overrides))
	return nil
}

func selectionForCommand(cmd *cobra.Command, fieldsValues []string) (sparseSelection, error) {
	resource, ok := resourceForSparseFields(cmd)
	if !ok {
		return sparseSelection{}, fmt.Errorf("--fields is only supported on view <resource> list/show commands")
	}
	if len(fieldsValues) == 0 {
		values, err := cmd.Flags().GetStringArray("fields")
		if err != nil {
			return sparseSelection{}, err
		}
		fieldsValues = values
	}
	return resolveSparseSelection(resource, fieldsValues)
}

func resourceForSparseFields(cmd *cobra.Command) (string, bool) {
	parts := strings.Fields(cmd.CommandPath())
	if len(parts) < 4 {
		return "", false
	}
	if parts[1] != "view" {
		return "", false
	}
	action := parts[3]
	if action != "list" && action != "show" {
		return "", false
	}
	return parts[2], true
}

func resolveSparseSelection(resource string, values []string) (sparseSelection, error) {
	resourceMap, err := loadResourceMap()
	if err != nil {
		return sparseSelection{}, fmt.Errorf("failed to load resource map: %w", err)
	}
	spec, ok := resourceMap.Resources[resource]
	if !ok {
		return sparseSelection{}, fmt.Errorf("--fields is not configured for resource %q yet", resource)
	}
	requested := parseFieldsValues(values)
	if len(requested) == 0 {
		return sparseSelection{}, fmt.Errorf("--fields requires at least one field")
	}
	relations := resourceMap.Relationships[resource]
	fields, primary, typed, include, relationLabels, relationIDFields, err := resolveSparseFields(resource, spec, relations, resourceMap.Resources, requested)
	if err != nil {
		return sparseSelection{}, err
	}
	return sparseSelection{
		Fields:           fields,
		Primary:          primary,
		Typed:            typed,
		Include:          include,
		RelationLabels:   relationLabels,
		RelationIDFields: relationIDFields,
	}, nil
}

func buildSparseRows(resp jsonAPIResponse, selection sparseSelection) []map[string]any {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]map[string]any, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := map[string]any{
			"id": resource.ID,
		}
		for _, field := range selection.Fields {
			row[field] = sparseFieldValue(resource, field, included, selection.RelationLabels, selection.RelationIDFields)
		}
		rows = append(rows, row)
	}
	return rows
}

func renderSparseListIfRequested(cmd *cobra.Command, resp jsonAPIResponse) (bool, error) {
	var selection sparseSelection
	if cmd.Flags().Changed("fields") {
		selected, err := selectionForCommand(cmd, nil)
		if err != nil {
			return false, err
		}
		selection = selected
	} else {
		resource, ok := resourceForSparseList(cmd)
		if !ok {
			return false, nil
		}
		selected, ok, err := defaultListSelection(resource)
		if err != nil || !ok {
			return false, err
		}
		selection = selected
	}
	rows := buildSparseRows(resp, selection)
	jsonOut, _ := cmd.Flags().GetBool("json")
	if jsonOut {
		return true, writeJSON(cmd.OutOrStdout(), rows)
	}
	return true, renderSparseTable(cmd, selection, rows)
}

func renderSparseShowIfRequested(cmd *cobra.Command, resp jsonAPISingleResponse) (bool, error) {
	var selection sparseSelection
	if cmd.Flags().Changed("fields") {
		selected, err := selectionForCommand(cmd, nil)
		if err != nil {
			return false, err
		}
		selection = selected
	} else {
		resource, ok := resourceForSparseFields(cmd)
		if !ok {
			return false, nil
		}
		selected, ok, err := defaultShowSelection(resource)
		if err != nil || !ok {
			return false, err
		}
		selection = selected
	}
	rows := buildSparseRows(
		jsonAPIResponse{Data: []jsonAPIResource{resp.Data}, Included: resp.Included},
		selection,
	)
	jsonOut, _ := cmd.Flags().GetBool("json")
	if jsonOut {
		if len(rows) == 0 {
			return true, writeJSON(cmd.OutOrStdout(), map[string]any{})
		}
		return true, writeJSON(cmd.OutOrStdout(), rows[0])
	}
	return true, renderSparseTable(cmd, selection, rows)
}

func defaultListSelection(resource string) (sparseSelection, bool, error) {
	resourceMap, err := loadResourceMap()
	if err != nil {
		return sparseSelection{}, false, err
	}
	spec, ok := resourceMap.Resources[resource]
	if !ok {
		return sparseSelection{}, false, nil
	}
	fields := append([]string{}, spec.LabelFields...)
	return sparseSelection{
		Fields:           fields,
		Primary:          fields,
		Typed:            map[string][]string{},
		Include:          []string{},
		RelationLabels:   map[string]map[string][]string{},
		RelationIDFields: map[string]string{},
	}, true, nil
}

func defaultShowSelection(resource string) (sparseSelection, bool, error) {
	resourceMap, err := loadResourceMap()
	if err != nil {
		return sparseSelection{}, false, err
	}
	spec, ok := resourceMap.Resources[resource]
	if !ok {
		return sparseSelection{}, false, nil
	}
	rels := resourceMap.Relationships[resource]
	requested := append([]string{}, spec.Attributes...)
	for rel := range rels {
		requested = append(requested, rel)
	}
	if len(requested) == 0 {
		return sparseSelection{
			Fields:           []string{},
			Primary:          []string{},
			Typed:            map[string][]string{},
			Include:          []string{},
			RelationLabels:   map[string]map[string][]string{},
			RelationIDFields: map[string]string{},
		}, true, nil
	}
	fields, primary, typed, include, relationLabels, relationIDFields, err := resolveSparseFields(
		resource,
		spec,
		rels,
		resourceMap.Resources,
		requested,
	)
	if err != nil {
		return sparseSelection{}, false, err
	}
	return sparseSelection{
		Fields:           fields,
		Primary:          primary,
		Typed:            typed,
		Include:          include,
		RelationLabels:   relationLabels,
		RelationIDFields: relationIDFields,
	}, true, nil
}

func renderSparseTable(cmd *cobra.Command, selection sparseSelection, rows []map[string]any) error {
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	headers := append([]string{"ID"}, selection.Fields...)
	fmt.Fprintln(writer, strings.ToUpper(strings.Join(headers, "\t")))
	for _, row := range rows {
		values := []string{formatSparseValue(row["id"])}
		for _, field := range selection.Fields {
			values = append(values, formatSparseValue(row[field]))
		}
		fmt.Fprintln(writer, strings.Join(values, "\t"))
	}
	return writer.Flush()
}

func sparseFieldValue(
	resource jsonAPIResource,
	field string,
	included map[string]jsonAPIResource,
	relationLabels map[string]map[string][]string,
	relationIDFields map[string]string,
) any {
	if relation, ok := relationIDFields[field]; ok {
		relationship, ok := resource.Relationships[relation]
		if !ok || relationship.Data == nil {
			return nil
		}
		return relationship.Data.ID
	}
	if labelsByType, ok := relationLabels[field]; ok {
		relationship, ok := resource.Relationships[field]
		if !ok || relationship.Data == nil {
			return nil
		}
		inc, ok := included[resourceKey(relationship.Data.Type, relationship.Data.ID)]
		if !ok {
			return relationship.Data.ID
		}
		labels := labelsByType[relationship.Data.Type]
		if len(labels) == 0 {
			for _, candidate := range labelsByType {
				if len(candidate) > 0 {
					labels = candidate
					break
				}
			}
		}
		if len(labels) == 0 {
			return relationship.Data.ID
		}
		for _, label := range labels {
			value := stringAttr(inc.Attributes, label)
			if value != "" {
				return value
			}
		}
		return relationship.Data.ID
	}
	return resource.Attributes[field]
}

func formatSparseValue(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func resolveSparseFields(
	resourceType string,
	spec resourceSpec,
	relations map[string]relationshipSpec,
	resources map[string]resourceSpec,
	requested []string,
) ([]string, []string, map[string][]string, []string, map[string]map[string][]string, map[string]string, error) {
	fields := []string{}
	primary := []string{}
	typed := make(map[string][]string)
	include := []string{}
	relationLabels := make(map[string]map[string][]string)
	relationIDFields := make(map[string]string)
	seen := make(map[string]bool)
	allowedPrimary := makeSet(spec.Attributes)

	for _, rawField := range requested {
		field := normalizeFieldName(rawField)
		if field == "" {
			continue
		}
		if !seen[field] {
			fields = append(fields, field)
			seen[field] = true
		}

		if strings.Contains(field, ".") {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("field %q is not supported; request related resources directly", field)
		}
		if relSpec, ok := relations[field]; ok {
			primary = appendUnique(primary, field)
			include = appendUnique(include, field)
			relID := field + "-id"
			if !seen[relID] {
				fields = append(fields, relID)
				seen[relID] = true
			}
			relationIDFields[relID] = field
			labelsByType := relationLabels[field]
			if labelsByType == nil {
				labelsByType = make(map[string][]string)
			}
			for _, targetResource := range relSpec.Resources {
				targetSpec, ok := resources[targetResource]
				if !ok {
					continue
				}
				serverTypes := targetSpec.ServerTypes
				if len(serverTypes) == 0 {
					serverTypes = []string{targetResource}
				}
				for _, serverType := range serverTypes {
					labelsByType[serverType] = targetSpec.LabelFields
					if contains(spec.ServerTypes, serverType) {
						primary = appendUnique(primary, targetSpec.LabelFields...)
					} else {
						typed[serverType] = appendUnique(typed[serverType], targetSpec.LabelFields...)
					}
				}
			}
			if len(labelsByType) > 0 {
				relationLabels[field] = labelsByType
			}
			continue
		}
		if !allowedPrimary[field] {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("field %q is not allowed for this resource", field)
		}
		primary = appendUnique(primary, field)
	}

	return fields, primary, typed, include, relationLabels, relationIDFields, nil
}

func parseFieldsValues(values []string) []string {
	fields := []string{}
	for _, raw := range values {
		fields = append(fields, parseCSV(raw)...)
	}
	return fields
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

func normalizeFieldName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

func makeSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
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
