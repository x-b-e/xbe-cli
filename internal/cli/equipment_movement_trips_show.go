package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type equipmentMovementTripsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type equipmentMovementTripDetails struct {
	ID                                         string   `json:"id"`
	JobNumber                                  string   `json:"job_number,omitempty"`
	BrokerID                                   string   `json:"broker_id,omitempty"`
	TrailerClassificationID                    string   `json:"trailer_classification_id,omitempty"`
	TrailerClassificationEquivalentIDs         []string `json:"trailer_classification_equivalent_ids,omitempty"`
	ServiceTypeUnitOfMeasureIDs                []string `json:"service_type_unit_of_measure_ids,omitempty"`
	ExplicitDriverDayMobilizationBeforeMinutes string   `json:"explicit_driver_day_mobilization_before_minutes,omitempty"`
	JobProductionPlanID                        string   `json:"job_production_plan_id,omitempty"`
	CustomerCostAllocationID                   string   `json:"customer_cost_allocation_id,omitempty"`
	StopIDs                                    []string `json:"stop_ids,omitempty"`
}

func newEquipmentMovementTripsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show equipment movement trip details",
		Long: `Show the full details of an equipment movement trip.

Output Fields:
  ID
  Job Number
  Broker ID
  Trailer Classification ID
  Trailer Classification Equivalent IDs
  Service Type Unit Of Measure IDs
  Explicit Driver Day Mobilization Before Minutes
  Job Production Plan ID
  Customer Cost Allocation ID
  Stops

Arguments:
  <id>    The trip ID (required). Use the list command to find IDs.`,
		Example: `  # Show an equipment movement trip
  xbe view equipment-movement-trips show 123

  # JSON output
  xbe view equipment-movement-trips show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runEquipmentMovementTripsShow,
	}
	initEquipmentMovementTripsShowFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementTripsCmd.AddCommand(newEquipmentMovementTripsShowCmd())
}

func initEquipmentMovementTripsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementTripsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseEquipmentMovementTripsShowOptions(cmd)
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
		return fmt.Errorf("equipment movement trip id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "broker,trailer-classification,job-production-plan,customer-cost-allocation,stops")

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-trips/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildEquipmentMovementTripDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderEquipmentMovementTripDetails(cmd, details)
}

func parseEquipmentMovementTripsShowOptions(cmd *cobra.Command) (equipmentMovementTripsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementTripsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildEquipmentMovementTripDetails(resp jsonAPISingleResponse) equipmentMovementTripDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := equipmentMovementTripDetails{
		ID:                                 resource.ID,
		JobNumber:                          stringAttr(attrs, "job-number"),
		TrailerClassificationEquivalentIDs: stringSliceAttr(attrs, "trailer-classification-equivalent-ids"),
		ServiceTypeUnitOfMeasureIDs:        stringSliceAttr(attrs, "service-type-unit-of-measure-ids"),
		ExplicitDriverDayMobilizationBeforeMinutes: stringAttr(attrs, "explicit-driver-day-mobilization-before-minutes"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer-classification"]; ok && rel.Data != nil {
		details.TrailerClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer-cost-allocation"]; ok && rel.Data != nil {
		details.CustomerCostAllocationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["stops"]; ok && rel.raw != nil {
		for _, ref := range relationshipIDs(rel) {
			if ref.ID != "" {
				details.StopIDs = append(details.StopIDs, ref.ID)
			}
		}
	}

	return details
}

func renderEquipmentMovementTripDetails(cmd *cobra.Command, details equipmentMovementTripDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobNumber != "" {
		fmt.Fprintf(out, "Job Number: %s\n", details.JobNumber)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.TrailerClassificationID != "" {
		fmt.Fprintf(out, "Trailer Classification ID: %s\n", details.TrailerClassificationID)
	}
	if len(details.TrailerClassificationEquivalentIDs) > 0 {
		fmt.Fprintf(out, "Trailer Classification Equivalent IDs: %s\n", strings.Join(details.TrailerClassificationEquivalentIDs, ", "))
	}
	if len(details.ServiceTypeUnitOfMeasureIDs) > 0 {
		fmt.Fprintf(out, "Service Type Unit Of Measure IDs: %s\n", strings.Join(details.ServiceTypeUnitOfMeasureIDs, ", "))
	}
	if details.ExplicitDriverDayMobilizationBeforeMinutes != "" {
		fmt.Fprintf(out, "Explicit Driver Day Mobilization Before Minutes: %s\n", details.ExplicitDriverDayMobilizationBeforeMinutes)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.CustomerCostAllocationID != "" {
		fmt.Fprintf(out, "Customer Cost Allocation ID: %s\n", details.CustomerCostAllocationID)
	}
	if len(details.StopIDs) > 0 {
		fmt.Fprintf(out, "Stops: %s\n", strings.Join(details.StopIDs, ", "))
	}

	return nil
}
