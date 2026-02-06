package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeMetapathRow struct {
	Target      string `json:"target"`
	PathKind    string `json:"path_kind"`
	SharedCount int    `json:"shared_features"`
}

func newKnowledgeMetapathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metapath <resource>",
		Short: "Show similarity via shared features (metapaths)",
		Long: `Show resource similarity via shared feature paths.

Use this when expanding exploration beyond direct relationships. Similarity can
come from shared command fields, summary dimensions/metrics, or filter targets.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runKnowledgeMetapath,
		Example: `  # Shared command-field similarity
  xbe knowledge metapath jobs --kind command_field

  # Compare via shared summary dimensions
  xbe knowledge metapath transport-summaries --kind summary_dimension`,
	}
	cmd.Flags().String("kind", "", "Filter by feature kind (command_field, summary_dimension, summary_metric, filter_target)")
	return cmd
}

func runKnowledgeMetapath(cmd *cobra.Command, args []string) error {
	rawResource := strings.TrimSpace(args[0])
	if err := ensureNotEmpty(rawResource, "resource"); err != nil {
		return err
	}
	kind, err := validateEnum(
		"--kind",
		getStringFlag(cmd, "kind"),
		allowedValues("command_field", "summary_dimension", "summary_metric", "filter_target"),
	)
	if err != nil {
		return err
	}

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	resource, err := normalizeKnowledgeResourceArg(cmd, db, dbPath, rawResource, "resource")
	if err != nil {
		return err
	}

	ctx := context.Background()
	argsSQL := []any{resource}
	querySQL := `
SELECT target_resource, path_kind, shared_features
FROM resource_metapath_similarity
WHERE source_resource = ?`

	if kind != "" {
		querySQL += " AND path_kind = ?"
		argsSQL = append(argsSQL, kind)
	}

	querySQL += " ORDER BY shared_features DESC, target_resource LIMIT ? OFFSET ?"
	limit := getIntFlag(cmd, "limit")
	offset := getIntFlag(cmd, "offset")
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	argsSQL = append(argsSQL, limit, offset)

	rows, err := queryContext(ctx, db, querySQL, argsSQL...)
	if err != nil {
		return checkDBError(err, dbPath)
	}
	defer rows.Close()

	results := []knowledgeMetapathRow{}
	for rows.Next() {
		var target, pathKind string
		var shared int
		if err := rows.Scan(&target, &pathKind, &shared); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeMetapathRow{Target: target, PathKind: pathKind, SharedCount: shared})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No metapath matches found for %s.\n", resource)
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "TARGET\tPATH_KIND\tSHARED")
	for _, row := range results {
		fmt.Fprintf(w, "%s\t%s\t%d\n", row.Target, row.PathKind, row.SharedCount)
	}
	return w.Flush()
}
