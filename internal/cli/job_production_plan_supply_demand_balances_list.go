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

type jobProductionPlanSupplyDemandBalancesListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	JobProductionPlan           string
	UseObservedSupplyParameters string
}

type jobProductionPlanSupplyDemandBalanceRow struct {
	ID                           string   `json:"id"`
	JobProductionPlanID          string   `json:"job_production_plan_id,omitempty"`
	UseObservedSupplyParameters  bool     `json:"use_observed_supply_parameters"`
	PlannedNonProductionTruckCnt int      `json:"planned_non_production_truck_count"`
	PlannedPracticalSurplusPct   *float64 `json:"planned_practical_surplus_pct,omitempty"`
	ActualPracticalSurplusPct    *float64 `json:"actual_practical_surplus_pct,omitempty"`
	CreatedAt                    string   `json:"created_at,omitempty"`
}

func newJobProductionPlanSupplyDemandBalancesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan supply/demand balances",
		Long: `List job production plan supply/demand balances with filtering and pagination.

Output Columns:
  ID        Balance identifier
  JOB_PLAN  Job production plan ID
  OBSERVED  Use observed supply parameters (yes/no)
  NON_PROD  Planned non-production truck count
  PLANNED%  Planned practical surplus percent
  ACTUAL%   Actual practical surplus percent
  CREATED   Created timestamp

Filters:
  --job-production-plan             Filter by job production plan ID
  --use-observed-supply-parameters  Filter by observed supply parameters (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List supply/demand balances
  xbe view job-production-plan-supply-demand-balances list

  # Filter by job production plan
  xbe view job-production-plan-supply-demand-balances list --job-production-plan 123

  # Filter by observed supply parameters
  xbe view job-production-plan-supply-demand-balances list --use-observed-supply-parameters true

  # Output as JSON
  xbe view job-production-plan-supply-demand-balances list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanSupplyDemandBalancesList,
	}
	initJobProductionPlanSupplyDemandBalancesListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSupplyDemandBalancesCmd.AddCommand(newJobProductionPlanSupplyDemandBalancesListCmd())
}

func initJobProductionPlanSupplyDemandBalancesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("use-observed-supply-parameters", "", "Filter by observed supply parameters (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSupplyDemandBalancesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanSupplyDemandBalancesListOptions(cmd)
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
	query.Set("fields[job-production-plan-supply-demand-balances]", "use-observed-supply-parameters,planned-non-production-truck-count,planned-practical-surplus-pct,actual-practical-surplus-pct,created-at,job-production-plan")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[use-observed-supply-parameters]", opts.UseObservedSupplyParameters)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-supply-demand-balances", query)
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

	rows := buildJobProductionPlanSupplyDemandBalanceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanSupplyDemandBalancesTable(cmd, rows)
}

func parseJobProductionPlanSupplyDemandBalancesListOptions(cmd *cobra.Command) (jobProductionPlanSupplyDemandBalancesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	useObservedSupplyParameters, _ := cmd.Flags().GetString("use-observed-supply-parameters")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSupplyDemandBalancesListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		JobProductionPlan:           jobProductionPlan,
		UseObservedSupplyParameters: useObservedSupplyParameters,
	}, nil
}

func buildJobProductionPlanSupplyDemandBalanceRows(resp jsonAPIResponse) []jobProductionPlanSupplyDemandBalanceRow {
	rows := make([]jobProductionPlanSupplyDemandBalanceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := jobProductionPlanSupplyDemandBalanceRow{
			ID:                           resource.ID,
			UseObservedSupplyParameters:  boolAttr(attrs, "use-observed-supply-parameters"),
			PlannedNonProductionTruckCnt: intAttr(attrs, "planned-non-production-truck-count"),
			PlannedPracticalSurplusPct:   floatAttrPointer(attrs, "planned-practical-surplus-pct"),
			ActualPracticalSurplusPct:    floatAttrPointer(attrs, "actual-practical-surplus-pct"),
			CreatedAt:                    formatDateTime(stringAttr(attrs, "created-at")),
		}

		row.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanSupplyDemandBalancesTable(cmd *cobra.Command, rows []jobProductionPlanSupplyDemandBalanceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan supply/demand balances found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB_PLAN\tOBSERVED\tNON_PROD\tPLANNED%\tACTUAL%\tCREATED")
	for _, row := range rows {
		observed := "no"
		if row.UseObservedSupplyParameters {
			observed = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			observed,
			row.PlannedNonProductionTruckCnt,
			formatPct(row.PlannedPracticalSurplusPct),
			formatPct(row.ActualPracticalSurplusPct),
			row.CreatedAt,
		)
	}
	return writer.Flush()
}

func formatPct(value *float64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%.1f%%", *value*100)
}

func floatAttrPointer(attrs map[string]any, key string) *float64 {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case float64:
		v := typed
		return &v
	case float32:
		v := float64(typed)
		return &v
	case int:
		v := float64(typed)
		return &v
	case int64:
		v := float64(typed)
		return &v
	case string:
		if f, err := strconv.ParseFloat(typed, 64); err == nil {
			return &f
		}
	}
	return nil
}
