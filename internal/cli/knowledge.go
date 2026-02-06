package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
prior context.

Alias: xbe kb ...`,
	Example: `  # Read the first-run guide
  xbe knowledge guide

  # Search across resources, commands, fields, and summaries
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

	knowledgeCmd.AddCommand(newKnowledgeGuideCmd())
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
	knowledgeCmd.AddCommand(newKnowledgeClientRoutesCmd())

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

func allowedValues(values ...string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		out[normalized] = true
	}
	return out
}

func allowedValuesList(allowed map[string]bool) string {
	if len(allowed) == 0 {
		return ""
	}
	items := make([]string, 0, len(allowed))
	for value := range allowed {
		items = append(items, value)
	}
	sort.Strings(items)
	return strings.Join(items, ", ")
}

func validateEnum(flagName, value string, allowed map[string]bool) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return "", nil
	}
	if allowed[normalized] {
		return normalized, nil
	}
	return "", fmt.Errorf("invalid %s value %q (valid: %s)", flagName, value, allowedValuesList(allowed))
}

func validateCSVEnum(flagName, raw string, allowed map[string]bool) ([]string, error) {
	values := parseCSVFilter(raw)
	if len(values) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized, err := validateEnum(flagName, value, allowed)
		if err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
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

func knowledgeResourceExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	row := db.QueryRowContext(ctx, "SELECT 1 FROM resources WHERE name = ? LIMIT 1", name)
	var one int
	if err := row.Scan(&one); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return one == 1, nil
}

func summaryResourceExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	row := db.QueryRowContext(ctx, "SELECT 1 FROM summary_resource_targets WHERE summary_resource = ? LIMIT 1", name)
	var one int
	if err := row.Scan(&one); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return one == 1, nil
}

