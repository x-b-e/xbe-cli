package cli

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeSummaryRow struct {
	Summary    string   `json:"summary"`
	Primaries  []string `json:"primaries,omitempty"`
	Dimensions int      `json:"dimensions"`
	Metrics    int      `json:"metrics"`
}

type knowledgeSummaryDetail struct {
	Summary    string   `json:"summary"`
	Primaries  []string `json:"primaries,omitempty"`
	Dimensions []string `json:"dimensions,omitempty"`
	Metrics    []string `json:"metrics,omitempty"`
}

func newKnowledgeSummariesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summaries",
		Short: "List summary resources and their features",
		RunE:  runKnowledgeSummaries,
		Example: `  # List summary resources
  xbe knowledge summaries

  # Show details for a summary
  xbe knowledge summaries --summary transport-summaries --details`,
	}
	cmd.Flags().String("summary", "", "Filter by summary resource")
	cmd.Flags().Bool("details", false, "Include dimensions and metrics")
	return cmd
}

func runKnowledgeSummaries(cmd *cobra.Command, _ []string) error {
	summaryFilter := strings.TrimSpace(getStringFlag(cmd, "summary"))
	showDetails := getBoolFlag(cmd, "details")

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()

	if showDetails {
		return runKnowledgeSummariesDetail(cmd, ctx, db, dbPath, summaryFilter)
	}

	args := []any{}
	querySQL := `
SELECT s.summary_resource,
       GROUP_CONCAT(DISTINCT s.primary_resource),
       (SELECT COUNT(*) FROM summary_dimensions d WHERE d.summary_resource = s.summary_resource),
       (SELECT COUNT(*) FROM summary_metrics m WHERE m.summary_resource = s.summary_resource)
FROM summary_resource_targets s
WHERE 1=1`

	if summaryFilter != "" {
		querySQL += " AND s.summary_resource = ?"
		args = append(args, summaryFilter)
	}

	querySQL += " GROUP BY s.summary_resource ORDER BY s.summary_resource LIMIT ? OFFSET ?"
	limit := getIntFlag(cmd, "limit")
	offset := getIntFlag(cmd, "offset")
	if limit <= 0 {
		limit = 200
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

	results := []knowledgeSummaryRow{}
	for rows.Next() {
		var summary, primariesRaw string
		var dims, metrics int
		if err := rows.Scan(&summary, &primariesRaw, &dims, &metrics); err != nil {
			return checkDBError(err, dbPath)
		}
		primaries := []string{}
		if primariesRaw != "" {
			primaries = strings.Split(primariesRaw, ",")
		}
		results = append(results, knowledgeSummaryRow{
			Summary:    summary,
			Primaries:  primaries,
			Dimensions: dims,
			Metrics:    metrics,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No summary resources found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "SUMMARY\tPRIMARIES\tDIMENSIONS\tMETRICS")
	for _, row := range results {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\n", row.Summary, strings.Join(row.Primaries, ", "), row.Dimensions, row.Metrics)
	}
	return w.Flush()
}

func runKnowledgeSummariesDetail(cmd *cobra.Command, ctx context.Context, db *sql.DB, dbPath string, summaryFilter string) error {
	args := []any{}
	querySQL := `
SELECT DISTINCT summary_resource
FROM summary_resource_targets
WHERE 1=1`
	if summaryFilter != "" {
		querySQL += " AND summary_resource = ?"
		args = append(args, summaryFilter)
	}
	querySQL += " ORDER BY summary_resource LIMIT ? OFFSET ?"
	limit := getIntFlag(cmd, "limit")
	offset := getIntFlag(cmd, "offset")
	if limit <= 0 {
		limit = 200
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

	results := []knowledgeSummaryDetail{}
	for rows.Next() {
		var summary string
		if err := rows.Scan(&summary); err != nil {
			return checkDBError(err, dbPath)
		}

		primRows, err := queryContext(ctx, db, "SELECT primary_resource FROM summary_resource_targets WHERE summary_resource = ? ORDER BY primary_resource", summary)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		primaries, err := collectStrings(primRows)
		primRows.Close()
		if err != nil {
			return checkDBError(err, dbPath)
		}

		dimRows, err := queryContext(ctx, db, "SELECT name FROM summary_dimensions WHERE summary_resource = ? ORDER BY name", summary)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		dimensions, err := collectStrings(dimRows)
		dimRows.Close()
		if err != nil {
			return checkDBError(err, dbPath)
		}

		metricRows, err := queryContext(ctx, db, "SELECT name FROM summary_metrics WHERE summary_resource = ? ORDER BY name", summary)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		metrics, err := collectStrings(metricRows)
		metricRows.Close()
		if err != nil {
			return checkDBError(err, dbPath)
		}

		results = append(results, knowledgeSummaryDetail{
			Summary:    summary,
			Primaries:  primaries,
			Dimensions: dimensions,
			Metrics:    metrics,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No summary resources found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	for _, summary := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "Summary: %s\n", summary.Summary)
		if len(summary.Primaries) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  Primaries: %s\n", strings.Join(summary.Primaries, ", "))
		}
		if len(summary.Dimensions) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  Dimensions:")
			for _, dim := range summary.Dimensions {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", dim)
			}
		}
		if len(summary.Metrics) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  Metrics:")
			for _, metric := range summary.Metrics {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", metric)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}
