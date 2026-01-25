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

type maintenanceRequirementRuleMaintenanceRequirementSetsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Sort                         string
	MaintenanceRequirementRuleID string
	MaintenanceRequirementSetID  string
}

type maintenanceRequirementRuleMaintenanceRequirementSetRow struct {
	ID                           string `json:"id"`
	MaintenanceRequirementRuleID string `json:"maintenance_requirement_rule_id,omitempty"`
	MaintenanceRequirementSetID  string `json:"maintenance_requirement_set_id,omitempty"`
}

func newMaintenanceRequirementRuleMaintenanceRequirementSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maintenance requirement rule maintenance requirement sets",
		Long: `List maintenance requirement rule maintenance requirement sets.

Output Columns:
  ID       Maintenance requirement rule maintenance requirement set identifier
  RULE     Maintenance requirement rule ID
  SET      Maintenance requirement set ID

Filters:
  --maintenance-requirement-rule  Filter by maintenance requirement rule ID
  --maintenance-requirement-set   Filter by maintenance requirement set ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List maintenance requirement rule maintenance requirement sets
  xbe view maintenance-requirement-rule-maintenance-requirement-sets list

  # Filter by maintenance requirement rule
  xbe view maintenance-requirement-rule-maintenance-requirement-sets list --maintenance-requirement-rule 123

  # Filter by maintenance requirement set
  xbe view maintenance-requirement-rule-maintenance-requirement-sets list --maintenance-requirement-set 456

  # Output as JSON
  xbe view maintenance-requirement-rule-maintenance-requirement-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runMaintenanceRequirementRuleMaintenanceRequirementSetsList,
	}
	initMaintenanceRequirementRuleMaintenanceRequirementSetsListFlags(cmd)
	return cmd
}

func init() {
	maintenanceRequirementRuleMaintenanceRequirementSetsCmd.AddCommand(newMaintenanceRequirementRuleMaintenanceRequirementSetsListCmd())
}

func initMaintenanceRequirementRuleMaintenanceRequirementSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("maintenance-requirement-rule", "", "Filter by maintenance requirement rule ID")
	cmd.Flags().String("maintenance-requirement-set", "", "Filter by maintenance requirement set ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaintenanceRequirementRuleMaintenanceRequirementSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaintenanceRequirementRuleMaintenanceRequirementSetsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[maintenance-requirement-rule]", opts.MaintenanceRequirementRuleID)
	setFilterIfPresent(query, "filter[maintenance-requirement-set]", opts.MaintenanceRequirementSetID)

	body, _, err := client.Get(cmd.Context(), "/v1/maintenance-requirement-rule-maintenance-requirement-sets", query)
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

	rows := buildMaintenanceRequirementRuleMaintenanceRequirementSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaintenanceRequirementRuleMaintenanceRequirementSetsTable(cmd, rows)
}

func parseMaintenanceRequirementRuleMaintenanceRequirementSetsListOptions(cmd *cobra.Command) (maintenanceRequirementRuleMaintenanceRequirementSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	maintenanceRequirementRuleID, _ := cmd.Flags().GetString("maintenance-requirement-rule")
	maintenanceRequirementSetID, _ := cmd.Flags().GetString("maintenance-requirement-set")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return maintenanceRequirementRuleMaintenanceRequirementSetsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Sort:                         sort,
		MaintenanceRequirementRuleID: maintenanceRequirementRuleID,
		MaintenanceRequirementSetID:  maintenanceRequirementSetID,
	}, nil
}

func buildMaintenanceRequirementRuleMaintenanceRequirementSetRows(resp jsonAPIResponse) []maintenanceRequirementRuleMaintenanceRequirementSetRow {
	rows := make([]maintenanceRequirementRuleMaintenanceRequirementSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := maintenanceRequirementRuleMaintenanceRequirementSetRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["maintenance-requirement-rule"]; ok && rel.Data != nil {
			row.MaintenanceRequirementRuleID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["maintenance-requirement-set"]; ok && rel.Data != nil {
			row.MaintenanceRequirementSetID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderMaintenanceRequirementRuleMaintenanceRequirementSetsTable(cmd *cobra.Command, rows []maintenanceRequirementRuleMaintenanceRequirementSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No maintenance requirement rule maintenance requirement sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRULE\tSET")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.MaintenanceRequirementRuleID,
			row.MaintenanceRequirementSetID,
		)
	}
	return writer.Flush()
}
