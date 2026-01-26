package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeRelationRow struct {
	Source    string `json:"source"`
	Relation  string `json:"relation"`
	Target    string `json:"target"`
	EdgeKind  string `json:"edge_kind"`
	Condition string `json:"condition,omitempty"`
}

func newKnowledgeRelationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relations",
		Short: "List relationships between resources",
		RunE:  runKnowledgeRelations,
		Example: `  # Relationships from a resource
  xbe knowledge relations --resource jobs

  # Relationships targeting a resource
  xbe knowledge relations --target brokers`,
	}
	cmd.Flags().String("resource", "", "Filter by source resource")
	cmd.Flags().String("target", "", "Filter by target resource")
	cmd.Flags().String("kind", "", "Filter by edge kind (relationship, summary)")
	return cmd
}

func runKnowledgeRelations(cmd *cobra.Command, _ []string) error {
	resource := strings.TrimSpace(getStringFlag(cmd, "resource"))
	target := strings.TrimSpace(getStringFlag(cmd, "target"))
	kind := strings.TrimSpace(getStringFlag(cmd, "kind"))

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	args := []any{}
	querySQL := `
SELECT source_resource, relationship, target_resource, edge_kind, COALESCE(condition, '')
FROM resource_graph_edges
WHERE 1=1`

	if resource != "" {
		querySQL += " AND source_resource = ?"
		args = append(args, resource)
	}
	if target != "" {
		querySQL += " AND target_resource = ?"
		args = append(args, target)
	}
	if kind != "" {
		querySQL += " AND edge_kind = ?"
		args = append(args, kind)
	}

	querySQL += " ORDER BY source_resource, relationship, target_resource LIMIT ? OFFSET ?"
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

	results := []knowledgeRelationRow{}
	for rows.Next() {
		var source, rel, targetRes, edgeKind, condition string
		if err := rows.Scan(&source, &rel, &targetRes, &edgeKind, &condition); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeRelationRow{
			Source:    source,
			Relation:  rel,
			Target:    targetRes,
			EdgeKind:  edgeKind,
			Condition: condition,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No relationships found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "SOURCE\tRELATION\tTARGET\tKIND\tCONDITION")
	for _, row := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", row.Source, row.Relation, row.Target, row.EdgeKind, row.Condition)
	}
	return w.Flush()
}
