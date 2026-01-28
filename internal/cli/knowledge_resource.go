package cli

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeResourceDetail struct {
	Name                          string                          `json:"name"`
	LabelFields                   []string                        `json:"label_fields,omitempty"`
	ServerTypes                   []string                        `json:"server_types,omitempty"`
	VersionChanges                bool                            `json:"version_changes"`
	VersionChangesOptionalFeature []string                        `json:"version_changes_optional_features,omitempty"`
	Fields                        []knowledgeResourceField        `json:"fields,omitempty"`
	Relationships                 []knowledgeResourceRelationship `json:"relationships,omitempty"`
	SummaryTargets                []knowledgeSummaryTarget        `json:"summary_targets,omitempty"`
	SummarySources                []knowledgeSummaryTarget        `json:"summary_sources,omitempty"`
	SummaryDimensions             []knowledgeSummaryFeature       `json:"summary_dimensions,omitempty"`
	SummaryMetrics                []knowledgeSummaryFeature       `json:"summary_metrics,omitempty"`
	Commands                      []knowledgeResourceCommand      `json:"commands,omitempty"`
}

type knowledgeResourceField struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	IsLabel bool   `json:"is_label"`
}

type knowledgeResourceRelationship struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type knowledgeSummaryTarget struct {
	Resource  string `json:"resource"`
	Condition string `json:"condition,omitempty"`
}

type knowledgeSummaryFeature struct {
	Name string `json:"name"`
	Kind string `json:"kind,omitempty"`
}

type knowledgeResourceCommand struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Verb string `json:"verb"`
}

func newKnowledgeResourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource <name>",
		Short: "Show details about a resource",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runKnowledgeResource,
		Example: `  # Show all details
  xbe knowledge resource jobs

  # Only show relationships and commands
  xbe knowledge resource jobs --sections relationships,commands`,
	}
	cmd.Flags().String("sections", "", "Comma-separated sections (fields,relationships,summaries,summary-features,commands)")
	return cmd
}