func summaryResourceForCommandName(ctx context.Context, db *sql.DB, commandName string) ([]string, error) {
	commandName = strings.ToLower(strings.TrimSpace(commandName))
	if commandName == "" {
		return nil, nil
	}

	rows, err := db.QueryContext(
		ctx,
		`SELECT DISTINCT crl.resource
FROM commands c
JOIN command_resource_links crl ON crl.command_id = c.id
WHERE crl.command_kind = 'summarize'
  AND (
    c.full_path = ?
    OR c.full_path LIKE ?
  )
ORDER BY crl.resource`,
		commandName,
		"summarize "+commandName+" %",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var resource string
		if err := rows.Scan(&resource); err != nil {
			return nil, err
		}
		out = append(out, resource)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func resolveKnowledgeResourceName(ctx context.Context, db *sql.DB, raw string) (string, bool, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if name == "" {
		return "", false, nil
	}

	exists, err := knowledgeResourceExists(ctx, db, name)
	if err != nil {
		return "", false, err
	}
	if exists {
		return name, false, nil
	}

	if summaryResources, err := summaryResourceForCommandName(ctx, db, name); err != nil {
		return "", false, err
	} else if len(summaryResources) == 1 {
		return summaryResources[0], true, nil
	} else if len(summaryResources) > 1 {
		return "", false, fmt.Errorf("resource %q is ambiguous (matches: %s)", raw, strings.Join(summaryResources, ", "))
	}

	candidates := []string{name + "s"}
	if strings.HasSuffix(name, "y") && len(name) > 1 {
		candidates = append(candidates, name[:len(name)-1]+"ies")
	}
	if strings.HasSuffix(name, "-summary") {
		candidates = append(candidates, strings.TrimSuffix(name, "-summary")+"-summaries")
	}
	if strings.HasSuffix(name, "s") && len(name) > 1 {
		candidates = append(candidates, singularizeWord(name))
	}

	seen := map[string]bool{}
	unique := []string{}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		unique = append(unique, candidate)
	}

	matches := []string{}
	for _, candidate := range unique {
		ok, err := knowledgeResourceExists(ctx, db, candidate)
		if err != nil {
			return "", false, err
		}
		if ok {
			matches = append(matches, candidate)
		}
	}
	if len(matches) == 1 {
		return matches[0], true, nil
	}
	if len(matches) > 1 {
		return "", false, fmt.Errorf("resource %q is ambiguous (matches: %s)", raw, strings.Join(matches, ", "))
	}

	return "", false, nil
}

func suggestKnowledgeResources(ctx context.Context, db *sql.DB, raw string, limit int) ([]string, error) {
	needle := strings.ToLower(strings.TrimSpace(raw))
	if needle == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}

	querySuggestion := func(pattern, prefix string, n int) ([]string, error) {
		rows, err := db.QueryContext(
			ctx,
			`SELECT name
FROM resources
WHERE name LIKE ?
ORDER BY
  CASE
    WHEN name = ? THEN 0
    WHEN name LIKE ? THEN 1
    ELSE 2
  END,
  LENGTH(name),
  name
LIMIT ?`,
			pattern,
			needle,
			prefix,
			n,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []string{}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			out = append(out, name)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}

	seen := map[string]bool{}
	add := func(items []string, out *[]string) {
		for _, item := range items {
			if seen[item] {
				continue
			}
			seen[item] = true
			*out = append(*out, item)
			if len(*out) >= limit {
				return
			}
		}
	}

	results := []string{}
	primary, err := querySuggestion("%"+needle+"%", needle+"%", limit)
	if err != nil {
		return nil, err
	}
	add(primary, &results)
	if len(results) >= limit {
		return results, nil
	}

	for _, token := range strings.Split(needle, "-") {
		token = strings.TrimSpace(token)
		if len(token) < 3 {
			continue
		}
		more, err := querySuggestion("%"+token+"%", token+"%", limit-len(results))
		if err != nil {
			return nil, err
		}
		add(more, &results)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

func suggestSummaryResources(ctx context.Context, db *sql.DB, raw string, limit int) ([]string, error) {
	needle := strings.ToLower(strings.TrimSpace(raw))
	if needle == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}
	rows, err := db.QueryContext(
		ctx,
		`SELECT DISTINCT summary_resource
FROM summary_resource_targets
WHERE summary_resource LIKE ?
ORDER BY
  CASE
    WHEN summary_resource = ? THEN 0
    WHEN summary_resource LIKE ? THEN 1
    ELSE 2
  END,
  LENGTH(summary_resource),
  summary_resource
LIMIT ?`,
		"%"+needle+"%",
		needle,
		needle+"%",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var summary string
		if err := rows.Scan(&summary); err != nil {
			return nil, err
		}
		out = append(out, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func resourceHintSuffix(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}
	return fmt.Sprintf("; closest matches: %s", strings.Join(suggestions, ", "))
}

func normalizeKnowledgeResourceArg(cmd *cobra.Command, db *sql.DB, dbPath string, raw string, label string) (string, error) {
	ctx := context.Background()
	resolved, changed, err := resolveKnowledgeResourceName(ctx, db, raw)
	if err != nil {
		return "", checkDBError(err, dbPath)
	}
	if resolved == "" {
		suggestions, suggestErr := suggestKnowledgeResources(ctx, db, raw, 6)
		if suggestErr != nil {
			return "", checkDBError(suggestErr, dbPath)
		}
		return "", fmt.Errorf("%s %q not found (try 'xbe knowledge resources --query %q'%s)", label, raw, raw, resourceHintSuffix(suggestions))
	}
	if changed {
		fmt.Fprintf(cmd.ErrOrStderr(), "Interpreting %s %q as %q.\n", label, raw, resolved)
	}
	return resolved, nil
}

func normalizeKnowledgeResourceFlag(cmd *cobra.Command, db *sql.DB, dbPath string, raw string, flagName string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	return normalizeKnowledgeResourceArg(cmd, db, dbPath, raw, flagName)
}

func normalizeSummaryResourceFilter(cmd *cobra.Command, db *sql.DB, dbPath string, raw string) (string, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return "", nil
	}

	ctx := context.Background()
	exists, err := summaryResourceExists(ctx, db, raw)
	if err != nil {
		return "", checkDBError(err, dbPath)
	}
	if exists {
		return raw, nil
	}

	matches, err := summaryResourceForCommandName(ctx, db, raw)
	if err != nil {
		return "", checkDBError(err, dbPath)
	}
	if len(matches) == 1 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Interpreting --summary %q as %q.\n", raw, matches[0])
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("--summary %q is ambiguous (matches: %s)", raw, strings.Join(matches, ", "))
	}

	if strings.HasSuffix(raw, "-summary") {
		candidate := strings.TrimSuffix(raw, "-summary") + "-summaries"
		ok, err := summaryResourceExists(ctx, db, candidate)
		if err != nil {
			return "", checkDBError(err, dbPath)
		}
		if ok {
			fmt.Fprintf(cmd.ErrOrStderr(), "Interpreting --summary %q as %q.\n", raw, candidate)
			return candidate, nil
		}
	}

	suggestions, err := suggestSummaryResources(ctx, db, raw, 6)
	if err != nil {
		return "", checkDBError(err, dbPath)
	}
	return "", fmt.Errorf("summary resource %q not found (try 'xbe knowledge summaries --summary <name>'%s)", raw, resourceHintSuffix(suggestions))
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
