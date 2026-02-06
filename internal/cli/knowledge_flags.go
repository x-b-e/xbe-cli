package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeFlagRow struct {
	Flag      string `json:"flag"`
	Command   string `json:"command"`
	Resource  string `json:"resource,omitempty"`
	Relation  string `json:"relation,omitempty"`
	Field     string `json:"field,omitempty"`
	MatchKind string `json:"match_kind,omitempty"`
	Modifier  string `json:"modifier,omitempty"`
}

func newKnowledgeFlagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flags",
		Short: "List flags and their field semantics",
		Long: `List CLI flags and their mapped field semantics.

This helps explain whether a flag filters a direct attribute, traverses a
relationship, or remains unmapped.`,
		RunE: runKnowledgeFlags,
		Example: `  # Flags for a specific command
  xbe knowledge flags --command "view jobs list"

  # Unmapped flags
  xbe knowledge flags --mapped=false`,
	}
	cmd.Flags().String("query", "", "Substring filter for flag name or description")
	cmd.Flags().String("command", "", "Filter by command path (substring match)")
	cmd.Flags().String("resource", "", "Filter by resource")
	cmd.Flags().String("mapped", "", "Filter by mapping status (true/false)")
	return cmd
}

func runKnowledgeFlags(cmd *cobra.Command, _ []string) error {
	query := strings.TrimSpace(getStringFlag(cmd, "query"))
	commandFilter := strings.TrimSpace(getStringFlag(cmd, "command"))
	resource := strings.TrimSpace(getStringFlag(cmd, "resource"))
	mapped := strings.TrimSpace(getStringFlag(cmd, "mapped"))

	mappedFilter := ""
	if mapped != "" {
		switch strings.ToLower(mapped) {
		case "true", "yes", "1":
			mappedFilter = "mapped"
		case "false", "no", "0":
			mappedFilter = "unmapped"
		default:
			return fmt.Errorf("invalid --mapped value: %s", mapped)
		}
	}

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	if resource != "" {
		resolvedResource, err := normalizeKnowledgeResourceFlag(cmd, db, dbPath, resource, "--resource")
		if err != nil {
			return err
		}
		resource = resolvedResource
	}

	ctx := context.Background()
	args := []any{}
	querySQL := `
SELECT f.name,
       c.full_path,
       COALESCE(cfl.resource, ''),
       COALESCE(cfl.relation, ''),
       COALESCE(cfl.field, ''),
       COALESCE(cfl.match_kind, ''),
       COALESCE(cfl.modifier, '')
FROM flags f
JOIN commands c ON c.id = f.command_id
LEFT JOIN command_field_links cfl
  ON cfl.command_id = f.command_id AND cfl.flag_name = f.name
LEFT JOIN command_resource_links crl
  ON crl.command_id = f.command_id
WHERE 1=1`

	if query != "" {
		pattern := "%" + query + "%"
		querySQL += " AND (f.name LIKE ? OR f.description LIKE ?)"
		args = append(args, pattern, pattern)
	}
	if commandFilter != "" {
		pattern := "%" + commandFilter + "%"
		querySQL += " AND c.full_path LIKE ?"
		args = append(args, pattern)
	}
	if resource != "" {
		querySQL += " AND crl.resource = ?"
		args = append(args, resource)
	}
	if mappedFilter == "mapped" {
		querySQL += " AND cfl.field IS NOT NULL"
	} else if mappedFilter == "unmapped" {
		querySQL += " AND cfl.field IS NULL"
	}

	querySQL += " ORDER BY c.full_path, f.name LIMIT ? OFFSET ?"
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

	results := []knowledgeFlagRow{}
	for rows.Next() {
		var flag, cmdPath, resName, relation, field, matchKind, modifier string
		if err := rows.Scan(&flag, &cmdPath, &resName, &relation, &field, &matchKind, &modifier); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeFlagRow{
			Flag:      flag,
			Command:   cmdPath,
			Resource:  resName,
			Relation:  relation,
			Field:     field,
			MatchKind: matchKind,
			Modifier:  modifier,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No flags found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "FLAG\tCOMMAND\tRESOURCE\tRELATION\tFIELD\tMATCH\tMODIFIER")
	for _, row := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", row.Flag, row.Command, row.Resource, row.Relation, row.Field, row.MatchKind, row.Modifier)
	}
	return w.Flush()
}
