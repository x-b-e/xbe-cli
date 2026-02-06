package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeSearchResult struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}

func newKnowledgeSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search across resources, commands, fields, and summaries",
		Long: `Search the knowledge graph by free text.

Use this when you do not know the exact resource or command name yet.
Then pivot to:
  - xbe knowledge resource <name>
  - xbe knowledge commands --resource <name>`,
		Args: cobra.MinimumNArgs(1),
		RunE: runKnowledgeSearch,
		Example: `  # Search everything
  xbe knowledge search job

  # Limit to resources + commands
  xbe knowledge search job --kind resources,commands

  # Search only relationship edges
  xbe knowledge search trucker --kind relationships`,
	}
	cmd.Flags().String("kind", "", "Comma-separated kinds to search (resources,commands,fields,flags,relationships,summaries,dimensions,metrics)")
	return cmd
}

func runKnowledgeSearch(cmd *cobra.Command, args []string) error {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("query is required")
	}
	kinds, err := validateCSVEnum(
		"--kind",
		getStringFlag(cmd, "kind"),
		allowedValues("resources", "commands", "fields", "flags", "relationships", "summaries", "dimensions", "metrics"),
	)
	if err != nil {
		return err
	}
	includeAll := len(kinds) == 0
	kindSet := map[string]bool{}
	for _, kind := range kinds {
		kindSet[kind] = true
	}

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	pattern := "%" + query + "%"
	ctx := context.Background()
	results := make([]knowledgeSearchResult, 0, 128)

	shouldSearch := func(kind string) bool {
		if includeAll {
			return true
		}
		return kindSet[kind]
	}

	if shouldSearch("resources") {
		rows, err := queryContext(ctx, db, "SELECT name FROM resources WHERE name LIKE ? ORDER BY name", pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "resource", Name: name})
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if shouldSearch("commands") {
		rows, err := queryContext(ctx, db, "SELECT full_path, description FROM commands WHERE full_path LIKE ? OR description LIKE ? ORDER BY full_path", pattern, pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var name, desc string
			if err := rows.Scan(&name, &desc); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "command", Name: name, Detail: desc})
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if shouldSearch("fields") {
		rows, err := queryContext(ctx, db, "SELECT resource, name, kind FROM resource_fields WHERE name LIKE ? OR resource LIKE ? ORDER BY resource, name", pattern, pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var resource, name, kind string
			if err := rows.Scan(&resource, &name, &kind); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "field", Name: resource + "." + name, Detail: kind})
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if shouldSearch("flags") {
		rows, err := queryContext(ctx, db, `
SELECT f.name, c.full_path, f.description
FROM flags f
JOIN commands c ON c.id = f.command_id
WHERE f.name LIKE ? OR f.description LIKE ?
ORDER BY f.name, c.full_path`, pattern, pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var name, cmdPath, desc string
			if err := rows.Scan(&name, &cmdPath, &desc); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "flag", Name: name, Detail: cmdPath})
			if desc != "" {
				results = append(results, knowledgeSearchResult{Kind: "flag_desc", Name: name, Detail: desc})
			}
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if shouldSearch("relationships") {
		rows, err := queryContext(ctx, db, `
SELECT source_resource, relationship, target_resource, edge_kind
FROM resource_graph_edges
WHERE source_resource LIKE ? OR target_resource LIKE ? OR relationship LIKE ?
ORDER BY source_resource, relationship`, pattern, pattern, pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var source, rel, target, kind string
			if err := rows.Scan(&source, &rel, &target, &kind); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "relationship", Name: source + "." + rel, Detail: target + " (" + kind + ")"})
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if shouldSearch("summaries") {
		rows, err := queryContext(ctx, db, "SELECT DISTINCT summary_resource FROM summary_resource_targets WHERE summary_resource LIKE ? ORDER BY summary_resource", pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "summary", Name: name})
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if shouldSearch("dimensions") {
		rows, err := queryContext(ctx, db, "SELECT summary_resource, name, kind FROM summary_dimensions WHERE summary_resource LIKE ? OR name LIKE ? ORDER BY summary_resource, name", pattern, pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var summary, name, kind string
			if err := rows.Scan(&summary, &name, &kind); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "summary_dimension", Name: summary + "." + name, Detail: kind})
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if shouldSearch("metrics") {
		rows, err := queryContext(ctx, db, "SELECT summary_resource, name FROM summary_metrics WHERE summary_resource LIKE ? OR name LIKE ? ORDER BY summary_resource, name", pattern, pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		defer rows.Close()
		for rows.Next() {
			var summary, name string
			if err := rows.Scan(&summary, &name); err != nil {
				return checkDBError(err, dbPath)
			}
			results = append(results, knowledgeSearchResult{Kind: "summary_metric", Name: summary + "." + name})
		}
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No matches found.")
		return nil
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Kind == results[j].Kind {
			return results[i].Name < results[j].Name
		}
		return results[i].Kind < results[j].Kind
	})

	limit := getIntFlag(cmd, "limit")
	offset := getIntFlag(cmd, "offset")
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = len(results)
	}
	start := offset
	if start > len(results) {
		start = len(results)
	}
	end := start + limit
	if end > len(results) {
		end = len(results)
	}
	results = results[start:end]

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "KIND\tNAME\tDETAIL")
	for _, row := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\n", row.Kind, row.Name, row.Detail)
	}
	return w.Flush()
}
