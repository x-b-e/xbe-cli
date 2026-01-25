package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectTransportPlanStrategySetsListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	StrategyPattern string
}

type projectTransportPlanStrategySetRow struct {
	ID              string `json:"id"`
	Name            string `json:"name,omitempty"`
	StrategyPattern string `json:"strategy_pattern,omitempty"`
	IsActive        bool   `json:"is_active,omitempty"`
}

func newProjectTransportPlanStrategySetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan strategy sets",
		Long: `List project transport plan strategy sets with filtering and pagination.

Strategy sets group transport plan strategies into reusable patterns for
project transport plans.

Output Columns:
  ID       Strategy set identifier
  NAME     Strategy set name
  PATTERN  Strategy pattern derived from strategies
  ACTIVE   Whether the set is active

Filters:
  --strategy-pattern  Filter by strategy pattern

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan strategy sets
  xbe view project-transport-plan-strategy-sets list

  # Filter by strategy pattern
  xbe view project-transport-plan-strategy-sets list --strategy-pattern \"[pick]-[drop]\"

  # Output as JSON
  xbe view project-transport-plan-strategy-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanStrategySetsList,
	}
	initProjectTransportPlanStrategySetsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStrategySetsCmd.AddCommand(newProjectTransportPlanStrategySetsListCmd())
}

func initProjectTransportPlanStrategySetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("strategy-pattern", "", "Filter by strategy pattern")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStrategySetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanStrategySetsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-strategy-sets]", "name,strategy-pattern,is-active")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "name")
	}
	setFilterIfPresent(query, "filter[strategy-pattern]", opts.StrategyPattern)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-strategy-sets", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildProjectTransportPlanStrategySetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanStrategySetsTable(cmd, rows)
}

func parseProjectTransportPlanStrategySetsListOptions(cmd *cobra.Command) (projectTransportPlanStrategySetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	strategyPattern, _ := cmd.Flags().GetString("strategy-pattern")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStrategySetsListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		StrategyPattern: strategyPattern,
	}, nil
}

func buildProjectTransportPlanStrategySetRows(resp jsonAPIResponse) []projectTransportPlanStrategySetRow {
	rows := make([]projectTransportPlanStrategySetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, projectTransportPlanStrategySetRow{
			ID:              resource.ID,
			Name:            stringAttr(resource.Attributes, "name"),
			StrategyPattern: stringAttr(resource.Attributes, "strategy-pattern"),
			IsActive:        boolAttr(resource.Attributes, "is-active"),
		})
	}
	return rows
}

func renderProjectTransportPlanStrategySetsTable(cmd *cobra.Command, rows []projectTransportPlanStrategySetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan strategy sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tPATTERN\tACTIVE")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 24),
			truncateString(row.StrategyPattern, 36),
			active,
		)
	}
	return writer.Flush()
}
