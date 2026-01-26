package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeFieldRow struct {
	Resource string `json:"resource"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	IsLabel  bool   `json:"is_label"`
}

func newKnowledgeFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "List fields and their resources",
		RunE:  runKnowledgeFields,
		Example: `  # Fields for a resource
  xbe knowledge fields --resource jobs

  # Search field names
  xbe knowledge fields --query status`,
	}
	cmd.Flags().String("resource", "", "Only fields for a resource")
	cmd.Flags().String("query", "", "Substring filter for field or resource")
	cmd.Flags().String("kind", "", "Filter by kind (attribute, relationship)")
	return cmd
}

func runKnowledgeFields(cmd *cobra.Command, _ []string) error {
	resource := strings.TrimSpace(getStringFlag(cmd, "resource"))
	query := strings.TrimSpace(getStringFlag(cmd, "query"))
	kind := strings.TrimSpace(getStringFlag(cmd, "kind"))

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	args := []any{}
	querySQL := "SELECT resource, name, kind, is_label FROM resource_fields WHERE 1=1"

	if resource != "" {
		querySQL += " AND resource = ?"
		args = append(args, resource)
	}
	if query != "" {
		pattern := "%" + query + "%"
		querySQL += " AND (name LIKE ? OR resource LIKE ?)"
		args = append(args, pattern, pattern)
	}
	if kind != "" {
		querySQL += " AND kind = ?"
		args = append(args, kind)
	}

	querySQL += " ORDER BY resource, name LIMIT ? OFFSET ?"
	limit := getIntFlag(cmd, "limit")
	offset := getIntFlag(cmd, "offset")
	if limit <= 0 {
		limit = 500
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

	results := []knowledgeFieldRow{}
	for rows.Next() {
		var resName, fieldName, fieldKind string
		var isLabel int
		if err := rows.Scan(&resName, &fieldName, &fieldKind, &isLabel); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeFieldRow{
			Resource: resName,
			Name:     fieldName,
			Kind:     fieldKind,
			IsLabel:  isLabel == 1,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No fields found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "RESOURCE\tFIELD\tKIND\tLABEL")
	for _, row := range results {
		label := ""
		if row.IsLabel {
			label = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", row.Resource, row.Name, row.Kind, label)
	}
	return w.Flush()
}
