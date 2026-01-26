package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeFilterRow struct {
	Command     string `json:"command"`
	Resource    string `json:"resource"`
	Flag        string `json:"flag"`
	Path        string `json:"path"`
	Target      string `json:"target"`
	TargetField string `json:"target_field,omitempty"`
	HopCount    int    `json:"hop_count"`
	MatchKind   string `json:"match_kind"`
	Modifier    string `json:"modifier,omitempty"`
}

func newKnowledgeFiltersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filters",
		Short: "Show inferred filter paths for list commands",
		RunE:  runKnowledgeFilters,
		Example: `  # Filter paths for a resource
  xbe knowledge filters --resource jobs

  # Filter paths for a command
  xbe knowledge filters --command "view jobs list"

  # Filter paths for a specific flag
  xbe knowledge filters --flag broker`,
	}
	cmd.Flags().String("resource", "", "Filter by resource")
	cmd.Flags().String("command", "", "Filter by command path (substring)")
	cmd.Flags().String("flag", "", "Filter by flag name")
	return cmd
}

func runKnowledgeFilters(cmd *cobra.Command, _ []string) error {
	resource := strings.TrimSpace(getStringFlag(cmd, "resource"))
	commandQuery := strings.TrimSpace(getStringFlag(cmd, "command"))
	flag := strings.TrimSpace(getStringFlag(cmd, "flag"))

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()

	commandID := ""
	commandPath := ""
	if commandQuery != "" {
		pattern := "%" + commandQuery + "%"
		rows, err := queryContext(ctx, db, "SELECT id, full_path FROM commands WHERE full_path LIKE ? ORDER BY full_path", pattern)
		if err != nil {
			return checkDBError(err, dbPath)
		}
		matches := []string{}
		ids := []string{}
		for rows.Next() {
			var id, path string
			if err := rows.Scan(&id, &path); err != nil {
				rows.Close()
				return checkDBError(err, dbPath)
			}
			matches = append(matches, path)
			ids = append(ids, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return checkDBError(err, dbPath)
		}
		selected, err := ensureSingleMatch(matches, "command")
		if err != nil {
			return err
		}
		commandPath = selected
		for i, path := range matches {
			if path == selected {
				commandID = ids[i]
				break
			}
		}
	}

	args := []any{}
	querySQL := `
SELECT c.full_path,
       f.resource,
       f.flag_name,
       f.path,
       f.target_resource,
       COALESCE(f.target_field, ''),
       f.hop_count,
       f.match_kind,
       COALESCE(f.modifier, '')
FROM command_filter_paths f
JOIN commands c ON c.id = f.command_id
WHERE 1=1`

	if commandID != "" {
		querySQL += " AND f.command_id = ?"
		args = append(args, commandID)
	}
	if resource != "" {
		querySQL += " AND f.resource = ?"
		args = append(args, resource)
	}
	if flag != "" {
		querySQL += " AND f.flag_name LIKE ?"
		args = append(args, "%"+flag+"%")
	}

	querySQL += " ORDER BY c.full_path, f.flag_name LIMIT ? OFFSET ?"
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

	results := []knowledgeFilterRow{}
	for rows.Next() {
		var cmdPath, resName, flagName, path, target, targetField, matchKind, modifier string
		var hop int
		if err := rows.Scan(&cmdPath, &resName, &flagName, &path, &target, &targetField, &hop, &matchKind, &modifier); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeFilterRow{
			Command:     cmdPath,
			Resource:    resName,
			Flag:        flagName,
			Path:        path,
			Target:      target,
			TargetField: targetField,
			HopCount:    hop,
			MatchKind:   matchKind,
			Modifier:    modifier,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		if commandPath != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "No filter paths found for %s.\n", commandPath)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "No filter paths found.")
		}
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "COMMAND\tRESOURCE\tFLAG\tPATH\tTARGET\tTARGET_FIELD\tHOPS\tMATCH\tMODIFIER")
	for _, row := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			row.Command,
			row.Resource,
			row.Flag,
			row.Path,
			row.Target,
			row.TargetField,
			row.HopCount,
			row.MatchKind,
			row.Modifier,
		)
	}
	return w.Flush()
}
