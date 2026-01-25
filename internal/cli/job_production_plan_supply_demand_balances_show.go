package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanSupplyDemandBalancesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanSupplyDemandBalanceDetails struct {
	ID                                  string   `json:"id"`
	JobProductionPlanID                 string   `json:"job_production_plan_id,omitempty"`
	UseObservedSupplyParameters         bool     `json:"use_observed_supply_parameters"`
	TonsPerCycleExplicit                *float64 `json:"tons_per_cycle_explicit,omitempty"`
	MaterialSiteMinutesPerCycleExplicit *float64 `json:"material_site_minutes_per_cycle_explicit,omitempty"`
	DrivingMinutesPerCycleExplicit      *float64 `json:"driving_minutes_per_cycle_explicit,omitempty"`
	PlannedMaterialTransactions         any      `json:"planned_material_transactions,omitempty"`
	ActualMaterialTransactions          any      `json:"actual_material_transactions,omitempty"`
	ActualTrucks                        any      `json:"actual_trucks,omitempty"`
	PlannedTrucks                       any      `json:"planned_trucks,omitempty"`
	PlannedNonProductionTruckCount      int      `json:"planned_non_production_truck_count"`
	PlannedBalances                     any      `json:"planned_balances,omitempty"`
	ActualBalances                      any      `json:"actual_balances,omitempty"`
	AverageActualPracticalSurplus       *float64 `json:"average_actual_practical_surplus,omitempty"`
	SumActualPracticalSurplus           *float64 `json:"sum_actual_practical_surplus,omitempty"`
	AveragePlannedPracticalSurplus      *float64 `json:"average_planned_practical_surplus,omitempty"`
	SumPlannedPracticalSurplus          *float64 `json:"sum_planned_practical_surplus,omitempty"`
	ActualPracticalSurplusPct           *float64 `json:"actual_practical_surplus_pct,omitempty"`
	PlannedPracticalSurplusPct          *float64 `json:"planned_practical_surplus_pct,omitempty"`
	CreatedAt                           string   `json:"created_at,omitempty"`
	UpdatedAt                           string   `json:"updated_at,omitempty"`
}

func newJobProductionPlanSupplyDemandBalancesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan supply/demand balance details",
		Long: `Show the full details of a job production plan supply/demand balance.

Output Fields:
  ID
  Job Production Plan ID
  Use Observed Supply Parameters
  Tons Per Cycle Explicit
  Material Site Minutes Per Cycle Explicit
  Driving Minutes Per Cycle Explicit
  Planned Material Transactions
  Actual Material Transactions
  Planned Trucks
  Actual Trucks
  Planned Non-Production Truck Count
  Planned Balances
  Actual Balances
  Average Actual Practical Surplus
  Sum Actual Practical Surplus
  Average Planned Practical Surplus
  Sum Planned Practical Surplus
  Actual Practical Surplus Pct
  Planned Practical Surplus Pct
  Created At
  Updated At

Arguments:
  <id>    The supply/demand balance ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a supply/demand balance
  xbe view job-production-plan-supply-demand-balances show 123

  # Get JSON output
  xbe view job-production-plan-supply-demand-balances show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanSupplyDemandBalancesShow,
	}
	initJobProductionPlanSupplyDemandBalancesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSupplyDemandBalancesCmd.AddCommand(newJobProductionPlanSupplyDemandBalancesShowCmd())
}

func initJobProductionPlanSupplyDemandBalancesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSupplyDemandBalancesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanSupplyDemandBalancesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan supply/demand balance id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-supply-demand-balances]", "use-observed-supply-parameters,tons-per-cycle-explicit,material-site-minutes-per-cycle-explicit,driving-minutes-per-cycle-explicit,planned-material-transactions,actual-material-transactions,actual-trucks,planned-trucks,planned-non-production-truck-count,planned-balances,actual-balances,average-actual-practical-surplus,sum-actual-practical-surplus,average-planned-practical-surplus,sum-planned-practical-surplus,actual-practical-surplus-pct,planned-practical-surplus-pct,created-at,updated-at,job-production-plan")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-supply-demand-balances/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildJobProductionPlanSupplyDemandBalanceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanSupplyDemandBalanceDetails(cmd, details)
}

func parseJobProductionPlanSupplyDemandBalancesShowOptions(cmd *cobra.Command) (jobProductionPlanSupplyDemandBalancesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSupplyDemandBalancesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanSupplyDemandBalanceDetails(resp jsonAPISingleResponse) jobProductionPlanSupplyDemandBalanceDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanSupplyDemandBalanceDetails{
		ID:                                  resource.ID,
		UseObservedSupplyParameters:         boolAttr(attrs, "use-observed-supply-parameters"),
		PlannedNonProductionTruckCount:      intAttr(attrs, "planned-non-production-truck-count"),
		CreatedAt:                           formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                           formatDateTime(stringAttr(attrs, "updated-at")),
		TonsPerCycleExplicit:                floatAttrPointer(attrs, "tons-per-cycle-explicit"),
		MaterialSiteMinutesPerCycleExplicit: floatAttrPointer(attrs, "material-site-minutes-per-cycle-explicit"),
		DrivingMinutesPerCycleExplicit:      floatAttrPointer(attrs, "driving-minutes-per-cycle-explicit"),
		AverageActualPracticalSurplus:       floatAttrPointer(attrs, "average-actual-practical-surplus"),
		SumActualPracticalSurplus:           floatAttrPointer(attrs, "sum-actual-practical-surplus"),
		AveragePlannedPracticalSurplus:      floatAttrPointer(attrs, "average-planned-practical-surplus"),
		SumPlannedPracticalSurplus:          floatAttrPointer(attrs, "sum-planned-practical-surplus"),
		ActualPracticalSurplusPct:           floatAttrPointer(attrs, "actual-practical-surplus-pct"),
		PlannedPracticalSurplusPct:          floatAttrPointer(attrs, "planned-practical-surplus-pct"),
	}

	details.PlannedMaterialTransactions = attrs["planned-material-transactions"]
	details.ActualMaterialTransactions = attrs["actual-material-transactions"]
	details.ActualTrucks = attrs["actual-trucks"]
	details.PlannedTrucks = attrs["planned-trucks"]
	details.PlannedBalances = attrs["planned-balances"]
	details.ActualBalances = attrs["actual-balances"]

	details.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")

	return details
}

func renderJobProductionPlanSupplyDemandBalanceDetails(cmd *cobra.Command, details jobProductionPlanSupplyDemandBalanceDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	fmt.Fprintf(out, "Use Observed Supply Parameters: %s\n", formatBool(details.UseObservedSupplyParameters))
	if details.TonsPerCycleExplicit != nil {
		fmt.Fprintf(out, "Tons Per Cycle Explicit: %s\n", formatFloat(details.TonsPerCycleExplicit, 2))
	}
	if details.MaterialSiteMinutesPerCycleExplicit != nil {
		fmt.Fprintf(out, "Material Site Minutes Per Cycle Explicit: %s\n", formatFloat(details.MaterialSiteMinutesPerCycleExplicit, 2))
	}
	if details.DrivingMinutesPerCycleExplicit != nil {
		fmt.Fprintf(out, "Driving Minutes Per Cycle Explicit: %s\n", formatFloat(details.DrivingMinutesPerCycleExplicit, 2))
	}
	fmt.Fprintf(out, "Planned Non-Production Truck Count: %d\n", details.PlannedNonProductionTruckCount)
	if details.AverageActualPracticalSurplus != nil {
		fmt.Fprintf(out, "Average Actual Practical Surplus: %s\n", formatFloat(details.AverageActualPracticalSurplus, 2))
	}
	if details.SumActualPracticalSurplus != nil {
		fmt.Fprintf(out, "Sum Actual Practical Surplus: %s\n", formatFloat(details.SumActualPracticalSurplus, 2))
	}
	if details.AveragePlannedPracticalSurplus != nil {
		fmt.Fprintf(out, "Average Planned Practical Surplus: %s\n", formatFloat(details.AveragePlannedPracticalSurplus, 2))
	}
	if details.SumPlannedPracticalSurplus != nil {
		fmt.Fprintf(out, "Sum Planned Practical Surplus: %s\n", formatFloat(details.SumPlannedPracticalSurplus, 2))
	}
	if details.ActualPracticalSurplusPct != nil {
		fmt.Fprintf(out, "Actual Practical Surplus Pct: %s\n", formatPercent(details.ActualPracticalSurplusPct))
	}
	if details.PlannedPracticalSurplusPct != nil {
		fmt.Fprintf(out, "Planned Practical Surplus Pct: %s\n", formatPercent(details.PlannedPracticalSurplusPct))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	if details.PlannedMaterialTransactions != nil {
		fmt.Fprintln(out, "\nPlanned Material Transactions:")
		fmt.Fprintln(out, formatJSONBlock(details.PlannedMaterialTransactions, "  "))
	}
	if details.ActualMaterialTransactions != nil {
		fmt.Fprintln(out, "\nActual Material Transactions:")
		fmt.Fprintln(out, formatJSONBlock(details.ActualMaterialTransactions, "  "))
	}
	if details.PlannedTrucks != nil {
		fmt.Fprintln(out, "\nPlanned Trucks:")
		fmt.Fprintln(out, formatJSONBlock(details.PlannedTrucks, "  "))
	}
	if details.ActualTrucks != nil {
		fmt.Fprintln(out, "\nActual Trucks:")
		fmt.Fprintln(out, formatJSONBlock(details.ActualTrucks, "  "))
	}
	if details.PlannedBalances != nil {
		fmt.Fprintln(out, "\nPlanned Balances:")
		fmt.Fprintln(out, formatJSONBlock(details.PlannedBalances, "  "))
	}
	if details.ActualBalances != nil {
		fmt.Fprintln(out, "\nActual Balances:")
		fmt.Fprintln(out, formatJSONBlock(details.ActualBalances, "  "))
	}

	return nil
}

func formatJSONBlock(value any, indent string) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	if indent == "" {
		return string(pretty)
	}
	lines := strings.Split(string(pretty), "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

func formatFloat(value *float64, precision int) string {
	if value == nil {
		return ""
	}
	format := "%." + strconv.Itoa(precision) + "f"
	return fmt.Sprintf(format, *value)
}

func formatPercent(value *float64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%.1f%%", *value*100)
}
