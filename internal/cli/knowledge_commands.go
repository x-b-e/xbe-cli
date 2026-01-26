package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type knowledgeCommandRow struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Verb        string `json:"verb,omitempty"`
	Resource    string `json:"resource,omitempty"`
}

func newKnowledgeCommandsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commands",
		Short: "List or search CLI commands in the knowledge base",
		RunE:  runKnowledgeCommands,
		Example: `  # Search commands by keyword
  xbe knowledge commands --query project

  # Commands tied to a resource
  xbe knowledge commands --resource jobs`,
	}
	cmd.Flags().String("query", "", "Substring filter for command path or description")
	cmd.Flags().String("resource", "", "Only commands that operate on a resource")
	cmd.Flags().String("kind", "", "Filter by command kind (view, do, summarize)")
	cmd.Flags().String("verb", "", "Filter by verb (list, show, create, update, delete)")
	return cmd
}

func runKnowledgeCommands(cmd *cobra.Command, _ []string) error {
	query := strings.TrimSpace(getStringFlag(cmd, "query"))
	resource := strings.TrimSpace(getStringFlag(cmd, "resource"))
	kind := strings.TrimSpace(getStringFlag(cmd, "kind"))
	verb := strings.TrimSpace(getStringFlag(cmd, "verb"))

	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	args := []any{}
	querySQL := `
SELECT c.full_path, c.description, COALESCE(crl.command_kind, ''), COALESCE(crl.verb, ''), COALESCE(crl.resource, '')
FROM commands c
LEFT JOIN command_resource_links crl ON crl.command_id = c.id
WHERE 1=1`

	if query != "" {
		querySQL += " AND (c.full_path LIKE ? OR c.description LIKE ?)"
		pattern := "%" + query + "%"
		args = append(args, pattern, pattern)
	}
	if resource != "" {
		querySQL += " AND crl.resource = ?"
		args = append(args, resource)
	}
	if kind != "" {
		querySQL += " AND crl.command_kind = ?"
		args = append(args, kind)
	}
	if verb != "" {
		querySQL += " AND crl.verb = ?"
		args = append(args, verb)
	}

	querySQL += " ORDER BY c.full_path LIMIT ? OFFSET ?"
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

	results := []knowledgeCommandRow{}
	for rows.Next() {
		var path, desc, cmdKind, cmdVerb, cmdResource string
		if err := rows.Scan(&path, &desc, &cmdKind, &cmdVerb, &cmdResource); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeCommandRow{
			Path:        path,
			Description: desc,
			Kind:        cmdKind,
			Verb:        cmdVerb,
			Resource:    cmdResource,
		})
	}
	if err := rows.Err(); err != nil {
		return checkDBError(err, dbPath)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No commands found.")
		return nil
	}

	if getBoolFlag(cmd, "json") {
		return renderKnowledgeJSON(cmd, results)
	}

	w := newTabWriter(cmd)
	fmt.Fprintln(w, "COMMAND\tKIND\tVERB\tRESOURCE\tDESCRIPTION")
	for _, row := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", row.Path, row.Kind, row.Verb, row.Resource, row.Description)
	}
	return w.Flush()
}
