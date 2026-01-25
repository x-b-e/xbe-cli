package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doSuperiorBowenCrewLedgersCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
}

type superiorBowenCrewLedgerRow struct {
	ID                  string `json:"id"`
	JobProductionPlanID string `json:"job_production_plan_id"`
	LaborHours          string `json:"labor_hours"`
	LaborCost           string `json:"labor_cost"`
	EquipmentHours      string `json:"equipment_hours"`
	EquipmentCost       string `json:"equipment_cost"`
	TotalCost           string `json:"total_cost"`
	CrewSize            string `json:"crew_size"`
	CostCode            string `json:"cost_code"`
	Prem                string `json:"prem"`
	ProductionQuantity  string `json:"production_quantity"`
}

func newDoSuperiorBowenCrewLedgersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Superior Bowen crew ledger",
		Long: `Create a Superior Bowen crew ledger for a job production plan.

Required flags:
  --job-production-plan   Job production plan ID (required)

Notes:
  - Only admin users can create Superior Bowen crew ledgers.
  - The job production plan must have a planner and raw job number configured.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a Superior Bowen crew ledger
  xbe do superior-bowen-crew-ledgers create --job-production-plan 123

  # Output as JSON
  xbe do superior-bowen-crew-ledgers create --job-production-plan 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoSuperiorBowenCrewLedgersCreate,
	}
	initDoSuperiorBowenCrewLedgersCreateFlags(cmd)
	return cmd
}

func init() {
	doSuperiorBowenCrewLedgersCmd.AddCommand(newDoSuperiorBowenCrewLedgersCreateCmd())
}

func initDoSuperiorBowenCrewLedgersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoSuperiorBowenCrewLedgersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoSuperiorBowenCrewLedgersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	opts.JobProductionPlan = strings.TrimSpace(opts.JobProductionPlan)
	if opts.JobProductionPlan == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "superior-bowen-crew-ledgers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/superior-bowen-crew-ledgers", jsonBody)
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

	row := superiorBowenCrewLedgerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created Superior Bowen crew ledger %s\n", row.ID)
	return nil
}

func superiorBowenCrewLedgerRowFromSingle(resp jsonAPISingleResponse) superiorBowenCrewLedgerRow {
	attrs := resp.Data.Attributes

	row := superiorBowenCrewLedgerRow{
		ID:                 resp.Data.ID,
		LaborHours:         stringAttr(attrs, "labor-hours"),
		LaborCost:          stringAttr(attrs, "labor-cost"),
		EquipmentHours:     stringAttr(attrs, "equipment-hours"),
		EquipmentCost:      stringAttr(attrs, "equipment-cost"),
		TotalCost:          stringAttr(attrs, "total-cost"),
		CrewSize:           stringAttr(attrs, "crew-size"),
		CostCode:           stringAttr(attrs, "cost-code"),
		Prem:               stringAttr(attrs, "prem"),
		ProductionQuantity: stringAttr(attrs, "production-quantity"),
	}

	if row.LaborHours == "" {
		row.LaborHours = stringAttr(attrs, "labor_hours")
	}
	if row.LaborCost == "" {
		row.LaborCost = stringAttr(attrs, "labor_cost")
	}
	if row.EquipmentHours == "" {
		row.EquipmentHours = stringAttr(attrs, "equipment_hours")
	}
	if row.EquipmentCost == "" {
		row.EquipmentCost = stringAttr(attrs, "equipment_cost")
	}
	if row.TotalCost == "" {
		row.TotalCost = stringAttr(attrs, "total_cost")
	}
	if row.CrewSize == "" {
		row.CrewSize = stringAttr(attrs, "crew_size")
	}
	if row.CostCode == "" {
		row.CostCode = stringAttr(attrs, "cost_code")
	}
	if row.Prem == "" {
		row.Prem = stringAttr(attrs, "prem")
	}
	if row.ProductionQuantity == "" {
		row.ProductionQuantity = stringAttr(attrs, "production_quantity")
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}

func parseDoSuperiorBowenCrewLedgersCreateOptions(cmd *cobra.Command) (doSuperiorBowenCrewLedgersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doSuperiorBowenCrewLedgersCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
	}, nil
}