func runKnowledgeResource(cmd *cobra.Command, args []string) error {
	resourceName := strings.TrimSpace(args[0])
	if err := ensureNotEmpty(resourceName, "resource name"); err != nil {
		return err
	}

	sections := parseCSVFilter(getStringFlag(cmd, "sections"))
	sectionSet := map[string]bool{}
	if len(sections) == 0 {
		sectionSet = map[string]bool{
			"fields":           true,
			"relationships":    true,
			"summaries":        true,
			"summary-features": true,
			"commands":         true,
		}
	} else {
		for _, section := range sections {
			sectionSet[strings.ToLower(section)] = true
		}
	}

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	row := db.QueryRowContext(ctx, "SELECT label_fields, server_types, version_changes, version_changes_optional_features FROM resources WHERE name = ?", resourceName)
	var labelRaw, serverRaw string
	var versionChangesRaw sql.NullInt64
	var versionChangesFeaturesRaw sql.NullString
	if err := row.Scan(&labelRaw, &serverRaw, &versionChangesRaw, &versionChangesFeaturesRaw); err != nil {
		return checkDBError(err, dbPath)
	}

	detail := knowledgeResourceDetail{
		Name:                          resourceName,
		LabelFields:                   parseJSONList(labelRaw),
		ServerTypes:                   parseJSONList(serverRaw),
		VersionChanges:                versionChangesRaw.Valid && versionChangesRaw.Int64 == 1,
		VersionChangesOptionalFeature: parseJSONList(versionChangesFeaturesRaw.String),
	}

	if sectionSet["fields"] {
		rows, err := queryContext(ctx, db, "SELECT name, kind, is_label FROM resource_fields WHERE resource = ? ORDER BY kind, name", resourceName)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		for rows.Next() {
			var name, kind string
			var isLabel int
			if err := rows.Scan(&name, &kind, &isLabel); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			detail.Fields = append(detail.Fields, knowledgeResourceField{Name: name, Kind: kind, IsLabel: isLabel == 1})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if sectionSet["relationships"] {
		rows, err := queryContext(ctx, db, "SELECT field, target_resource FROM resource_field_targets WHERE resource = ? ORDER BY field", resourceName)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		for rows.Next() {
			var name, target string
			if err := rows.Scan(&name, &target); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			detail.Relationships = append(detail.Relationships, knowledgeResourceRelationship{Name: name, Target: target})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if sectionSet["summaries"] {
		rows, err := queryContext(ctx, db, "SELECT primary_resource, COALESCE(condition, '') FROM summary_resource_targets WHERE summary_resource = ? ORDER BY primary_resource", resourceName)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		for rows.Next() {
			var resource, condition string
			if err := rows.Scan(&resource, &condition); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			detail.SummaryTargets = append(detail.SummaryTargets, knowledgeSummaryTarget{Resource: resource, Condition: condition})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}

		rows, err = queryContext(ctx, db, "SELECT summary_resource, COALESCE(condition, '') FROM summary_resource_targets WHERE primary_resource = ? ORDER BY summary_resource", resourceName)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		for rows.Next() {
			var summary, condition string
			if err := rows.Scan(&summary, &condition); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			detail.SummarySources = append(detail.SummarySources, knowledgeSummaryTarget{Resource: summary, Condition: condition})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if sectionSet["summary-features"] {
		rows, err := queryContext(ctx, db, "SELECT name, kind FROM summary_dimensions WHERE summary_resource = ? ORDER BY name", resourceName)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		for rows.Next() {
			var name, kind string
			if err := rows.Scan(&name, &kind); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			detail.SummaryDimensions = append(detail.SummaryDimensions, knowledgeSummaryFeature{Name: name, Kind: kind})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}

		rows, err = queryContext(ctx, db, "SELECT name FROM summary_metrics WHERE summary_resource = ? ORDER BY name", resourceName)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			detail.SummaryMetrics = append(detail.SummaryMetrics, knowledgeSummaryFeature{Name: name})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if sectionSet["commands"] {
		rows, err := queryContext(ctx, db, `
SELECT c.full_path, crl.command_kind, crl.verb
FROM command_resource_links crl
JOIN commands c ON c.id = crl.command_id
WHERE crl.resource = ?
ORDER BY c.full_path`, resourceName)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		for rows.Next() {
			var path, kind, verb string
			if err := rows.Scan(&path, &kind, &verb); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			detail.Commands = append(detail.Commands, knowledgeResourceCommand{Path: path, Kind: kind, Verb: verb})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, detail)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Resource: %s\n", detail.Name)
	if len(detail.LabelFields) > 0 {
		fmt.Fprintf(out, "Label fields: %s\n", strings.Join(detail.LabelFields, ", "))
	}
	if len(detail.ServerTypes) > 0 {
		fmt.Fprintf(out, "Server types: %s\n", strings.Join(detail.ServerTypes, ", "))
	}
	fmt.Fprintf(out, "Version changes: %s\n", boolToYesNo(detail.VersionChanges))
	if len(detail.VersionChangesOptionalFeature) > 0 {
		fmt.Fprintf(out, "Version changes optional features (auto-applied): %s\n", strings.Join(detail.VersionChangesOptionalFeature, ", "))
	}
	if detail.VersionChanges {
		fmt.Fprintf(out, "Version changes usage: xbe view %s show <id> --version-changes [--json]\n", detail.Name)
	}

	if len(detail.Fields) > 0 {
		fmt.Fprintln(out, "\nFields:")
		w := newTabWriter(cmd)
		fmt.Fprintln(w, "NAME\tKIND\tLABEL")
		for _, field := range detail.Fields {
			label := ""
			if field.IsLabel {
				label = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", field.Name, field.Kind, label)
		}
		_ = w.Flush()
	}

	if len(detail.Relationships) > 0 {
		fmt.Fprintln(out, "\nRelationships:")
		w := newTabWriter(cmd)
		fmt.Fprintln(w, "NAME\tTARGET")
		for _, rel := range detail.Relationships {
			fmt.Fprintf(w, "%s\t%s\n", rel.Name, rel.Target)
		}
		_ = w.Flush()
	}

	if len(detail.SummaryTargets) > 0 || len(detail.SummarySources) > 0 {
		fmt.Fprintln(out, "\nSummary Links:")
		if len(detail.SummaryTargets) > 0 {
			fmt.Fprintln(out, "  As summary for:")
			for _, target := range detail.SummaryTargets {
				if target.Condition != "" {
					fmt.Fprintf(out, "    %s (condition: %s)\n", target.Resource, target.Condition)
				} else {
					fmt.Fprintf(out, "    %s\n", target.Resource)
				}
			}
		}
		if len(detail.SummarySources) > 0 {
			fmt.Fprintln(out, "  Summaries available:")
			for _, source := range detail.SummarySources {
				if source.Condition != "" {
					fmt.Fprintf(out, "    %s (condition: %s)\n", source.Resource, source.Condition)
				} else {
					fmt.Fprintf(out, "    %s\n", source.Resource)
				}
			}
		}
	}

	if len(detail.SummaryDimensions) > 0 || len(detail.SummaryMetrics) > 0 {
		fmt.Fprintln(out, "\nSummary Features:")
		if len(detail.SummaryDimensions) > 0 {
			fmt.Fprintln(out, "  Dimensions:")
			names := make([]string, 0, len(detail.SummaryDimensions))
			for _, dim := range detail.SummaryDimensions {
				if dim.Kind != "" {
					names = append(names, fmt.Sprintf("%s (%s)", dim.Name, dim.Kind))
				} else {
					names = append(names, dim.Name)
				}
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Fprintf(out, "    %s\n", name)
			}
		}
		if len(detail.SummaryMetrics) > 0 {
			fmt.Fprintln(out, "  Metrics:")
			names := make([]string, 0, len(detail.SummaryMetrics))
			for _, metric := range detail.SummaryMetrics {
				names = append(names, metric.Name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Fprintf(out, "    %s\n", name)
			}
		}
	}

	if len(detail.Commands) > 0 {
		fmt.Fprintln(out, "\nCommands:")
		w := newTabWriter(cmd)
		fmt.Fprintln(w, "COMMAND\tKIND\tVERB")
		for _, command := range detail.Commands {
			fmt.Fprintf(w, "%s\t%s\t%s\n", command.Path, command.Kind, command.Verb)
		}
		_ = w.Flush()
	}

	return nil
}
