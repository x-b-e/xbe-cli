package cli

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeNeighborRow struct {
	Resource                    string  `json:"resource"`
	Score                       float64 `json:"score"`
	RelationshipCount           int     `json:"relationship_count"`
	SummaryCount                int     `json:"summary_count"`
	FilterPathCount             int     `json:"filter_path_count"`
	SharedCommandFieldCount     int     `json:"shared_command_field_count"`
	SharedSummaryDimensionCount int     `json:"shared_summary_dimension_count"`
	SharedSummaryMetricCount    int     `json:"shared_summary_metric_count"`
	SharedFilterTargetCount     int     `json:"shared_filter_target_count"`
}

type knowledgeNeighborComponent struct {
	Component string `json:"component"`
	Count     int    `json:"count"`
	Detail    string `json:"detail,omitempty"`
}

type knowledgeNeighborDetail struct {
	Resource   string                       `json:"resource"`
	Score      float64                      `json:"score"`
	Components []knowledgeNeighborComponent `json:"components,omitempty"`
}

func newKnowledgeNeighborsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "neighbors <resource>",
		Short: "Rank neighborhood resources for exploration",
		Long: `Rank adjacent resources using relationship, summary, and flag-path evidence.

Use this after inspecting one resource to pick high-value next resources.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runKnowledgeNeighbors,
		Example: `  # Top neighbors
  xbe knowledge neighbors jobs --limit 20

  # Explain why neighbors are connected
  xbe knowledge neighbors jobs --explain`,
	}
	cmd.Flags().Float64("min-score", 0, "Minimum neighbor score")
	cmd.Flags().Bool("explain", false, "Include component-level evidence")
	return cmd
}

func runKnowledgeNeighbors(cmd *cobra.Command, args []string) error {
	rawResource := strings.TrimSpace(args[0])
	if err := ensureNotEmpty(rawResource, "resource"); err != nil {
		return err
	}
	minScore, _ := cmd.Flags().GetFloat64("min-score")
	includeExplain := getBoolFlag(cmd, "explain")

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
	querySQL := `
SELECT target_resource,
       score,
       relationship_count,
       summary_count,
       filter_path_count,
       shared_command_field_count,
       shared_summary_dimension_count,
       shared_summary_metric_count,
       shared_filter_target_count
FROM resource_neighbor_scores
WHERE source_resource = ? AND score >= ?
ORDER BY score DESC, target_resource
LIMIT ? OFFSET ?`
	limit := getIntFlag(cmd, "limit")
	offset := getIntFlag(cmd, "offset")
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := queryContext(ctx, db, querySQL, resource, minScore, limit, offset)
	if err != nil {
		return checkDBError(err, dbPath)
	}
	defer rows.Close()

	neighbors := []knowledgeNeighborRow{}
	for rows.Next() {
		var target string
		var score float64
		var relCount, summaryCount, filterCount, sharedFields, sharedDims, sharedMetrics, sharedFilters int
		if err := rows.Scan(&target, &score, &relCount, &summaryCount, &filterCount, &sharedFields, &sharedDims, &sharedMetrics, &sharedFilters); err != nil {
			return checkDBError(err, dbPath)
		}
		neighbors = append(neighbors, knowledgeNeighborRow{
			Resource:                    target,
			Score:                       score,
			RelationshipCount:           relCount,
			SummaryCount:                summaryCount,
			FilterPathCount:             filterCount,
			SharedCommandFieldCount:     sharedFields,
			SharedSummaryDimensionCount: sharedDims,
			SharedSummaryMetricCount:    sharedMetrics,
			SharedFilterTargetCount:     sharedFilters,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(neighbors) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No neighbors found for %s.\n", resource)
		return nil
	}

	if includeExplain {
		return renderKnowledgeNeighborsWithExplain(cmd, ctx, db, dbPath, resource, neighbors)
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, neighbors)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "NEIGHBOR\tSCORE\tREL\tSUMMARY\tFILTERS\tSHARED_FIELDS\tSHARED_DIMS\tSHARED_METRICS\tSHARED_FILTERS")
	for _, row := range neighbors {
		fmt.Fprintf(w, "%s\t%.2f\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
			row.Resource,
			row.Score,
			row.RelationshipCount,
			row.SummaryCount,
			row.FilterPathCount,
			row.SharedCommandFieldCount,
			row.SharedSummaryDimensionCount,
			row.SharedSummaryMetricCount,
			row.SharedFilterTargetCount,
		)
	}
	return w.Flush()
}

func renderKnowledgeNeighborsWithExplain(cmd *cobra.Command, ctx context.Context, db *sql.DB, dbPath string, source string, neighbors []knowledgeNeighborRow) error {
	details := make([]knowledgeNeighborDetail, 0, len(neighbors))
	for _, neighbor := range neighbors {
		rows, err := queryContext(ctx, db, `
SELECT component_kind, component_count, COALESCE(detail, '')
FROM resource_neighbor_components
WHERE source_resource = ? AND target_resource = ?
ORDER BY component_count DESC, component_kind`, source, neighbor.Resource)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		components := []knowledgeNeighborComponent{}
		for rows.Next() {
			var kind string
			var count int
			var detail string
			if err := rows.Scan(&kind, &count, &detail); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			components = append(components, knowledgeNeighborComponent{Component: kind, Count: count, Detail: detail})
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
		details = append(details, knowledgeNeighborDetail{
			Resource:   neighbor.Resource,
			Score:      neighbor.Score,
			Components: components,
		})
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, details)
	}

	for _, detail := range details {
		fmt.Fprintf(cmd.OutOrStdout(), "Neighbor: %s (score %.2f)\n", detail.Resource, detail.Score)
		if len(detail.Components) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  No component details.")
			continue
		}
		for _, component := range detail.Components {
			if component.Detail != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s (%d) - %s\n", component.Component, component.Count, component.Detail)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s (%d)\n", component.Component, component.Count)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}
