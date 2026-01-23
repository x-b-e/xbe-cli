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

type doEquipmentRentalsUpdateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ID                      string
	EquipmentClassification string
	Equipment               string
	EquipmentSupplier       string
	Description             string
	Status                  string
	StartOn                 string
	StartOnPlanned          string
	EndOn                   string
	EndOnPlanned            string
	ApproximateCostPerDay   string
	CostPerHour             string
	TargetUtilizationHours  string
	ActualRentalUsageHours  string
	SkipWeekend             bool
}

func newDoEquipmentRentalsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment rental",
		Long: `Update an equipment rental.

Optional:
  --equipment-classification Equipment classification ID
  --equipment                Equipment ID
  --equipment-supplier       Equipment supplier ID
  --description              Description
  --status                   Status
  --start-on                 Actual start date
  --start-on-planned         Planned start date
  --end-on                   Actual end date
  --end-on-planned           Planned end date
  --approximate-cost-per-day Approximate cost per day
  --cost-per-hour            Cost per hour
  --target-utilization-hours Target utilization hours
  --actual-rental-usage-hours Actual rental usage hours
  --skip-weekend             Skip weekends in calculations`,
		Example: `  # Update description
  xbe do equipment-rentals update 123 --description "Updated description"

  # Update status
  xbe do equipment-rentals update 123 --status completed

  # Update dates
  xbe do equipment-rentals update 123 --end-on "2024-02-20"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentRentalsUpdate,
	}
	initDoEquipmentRentalsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentRentalsCmd.AddCommand(newDoEquipmentRentalsUpdateCmd())
}

func initDoEquipmentRentalsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment-classification", "", "Equipment classification ID")
	cmd.Flags().String("equipment", "", "Equipment ID")
	cmd.Flags().String("equipment-supplier", "", "Equipment supplier ID")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("status", "", "Status")
	cmd.Flags().String("start-on", "", "Actual start date")
	cmd.Flags().String("start-on-planned", "", "Planned start date")
	cmd.Flags().String("end-on", "", "Actual end date")
	cmd.Flags().String("end-on-planned", "", "Planned end date")
	cmd.Flags().String("approximate-cost-per-day", "", "Approximate cost per day")
	cmd.Flags().String("cost-per-hour", "", "Cost per hour")
	cmd.Flags().String("target-utilization-hours", "", "Target utilization hours")
	cmd.Flags().String("actual-rental-usage-hours", "", "Actual rental usage hours")
	cmd.Flags().Bool("skip-weekend", false, "Skip weekends in calculations")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentRentalsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentRentalsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
	}
	if cmd.Flags().Changed("start-on-planned") {
		attributes["start-on-planned"] = opts.StartOnPlanned
	}
	if cmd.Flags().Changed("end-on") {
		attributes["end-on"] = opts.EndOn
	}
	if cmd.Flags().Changed("end-on-planned") {
		attributes["end-on-planned"] = opts.EndOnPlanned
	}
	if cmd.Flags().Changed("approximate-cost-per-day") {
		attributes["approximate-cost-per-day"] = opts.ApproximateCostPerDay
	}
	if cmd.Flags().Changed("cost-per-hour") {
		attributes["cost-per-hour"] = opts.CostPerHour
	}
	if cmd.Flags().Changed("target-utilization-hours") {
		attributes["target-utilization-hours"] = opts.TargetUtilizationHours
	}
	if cmd.Flags().Changed("actual-rental-usage-hours") {
		attributes["actual-rental-usage-hours"] = opts.ActualRentalUsageHours
	}
	if cmd.Flags().Changed("skip-weekend") {
		attributes["skip-weekend"] = opts.SkipWeekend
	}

	if cmd.Flags().Changed("equipment-classification") {
		relationships["equipment-classification"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassification,
			},
		}
	}
	if cmd.Flags().Changed("equipment") {
		relationships["equipment"] = map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if cmd.Flags().Changed("equipment-supplier") {
		relationships["equipment-supplier"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-suppliers",
				"id":   opts.EquipmentSupplier,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "equipment-rentals",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-rentals/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := equipmentRentalRow{
			ID:          resp.Data.ID,
			Description: stringAttr(resp.Data.Attributes, "description"),
			Status:      stringAttr(resp.Data.Attributes, "status"),
			StartOn:     stringAttr(resp.Data.Attributes, "start-on"),
			EndOn:       stringAttr(resp.Data.Attributes, "end-on"),
		}
		if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment rental %s\n", resp.Data.ID)
	return nil
}

func parseDoEquipmentRentalsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentRentalsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentClassification, _ := cmd.Flags().GetString("equipment-classification")
	equipment, _ := cmd.Flags().GetString("equipment")
	equipmentSupplier, _ := cmd.Flags().GetString("equipment-supplier")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnPlanned, _ := cmd.Flags().GetString("start-on-planned")
	endOn, _ := cmd.Flags().GetString("end-on")
	endOnPlanned, _ := cmd.Flags().GetString("end-on-planned")
	approximateCostPerDay, _ := cmd.Flags().GetString("approximate-cost-per-day")
	costPerHour, _ := cmd.Flags().GetString("cost-per-hour")
	targetUtilizationHours, _ := cmd.Flags().GetString("target-utilization-hours")
	actualRentalUsageHours, _ := cmd.Flags().GetString("actual-rental-usage-hours")
	skipWeekend, _ := cmd.Flags().GetBool("skip-weekend")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentRentalsUpdateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ID:                      args[0],
		EquipmentClassification: equipmentClassification,
		Equipment:               equipment,
		EquipmentSupplier:       equipmentSupplier,
		Description:             description,
		Status:                  status,
		StartOn:                 startOn,
		StartOnPlanned:          startOnPlanned,
		EndOn:                   endOn,
		EndOnPlanned:            endOnPlanned,
		ApproximateCostPerDay:   approximateCostPerDay,
		CostPerHour:             costPerHour,
		TargetUtilizationHours:  targetUtilizationHours,
		ActualRentalUsageHours:  actualRentalUsageHours,
		SkipWeekend:             skipWeekend,
	}, nil
}
