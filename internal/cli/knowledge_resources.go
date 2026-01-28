package cli

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeResourceRow struct {
	Name                          string   `json:"name"`
	LabelFields                   []string `json:"label_fields,omitempty"`
	ServerTypes                   []string `json:"server_types,omitempty"`
	VersionChanges                bool     `json:"version_changes"`
	VersionChangesOptionalFeature []string `json:"version_changes_optional_features,omitempty"`
}

func newKnowledgeResourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "List resources in the knowledge base",
		RunE:  runKnowledgeResources,
		Example: `  # List all resources
  xbe knowledge resources

  # Filter resources that include a field
  xbe knowledge resources --field status

  # Filter resources that relate to brokers
  xbe knowledge resources --target brokers

  # Only resources with version changes
  xbe knowledge resources --version-changes`,
	}
	cmd.Flags().String("query", "", "Substring filter for resource names")
	cmd.Flags().String("field", "", "Only resources that define a field (attribute or relationship)")
	cmd.Flags().String("relationship", "", "Only resources with a relationship name")
	cmd.Flags().String("target", "", "Only resources with relationships targeting this resource")
	cmd.Flags().Bool("version-changes", false, "Only resources that support version changes")
	return cmd
}

func runKnowledgeResources(cmd *cobra.Command, _ []string) error {
	query := strings.TrimSpace(getStringFlag(cmd, "query"))
	field := strings.TrimSpace(getStringFlag(cmd, "field"))
	relationship := strings.TrimSpace(getStringFlag(cmd, "relationship"))
	target := strings.TrimSpace(getStringFlag(cmd, "target"))
	versionChangesOnly := getBoolFlag(cmd, "version-changes")

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	pattern := func(value string) string {
		if value == "" {
			return ""
		}
		return "%" + value + "%"
	}

	ctx := context.Background()
	querySQL := `
SELECT r.name, r.label_fields, r.server_types, r.version_changes, r.version_changes_optional_features
FROM resources r
WHERE 1=1`
	args := []any{}

	if query != "" {
		querySQL += " AND r.name LIKE ?"
		args = append(args, pattern(query))
	}
	if field != "" {
		querySQL += " AND EXISTS (SELECT 1 FROM resource_fields rf WHERE rf.resource = r.name AND rf.name LIKE ?)"
		args = append(args, pattern(field))
	}
	if relationship != "" {
		querySQL += " AND EXISTS (SELECT 1 FROM resource_field_targets rft WHERE rft.resource = r.name AND rft.field LIKE ?)"
		args = append(args, pattern(relationship))
	}
	if target != "" {
		querySQL += " AND EXISTS (SELECT 1 FROM resource_field_targets rft WHERE rft.resource = r.name AND rft.target_resource LIKE ?)"
		args = append(args, pattern(target))
	}
	if versionChangesOnly {
		querySQL += " AND r.version_changes = 1"
	}

	querySQL += " ORDER BY r.name LIMIT ? OFFSET ?"
	limit := getIntFlag(cmd, "limit")
	offset := getIntFlag(cmd, "offset")
	if limit <= 0 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)

	rows, err := queryContext(ctx, db, querySQL, args...)
	if err != nil {
		return checkDBError(err, dbPath)
	}
	defer rows.Close()

	results := []knowledgeResourceRow{}
	for rows.Next() {
		var name, labelFieldsRaw, serverTypesRaw string
		var versionChangesRaw sql.NullInt64
		var versionChangesFeaturesRaw sql.NullString
		if err := rows.Scan(&name, &labelFieldsRaw, &serverTypesRaw, &versionChangesRaw, &versionChangesFeaturesRaw); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeResourceRow{
			Name:                          name,
			LabelFields:                   parseJSONList(labelFieldsRaw),
			ServerTypes:                   parseJSONList(serverTypesRaw),
			VersionChanges:                versionChangesRaw.Valid && versionChangesRaw.Int64 == 1,
			VersionChangesOptionalFeature: parseJSONList(versionChangesFeaturesRaw.String),
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No resources found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "RESOURCE\tLABEL_FIELDS\tSERVER_TYPES\tVERSION_CHANGES\tVERSION_CHANGES_FEATURES")
	for _, row := range results {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\n",
			row.Name,
			joinOrDash(row.LabelFields),
			joinOrDash(row.ServerTypes),
			boolToYesNo(row.VersionChanges),
			joinOrDash(row.VersionChangesOptionalFeature),
		)
	}
	return w.Flush()
}
