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

type doEquipmentRentalsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	Broker                  string
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

func newDoEquipmentRentalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment rental",
		Long: `Create an equipment rental.

Required:
  --broker                   Broker ID

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
		Example: `  # Create an equipment rental
  xbe do equipment-rentals create --broker 123

  # Create with full details
  xbe do equipment-rentals create --broker 123 --description "Excavator Rental" \
    --status active --start-on "2024-01-15" --end-on-planned "2024-02-15"

  # Create with cost information
  xbe do equipment-rentals create --broker 123 --approximate-cost-per-day "500.00" \
    --target-utilization-hours "8"`,
		RunE: runDoEquipmentRentalsCreate,
	}
	initDoEquipmentRentalsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentRentalsCmd.AddCommand(newDoEquipmentRentalsCreateCmd())
}

func initDoEquipmentRentalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
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

	_ = cmd.MarkFlagRequired("broker")
}

func runDoEquipmentRentalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentRentalsCreateOptions(cmd)
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

	if opts.Description != "" {
		attributes["description"] = opts.Description
	}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.StartOn != "" {
		attributes["start-on"] = opts.StartOn
	}
	if opts.StartOnPlanned != "" {
		attributes["start-on-planned"] = opts.StartOnPlanned
	}
	if opts.EndOn != "" {
		attributes["end-on"] = opts.EndOn
	}
	if opts.EndOnPlanned != "" {
		attributes["end-on-planned"] = opts.EndOnPlanned
	}
	if opts.ApproximateCostPerDay != "" {
		attributes["approximate-cost-per-day"] = opts.ApproximateCostPerDay
	}
	if opts.CostPerHour != "" {
		attributes["cost-per-hour"] = opts.CostPerHour
	}
	if opts.TargetUtilizationHours != "" {
		attributes["target-utilization-hours"] = opts.TargetUtilizationHours
	}
	if opts.ActualRentalUsageHours != "" {
		attributes["actual-rental-usage-hours"] = opts.ActualRentalUsageHours
	}
	if cmd.Flags().Changed("skip-weekend") {
		attributes["skip-weekend"] = opts.SkipWeekend
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	if opts.EquipmentClassification != "" {
		relationships["equipment-classification"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-classifications",
				"id":   opts.EquipmentClassification,
			},
		}
	}
	if opts.Equipment != "" {
		relationships["equipment"] = map[string]any{
			"data": map[string]any{
				"type": "equipment",
				"id":   opts.Equipment,
			},
		}
	}
	if opts.EquipmentSupplier != "" {
		relationships["equipment-supplier"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-suppliers",
				"id":   opts.EquipmentSupplier,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-rentals",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-rentals", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment rental %s\n", resp.Data.ID)
	return nil
}

func parseDoEquipmentRentalsCreateOptions(cmd *cobra.Command) (doEquipmentRentalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
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

	return doEquipmentRentalsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		Broker:                  broker,
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
