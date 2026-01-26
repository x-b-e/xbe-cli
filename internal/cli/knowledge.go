package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

const (
	knowledgeDBEnv     = "XBE_KNOWLEDGE_DB"
	defaultKnowledgeDB = "cartographer_out/db/knowledge.sqlite"
)

var knowledgeCmd = &cobra.Command{
	Use:     "knowledge",
	Aliases: []string{"kb"},
	Short:   "Explore the local knowledge database",
	Long: `Explore the local knowledge database produced by the Cartographer pipeline.

This toolkit is designed for AI agents and power users who need to quickly map
resources, commands, fields, summaries, and neighborhood relationships without
prior context.`,
	Example: `  # Search across resources, commands, fields, and summaries
  xbe knowledge search job

  # Show a resource with relationships, summaries, and commands
  xbe knowledge resource jobs

  # Rank neighbors for exploration
  xbe knowledge neighbors jobs --limit 20

  # Show multi-hop filter paths inferred from commands
  xbe knowledge filters --resource jobs`,
	Annotations: map[string]string{"group": GroupKnowledge},
}

func init() {
	knowledgeCmd.PersistentFlags().String("db", "", "Internal override for knowledge database path")
	_ = knowledgeCmd.PersistentFlags().MarkHidden("db")
	knowledgeCmd.PersistentFlags().Bool("json", false, "Output JSON")
	knowledgeCmd.PersistentFlags().Int("limit", 50, "Maximum results to return")
	knowledgeCmd.PersistentFlags().Int("offset", 0, "Number of results to skip")

	knowledgeCmd.AddCommand(newKnowledgeSearchCmd())
	knowledgeCmd.AddCommand(newKnowledgeResourcesCmd())
	knowledgeCmd.AddCommand(newKnowledgeResourceCmd())
	knowledgeCmd.AddCommand(newKnowledgeCommandsCmd())
	knowledgeCmd.AddCommand(newKnowledgeFieldsCmd())
	knowledgeCmd.AddCommand(newKnowledgeFlagsCmd())
	knowledgeCmd.AddCommand(newKnowledgeRelationsCmd())
	knowledgeCmd.AddCommand(newKnowledgeSummariesCmd())
	knowledgeCmd.AddCommand(newKnowledgeNeighborsCmd())
	knowledgeCmd.AddCommand(newKnowledgeMetapathCmd())
	knowledgeCmd.AddCommand(newKnowledgeFiltersCmd())

	rootCmd.AddCommand(knowledgeCmd)
}

func resolveKnowledgeDBPath(cmd *cobra.Command) (string, error) {
	if value := strings.TrimSpace(os.Getenv(knowledgeDBEnv)); value != "" {
		return value, nil
	}
	if value, err := cmd.Flags().GetString("db"); err == nil && strings.TrimSpace(value) != "" {
		return value, nil
	}
	if value, err := cmd.InheritedFlags().GetString("db"); err == nil && strings.TrimSpace(value) != "" {
		return value, nil
	}
	if path, err := ensureEmbeddedKnowledgeDB(); err == nil {
		return path, nil
	} else if err != nil {
		if _, statErr := os.Stat(defaultKnowledgeDB); statErr == nil {
			return defaultKnowledgeDB, nil
		}
		return "", err
	}
	return defaultKnowledgeDB, nil
}

func openKnowledgeDB(cmd *cobra.Command) (*sql.DB, string, error) {
	path, err := resolveKnowledgeDBPath(cmd)
	if err != nil {
		return nil, "", err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, path, fmt.Errorf("resolve knowledge database path")
	}
	if _, err := os.Stat(absPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, absPath, fmt.Errorf("knowledge database not found; reinstall the CLI or run build_tools/compile.py when building from source")
		}
		if os.IsPermission(err) {
			return nil, absPath, fmt.Errorf("knowledge database is not accessible (permission denied)")
		}
		return nil, absPath, fmt.Errorf("unable to access knowledge database")
	}
	// Use a read-only connection when possible.
	db, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, absPath, fmt.Errorf("open knowledge db: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 60000"); err != nil {
		_ = db.Close()
		return nil, absPath, fmt.Errorf("configure knowledge db: %w", err)
	}
	if _, err := db.Exec("PRAGMA query_only = ON"); err != nil {
		_ = db.Close()
		return nil, absPath, fmt.Errorf("configure knowledge db read-only: %w", err)
	}
	return db, absPath, nil
}

func getBoolFlag(cmd *cobra.Command, name string) bool {
	if value, err := cmd.Flags().GetBool(name); err == nil {
		return value
	}
	if value, err := cmd.InheritedFlags().GetBool(name); err == nil {
		return value
	}
	return false
}

func getIntFlag(cmd *cobra.Command, name string) int {
	if value, err := cmd.Flags().GetInt(name); err == nil {
		return value
	}
	if value, err := cmd.InheritedFlags().GetInt(name); err == nil {
		return value
	}
	return 0
}

func getStringFlag(cmd *cobra.Command, name string) string {
	if value, err := cmd.Flags().GetString(name); err == nil {
		return value
	}
	if value, err := cmd.InheritedFlags().GetString(name); err == nil {
		return value
	}
	return ""
}

func renderKnowledgeJSON(cmd *cobra.Command, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(payload))
	return nil
}

func newTabWriter(cmd *cobra.Command) *tabwriter.Writer {
	return tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
}

func parseCSVFilter(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func ensureSingleMatch(matches []string, subject string) (string, error) {
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no matches for %s", subject)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple matches for %s: %s", subject, strings.Join(matches, ", "))
	}
}

func requireArg(cmd *cobra.Command, args []string, index int, label string) (string, error) {
	if len(args) <= index {
		return "", fmt.Errorf("%s is required", label)
	}
	value := strings.TrimSpace(args[index])
	if value == "" {
		return "", fmt.Errorf("%s is required", label)
	}
	return value, nil
}

func queryContext(ctx context.Context, db *sql.DB, query string, args ...any) (*sql.Rows, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func parseJSONList(raw string) []string {
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []string{raw}
	}
	return items
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func collectStrings(rows *sql.Rows) ([]string, error) {
	var out []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}

func ensureNotEmpty(value string, label string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", label)
	}
	return nil
}

func checkDBError(err error, dbPath string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	_ = dbPath
	return fmt.Errorf("knowledge db: %w", err)
}
