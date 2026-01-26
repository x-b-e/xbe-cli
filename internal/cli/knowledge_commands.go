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
	Permissions string `json:"permissions,omitempty"`
	SideEffects string `json:"side_effects,omitempty"`
	Validation  string `json:"validation_notes,omitempty"`
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
SELECT c.full_path, c.description, COALESCE(c.permissions, ''), COALESCE(c.side_effects, ''), COALESCE(c.validation_notes, ''),
       COALESCE(crl.command_kind, ''), COALESCE(crl.verb, ''), COALESCE(crl.resource, '')
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
		var path, desc, perms, sideEffects, validation, cmdKind, cmdVerb, cmdResource string
		if err := rows.Scan(&path, &desc, &perms, &sideEffects, &validation, &cmdKind, &cmdVerb, &cmdResource); err != nil {
			return checkDBError(err, dbPath)
		}
		results = append(results, knowledgeCommandRow{
			Path:        path,
			Description: desc,
			Permissions: perms,
			SideEffects: sideEffects,
			Validation:  validation,
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

	for _, row := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", row.Path)
		if row.Kind != "" || row.Verb != "" || row.Resource != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  kind: %s\n", row.Kind)
			fmt.Fprintf(cmd.OutOrStdout(), "  verb: %s\n", row.Verb)
			fmt.Fprintf(cmd.OutOrStdout(), "  resource: %s\n", row.Resource)
		}
		if row.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  description: %s\n", row.Description)
		}
		if row.Permissions != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  permissions: %s\n", row.Permissions)
		}
		if row.SideEffects != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  side_effects: %s\n", row.SideEffects)
		}
		if row.Validation != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  validation_notes: %s\n", row.Validation)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}
