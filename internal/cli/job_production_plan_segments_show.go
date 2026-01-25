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

type jobProductionPlanSegmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanSegmentDetails struct {
	ID                                                  string `json:"id"`
	JobProductionPlanID                                 string `json:"job_production_plan_id,omitempty"`
	JobProductionPlanSegmentSetID                       string `json:"job_production_plan_segment_set_id,omitempty"`
	MaterialSiteID                                      string `json:"material_site_id,omitempty"`
	MaterialTypeID                                      string `json:"material_type_id,omitempty"`
	CostCodeID                                          string `json:"cost_code_id,omitempty"`
	JobProductionPlanLocationID                         string `json:"job_production_plan_location_id,omitempty"`
	ProductionMeasurementID                             string `json:"production_measurement_id,omitempty"`
	ExplicitMaterialTypeMaterialSiteInventoryLocationID string `json:"explicit_material_type_material_site_inventory_location_id,omitempty"`
	MaterialTypeMaterialSiteInventoryLocationID         string `json:"material_type_material_site_inventory_location_id,omitempty"`
	Description                                         string `json:"description,omitempty"`
	NonProductionMinutes                                string `json:"non_production_minutes,omitempty"`
	IsExpectingWeighedTransactions                      bool   `json:"is_expecting_weighed_transactions"`
	ExplicitStartSiteKind                               string `json:"explicit_start_site_kind,omitempty"`
	ObservedPossibleCycleMinutes                        string `json:"observed_possible_cycle_minutes,omitempty"`
	LockObservedPossibleCycleMinutes                    bool   `json:"lock_observed_possible_cycle_minutes"`
	StartSiteKind                                       string `json:"start_site_kind,omitempty"`
	ProductionMinutes                                   string `json:"production_minutes,omitempty"`
	PlannedMinutes                                      string `json:"planned_minutes,omitempty"`
	Quantity                                            string `json:"quantity,omitempty"`
	QuantityPerHour                                     string `json:"quantity_per_hour,omitempty"`
	SelectedGoogleRoute                                 string `json:"selected_google_route,omitempty"`
	CalculatedMiles                                     string `json:"calculated_miles,omitempty"`
	Sequence                                            string `json:"sequence,omitempty"`
	SequenceIndex                                       string `json:"sequence_index,omitempty"`
	SequencePosition                                    string `json:"sequence_position,omitempty"`
	PlannedUnproductiveMinutesPerHour                   string `json:"planned_unproductive_minutes_per_hour,omitempty"`
	PlannedProductiveMinutesPerHour                     string `json:"planned_productive_minutes_per_hour,omitempty"`
	PeakProductionRatePerHour                           string `json:"peak_production_rate_per_hour,omitempty"`
	DrivingMinutesPerCycle                              string `json:"driving_minutes_per_cycle,omitempty"`
	MaterialSiteMinutesPerCycle                         string `json:"material_site_minutes_per_cycle,omitempty"`
	TonsPerCycle                                        string `json:"tons_per_cycle,omitempty"`
	CreatedAt                                           string `json:"created_at,omitempty"`
	UpdatedAt                                           string `json:"updated_at,omitempty"`
}

func newJobProductionPlanSegmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan segment details",
		Long: `Show the full details of a job production plan segment.

Output Fields:
  ID
  Job Production Plan ID
  Job Production Plan Segment Set ID
  Material Site ID
  Material Type ID
  Cost Code ID
  Job Production Plan Location ID
  Production Measurement ID
  Explicit Material Type Material Site Inventory Location ID
  Material Type Material Site Inventory Location ID
  Description
  Non Production Minutes
  Is Expecting Weighed Transactions
  Explicit Start Site Kind
  Observed Possible Cycle Minutes
  Lock Observed Possible Cycle Minutes
  Start Site Kind
  Production Minutes
  Planned Minutes
  Quantity
  Quantity Per Hour
  Selected Google Route
  Calculated Miles
  Sequence
  Sequence Index
  Sequence Position
  Planned Unproductive Minutes Per Hour
  Planned Productive Minutes Per Hour
  Peak Production Rate Per Hour
  Driving Minutes Per Cycle
  Material Site Minutes Per Cycle
  Tons Per Cycle
  Created At
  Updated At

Arguments:
  <id>    The segment ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a segment
  xbe view job-production-plan-segments show 123

  # JSON output
  xbe view job-production-plan-segments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanSegmentsShow,
	}
	initJobProductionPlanSegmentsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanSegmentsCmd.AddCommand(newJobProductionPlanSegmentsShowCmd())
}

func initJobProductionPlanSegmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanSegmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanSegmentsShowOptions(cmd)
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
		return fmt.Errorf("job production plan segment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-segments/"+id, nil)
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

	details := buildJobProductionPlanSegmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanSegmentDetails(cmd, details)
}

func parseJobProductionPlanSegmentsShowOptions(cmd *cobra.Command) (jobProductionPlanSegmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanSegmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanSegmentDetails(resp jsonAPISingleResponse) jobProductionPlanSegmentDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanSegmentDetails{
		ID:                                resource.ID,
		Description:                       stringAttr(attrs, "description"),
		NonProductionMinutes:              stringAttr(attrs, "non-production-minutes"),
		IsExpectingWeighedTransactions:    boolAttr(attrs, "is-expecting-weighed-transactions"),
		ExplicitStartSiteKind:             stringAttr(attrs, "explicit-start-site-kind"),
		ObservedPossibleCycleMinutes:      stringAttr(attrs, "observed-possible-cycle-minutes"),
		LockObservedPossibleCycleMinutes:  boolAttr(attrs, "lock-observed-possible-cycle-minutes"),
		StartSiteKind:                     stringAttr(attrs, "start-site-kind"),
		ProductionMinutes:                 stringAttr(attrs, "production-minutes"),
		PlannedMinutes:                    stringAttr(attrs, "planned-minutes"),
		Quantity:                          stringAttr(attrs, "quantity"),
		QuantityPerHour:                   stringAttr(attrs, "quantity-per-hour"),
		SelectedGoogleRoute:               stringAttr(attrs, "selected-google-route"),
		CalculatedMiles:                   stringAttr(attrs, "calculated-miles"),
		Sequence:                          stringAttr(attrs, "sequence"),
		SequenceIndex:                     stringAttr(attrs, "sequence-index"),
		SequencePosition:                  stringAttr(attrs, "sequence-position"),
		PlannedUnproductiveMinutesPerHour: stringAttr(attrs, "planned-unproductive-minutes-per-hour"),
		PlannedProductiveMinutesPerHour:   stringAttr(attrs, "planned-productive-minutes-per-hour"),
		PeakProductionRatePerHour:         stringAttr(attrs, "peak-production-rate-per-hour"),
		DrivingMinutesPerCycle:            stringAttr(attrs, "driving-minutes-per-cycle"),
		MaterialSiteMinutesPerCycle:       stringAttr(attrs, "material-site-minutes-per-cycle"),
		TonsPerCycle:                      stringAttr(attrs, "tons-per-cycle"),
		CreatedAt:                         formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                         formatDateTime(stringAttr(attrs, "updated-at")),
	}

	details.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")
	details.JobProductionPlanSegmentSetID = relationshipIDFromMap(resource.Relationships, "job-production-plan-segment-set")
	details.MaterialSiteID = relationshipIDFromMap(resource.Relationships, "material-site")
	details.MaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-type")
	details.CostCodeID = relationshipIDFromMap(resource.Relationships, "cost-code")
	details.JobProductionPlanLocationID = relationshipIDFromMap(resource.Relationships, "job-production-plan-location")
	details.ProductionMeasurementID = relationshipIDFromMap(resource.Relationships, "production-measurement")
	details.ExplicitMaterialTypeMaterialSiteInventoryLocationID = relationshipIDFromMap(resource.Relationships, "explicit-material-type-material-site-inventory-location")
	details.MaterialTypeMaterialSiteInventoryLocationID = relationshipIDFromMap(resource.Relationships, "material-type-material-site-inventory-location")

	return details
}

func renderJobProductionPlanSegmentDetails(cmd *cobra.Command, details jobProductionPlanSegmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.JobProductionPlanSegmentSetID != "" {
		fmt.Fprintf(out, "Job Production Plan Segment Set ID: %s\n", details.JobProductionPlanSegmentSetID)
	}
	if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site ID: %s\n", details.MaterialSiteID)
	}
	if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type ID: %s\n", details.MaterialTypeID)
	}
	if details.CostCodeID != "" {
		fmt.Fprintf(out, "Cost Code ID: %s\n", details.CostCodeID)
	}
	if details.JobProductionPlanLocationID != "" {
		fmt.Fprintf(out, "Job Production Plan Location ID: %s\n", details.JobProductionPlanLocationID)
	}
	if details.ProductionMeasurementID != "" {
		fmt.Fprintf(out, "Production Measurement ID: %s\n", details.ProductionMeasurementID)
	}
	if details.ExplicitMaterialTypeMaterialSiteInventoryLocationID != "" {
		fmt.Fprintf(out, "Explicit Material Type Material Site Inventory Location ID: %s\n", details.ExplicitMaterialTypeMaterialSiteInventoryLocationID)
	}
	if details.MaterialTypeMaterialSiteInventoryLocationID != "" {
		fmt.Fprintf(out, "Material Type Material Site Inventory Location ID: %s\n", details.MaterialTypeMaterialSiteInventoryLocationID)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	fmt.Fprintf(out, "Non Production Minutes: %s\n", details.NonProductionMinutes)
	fmt.Fprintf(out, "Is Expecting Weighed Transactions: %t\n", details.IsExpectingWeighedTransactions)
	if details.ExplicitStartSiteKind != "" {
		fmt.Fprintf(out, "Explicit Start Site Kind: %s\n", details.ExplicitStartSiteKind)
	}
	if details.ObservedPossibleCycleMinutes != "" {
		fmt.Fprintf(out, "Observed Possible Cycle Minutes: %s\n", details.ObservedPossibleCycleMinutes)
	}
	fmt.Fprintf(out, "Lock Observed Possible Cycle Minutes: %t\n", details.LockObservedPossibleCycleMinutes)
	if details.StartSiteKind != "" {
		fmt.Fprintf(out, "Start Site Kind: %s\n", details.StartSiteKind)
	}
	if details.ProductionMinutes != "" {
		fmt.Fprintf(out, "Production Minutes: %s\n", details.ProductionMinutes)
	}
	if details.PlannedMinutes != "" {
		fmt.Fprintf(out, "Planned Minutes: %s\n", details.PlannedMinutes)
	}
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.QuantityPerHour != "" {
		fmt.Fprintf(out, "Quantity Per Hour: %s\n", details.QuantityPerHour)
	}
	if details.SelectedGoogleRoute != "" {
		fmt.Fprintf(out, "Selected Google Route: %s\n", details.SelectedGoogleRoute)
	}
	if details.CalculatedMiles != "" {
		fmt.Fprintf(out, "Calculated Miles: %s\n", details.CalculatedMiles)
	}
	if details.Sequence != "" {
		fmt.Fprintf(out, "Sequence: %s\n", details.Sequence)
	}
	if details.SequenceIndex != "" {
		fmt.Fprintf(out, "Sequence Index: %s\n", details.SequenceIndex)
	}
	if details.SequencePosition != "" {
		fmt.Fprintf(out, "Sequence Position: %s\n", details.SequencePosition)
	}
	if details.PlannedUnproductiveMinutesPerHour != "" {
		fmt.Fprintf(out, "Planned Unproductive Minutes Per Hour: %s\n", details.PlannedUnproductiveMinutesPerHour)
	}
	if details.PlannedProductiveMinutesPerHour != "" {
		fmt.Fprintf(out, "Planned Productive Minutes Per Hour: %s\n", details.PlannedProductiveMinutesPerHour)
	}
	if details.PeakProductionRatePerHour != "" {
		fmt.Fprintf(out, "Peak Production Rate Per Hour: %s\n", details.PeakProductionRatePerHour)
	}
	if details.DrivingMinutesPerCycle != "" {
		fmt.Fprintf(out, "Driving Minutes Per Cycle: %s\n", details.DrivingMinutesPerCycle)
	}
	if details.MaterialSiteMinutesPerCycle != "" {
		fmt.Fprintf(out, "Material Site Minutes Per Cycle: %s\n", details.MaterialSiteMinutesPerCycle)
	}
	if details.TonsPerCycle != "" {
		fmt.Fprintf(out, "Tons Per Cycle: %s\n", details.TonsPerCycle)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
