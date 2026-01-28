package cli

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const metadataWrappedAnnotation = "metadata_wrapped"

var metadataSupportApplied bool

type commandMetadata struct {
	Permissions     string
	SideEffects     string
	ValidationNotes string
}

func applyCommandMetadataSupport(root *cobra.Command) {
	if metadataSupportApplied || root == nil {
		return
	}
	metadataSupportApplied = true

	attachMetadataFlags(viewCmd)
	attachMetadataFlags(doCmd)
	attachMetadataFlags(summarizeCmd)

	wrapCommandTree(viewCmd)
	wrapCommandTree(doCmd)
	wrapCommandTree(summarizeCmd)
}

func attachMetadataFlags(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	flags := cmd.PersistentFlags()
	if flags.Lookup("permissions") == nil {
		flags.Bool("permissions", false, "Show required permissions for this command")
	}
	if flags.Lookup("side-effects") == nil {
		flags.Bool("side-effects", false, "Show side effects for this command")
	}
	if flags.Lookup("validation-notes") == nil {
		flags.Bool("validation-notes", false, "Show validation notes for this command")
	}
	if flags.Lookup("metadata") == nil {
		flags.Bool("metadata", false, "Show permissions, side effects, and validation notes for this command")
	}
}

func wrapCommandTree(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	wrapCommand(cmd)
	for _, child := range cmd.Commands() {
		wrapCommandTree(child)
	}
}

func wrapCommand(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	if cmd.Annotations[metadataWrappedAnnotation] == "true" {
		return
	}
	cmd.Annotations[metadataWrappedAnnotation] = "true"

	if cmd.Args != nil {
		originalArgs := cmd.Args
		cmd.Args = func(cmd *cobra.Command, args []string) error {
			if wantsCommandMetadata(cmd) {
				return nil
			}
			return originalArgs(cmd, args)
		}
	}

	if cmd.RunE != nil {
		originalRunE := cmd.RunE
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			handled, err := handleCommandMetadataFlags(cmd)
			if handled || err != nil {
				return err
			}
			return originalRunE(cmd, args)
		}
		return
	}

	if cmd.Run != nil {
		originalRun := cmd.Run
		cmd.Run = func(cmd *cobra.Command, args []string) {
			handled, err := handleCommandMetadataFlags(cmd)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return
			}
			if handled {
				return
			}
			originalRun(cmd, args)
		}
		return
	}

	if cmd.Run == nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			handled, err := handleCommandMetadataFlags(cmd)
			if handled || err != nil {
				return err
			}
			return cmd.Help()
		}
	}
}

func wantsCommandMetadata(cmd *cobra.Command) bool {
	return getBoolFlag(cmd, "permissions") || getBoolFlag(cmd, "side-effects") || getBoolFlag(cmd, "validation-notes") || getBoolFlag(cmd, "metadata")
}

func handleCommandMetadataFlags(cmd *cobra.Command) (bool, error) {
	if !wantsCommandMetadata(cmd) {
		return false, nil
	}
	meta, ok, err := lookupCommandMetadata(cmd)
	if err != nil {
		return true, err
	}
	out := cmd.OutOrStdout()
	if !ok {
		fmt.Fprintln(out, "No metadata available.")
		return true, nil
	}

	showAll := getBoolFlag(cmd, "metadata")
	showPerms := showAll || getBoolFlag(cmd, "permissions")
	showSideEffects := showAll || getBoolFlag(cmd, "side-effects")
	showValidation := showAll || getBoolFlag(cmd, "validation-notes")
	printed := false

	if showPerms {
		if strings.TrimSpace(meta.Permissions) != "" {
			fmt.Fprintf(out, "permissions: %s\n", meta.Permissions)
			printed = true
		}
	}
	if showSideEffects {
		if strings.TrimSpace(meta.SideEffects) != "" {
			fmt.Fprintf(out, "side_effects: %s\n", meta.SideEffects)
			printed = true
		}
	}
	if showValidation {
		if strings.TrimSpace(meta.ValidationNotes) != "" {
			fmt.Fprintf(out, "validation_notes: %s\n", meta.ValidationNotes)
			printed = true
		}
	}

	if !printed {
		fmt.Fprintln(out, "No metadata available.")
	}
	return true, nil
}

func lookupCommandMetadata(cmd *cobra.Command) (commandMetadata, bool, error) {
	db, dbPath, err := openKnowledgeDB(cmd)
	if err != nil {
		return commandMetadata{}, false, err
	}
	defer db.Close()

	commandPath := knowledgeCommandPath(cmd)
	if commandPath == "" {
		return commandMetadata{}, false, nil
	}

	ctx := context.Background()
	meta, ok, err := fetchCommandMetadata(ctx, db, dbPath, commandPath)
	if err != nil || ok {
		return meta, ok, err
	}

	if fallback, ok := listFallbackCommandPath(commandPath); ok {
		return fetchCommandMetadata(ctx, db, dbPath, fallback)
	}

	return commandMetadata{}, false, nil
}

func fetchCommandMetadata(ctx context.Context, db *sql.DB, dbPath string, commandPath string) (commandMetadata, bool, error) {
	row := db.QueryRowContext(ctx, `
SELECT COALESCE(permissions, ''), COALESCE(side_effects, ''), COALESCE(validation_notes, '')
FROM commands
WHERE full_path = ?
LIMIT 1`, commandPath)

	var permissions, sideEffects, validationNotes string
	if err := row.Scan(&permissions, &sideEffects, &validationNotes); err != nil {
		if err == sql.ErrNoRows {
			return commandMetadata{}, false, nil
		}
		return commandMetadata{}, false, checkDBError(err, dbPath)
	}

	return commandMetadata{
		Permissions:     permissions,
		SideEffects:     sideEffects,
		ValidationNotes: validationNotes,
	}, true, nil
}

func knowledgeCommandPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	path := cmd.CommandPath()
	parts := strings.Fields(path)
	if len(parts) == 0 {
		return ""
	}
	if parts[0] == rootCmd.Name() {
		parts = parts[1:]
	}
	return strings.Join(parts, " ")
}

func listFallbackCommandPath(commandPath string) (string, bool) {
	parts := strings.Fields(commandPath)
	if len(parts) != 3 {
		return "", false
	}
	if parts[0] != "view" || parts[2] != "show" {
		return "", false
	}
	return strings.Join([]string{parts[0], parts[1], "list"}, " "), true
}
