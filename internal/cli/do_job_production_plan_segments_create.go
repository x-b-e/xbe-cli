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

type doJobProductionPlanSegmentsCreateOptions struct {
	BaseURL                                             string
	Token                                               string
	JSON                                                bool
	JobProductionPlanID                                 string
	JobProductionPlanSegmentSetID                       string
	MaterialSiteID                                      string
	MaterialTypeID                                      string
	CostCodeID                                          string
	ExplicitMaterialTypeMaterialSiteInventoryLocationID string
	Description                                         string
	NonProductionMinutes                                string
	IsExpectingWeighedTransactions                      string
	ExplicitStartSiteKind                               string
	ObservedPossibleCycleMinutes                        string
	LockObservedPossibleCycleMinutes                    string
	Quantity                                            string
	QuantityPerHour                                     string
	SelectedGoogleRoute                                 string
	SequencePosition                                    string
	PlannedUnproductiveMinutesPerHour                   string
	DrivingMinutesPerCycle                              string
	MaterialSiteMinutesPerCycle                         string
	TonsPerCycle                                        string
}

func newDoJobProductionPlanSegmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan segment",
		Long: `Create a job production plan segment.

Required flags:
  --job-production-plan   Job production plan ID (required)

Optional attributes:
  --description                          Description
  --non-production-minutes               Non production minutes
  --is-expecting-weighed-transactions    Expect weighed transactions (true/false)
  --explicit-start-site-kind             Explicit start site kind (material_site, job_site)
  --observed-possible-cycle-minutes      Observed possible cycle minutes
  --lock-observed-possible-cycle-minutes Lock observed possible cycle minutes (true/false)
  --quantity                             Planned quantity
  --quantity-per-hour                    Planned quantity per hour
  --selected-google-route                Selected google route (JSON)
  --sequence-position                    Sequence position
  --planned-unproductive-minutes-per-hour Planned unproductive minutes per hour
  --driving-minutes-per-cycle            Driving minutes per cycle
  --material-site-minutes-per-cycle      Material site minutes per cycle
  --tons-per-cycle                       Tons per cycle

Optional relationships:
  --job-production-plan-segment-set                 Segment set ID
  --material-site                                  Material site ID
  --material-type                                  Material type ID
  --cost-code                                      Cost code ID
  --explicit-material-type-material-site-inventory-location  Explicit inventory location ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a basic segment
  xbe do job-production-plan-segments create --job-production-plan 123 --description "Main haul" --quantity 250

  # Create with materials
  xbe do job-production-plan-segments create \
    --job-production-plan 123 \
    --material-site 456 \
    --material-type 789 \
    --quantity 500 \
    --quantity-per-hour 100`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanSegmentsCreate,
	}
	initDoJobProductionPlanSegmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanSegmentsCmd.AddCommand(newDoJobProductionPlanSegmentsCreateCmd())
}

func initDoJobProductionPlanSegmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("job-production-plan-segment-set", "", "Segment set ID")
	cmd.Flags().String("material-site", "", "Material site ID")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("cost-code", "", "Cost code ID")
	cmd.Flags().String("explicit-material-type-material-site-inventory-location", "", "Explicit inventory location ID")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("non-production-minutes", "", "Non production minutes")
	cmd.Flags().String("is-expecting-weighed-transactions", "", "Expect weighed transactions (true/false)")
	cmd.Flags().String("explicit-start-site-kind", "", "Explicit start site kind (material_site, job_site)")
	cmd.Flags().String("observed-possible-cycle-minutes", "", "Observed possible cycle minutes")
	cmd.Flags().String("lock-observed-possible-cycle-minutes", "", "Lock observed possible cycle minutes (true/false)")
	cmd.Flags().String("quantity", "", "Planned quantity")
	cmd.Flags().String("quantity-per-hour", "", "Planned quantity per hour")
	cmd.Flags().String("selected-google-route", "", "Selected google route (JSON)")
	cmd.Flags().String("sequence-position", "", "Sequence position")
	cmd.Flags().String("planned-unproductive-minutes-per-hour", "", "Planned unproductive minutes per hour")
	cmd.Flags().String("driving-minutes-per-cycle", "", "Driving minutes per cycle")
	cmd.Flags().String("material-site-minutes-per-cycle", "", "Material site minutes per cycle")
	cmd.Flags().String("tons-per-cycle", "", "Tons per cycle")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanSegmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanSegmentsCreateOptions(cmd)
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

	jobProductionPlanID := strings.TrimSpace(opts.JobProductionPlanID)
	if jobProductionPlanID == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}
	if cmd.Flags().Changed("non-production-minutes") {
		attributes["non-production-minutes"] = opts.NonProductionMinutes
	}
	if cmd.Flags().Changed("is-expecting-weighed-transactions") {
		attributes["is-expecting-weighed-transactions"] = opts.IsExpectingWeighedTransactions == "true"
	}
	if cmd.Flags().Changed("explicit-start-site-kind") {
		attributes["explicit-start-site-kind"] = opts.ExplicitStartSiteKind
	}
	if cmd.Flags().Changed("observed-possible-cycle-minutes") {
		attributes["observed-possible-cycle-minutes"] = opts.ObservedPossibleCycleMinutes
	}
	if cmd.Flags().Changed("lock-observed-possible-cycle-minutes") {
		attributes["lock-observed-possible-cycle-minutes"] = opts.LockObservedPossibleCycleMinutes == "true"
	}
	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("quantity-per-hour") {
		attributes["quantity-per-hour"] = opts.QuantityPerHour
	}
	if cmd.Flags().Changed("selected-google-route") {
		if strings.TrimSpace(opts.SelectedGoogleRoute) == "" {
			attributes["selected-google-route"] = nil
		} else {
			value, err := parseRawJSON(opts.SelectedGoogleRoute)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return err
			}
			attributes["selected-google-route"] = value
		}
	}
	if cmd.Flags().Changed("sequence-position") {
		attributes["sequence-position"] = opts.SequencePosition
	}
	if cmd.Flags().Changed("planned-unproductive-minutes-per-hour") {
		attributes["planned-unproductive-minutes-per-hour"] = opts.PlannedUnproductiveMinutesPerHour
	}
	if cmd.Flags().Changed("driving-minutes-per-cycle") {
		attributes["driving-minutes-per-cycle"] = opts.DrivingMinutesPerCycle
	}
	if cmd.Flags().Changed("material-site-minutes-per-cycle") {
		attributes["material-site-minutes-per-cycle"] = opts.MaterialSiteMinutesPerCycle
	}
	if cmd.Flags().Changed("tons-per-cycle") {
		attributes["tons-per-cycle"] = opts.TonsPerCycle
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   jobProductionPlanID,
			},
		},
	}

	if opts.JobProductionPlanSegmentSetID != "" {
		relationships["job-production-plan-segment-set"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-segment-sets",
				"id":   opts.JobProductionPlanSegmentSetID,
			},
		}
	}
	if opts.MaterialSiteID != "" {
		relationships["material-site"] = map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSiteID,
			},
		}
	}
	if opts.MaterialTypeID != "" {
		relationships["material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialTypeID,
			},
		}
	}
	if opts.CostCodeID != "" {
		relationships["cost-code"] = map[string]any{
			"data": map[string]any{
				"type": "cost-codes",
				"id":   opts.CostCodeID,
			},
		}
	}
	if opts.ExplicitMaterialTypeMaterialSiteInventoryLocationID != "" {
		relationships["explicit-material-type-material-site-inventory-location"] = map[string]any{
			"data": map[string]any{
				"type": "material-type-material-site-inventory-locations",
				"id":   opts.ExplicitMaterialTypeMaterialSiteInventoryLocationID,
			},
		}
	}

	requestData := map[string]any{
		"type":          "job-production-plan-segments",
		"attributes":    attributes,
		"relationships": relationships,
	}

	requestBody := map[string]any{"data": requestData}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-segments", jsonBody)
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

	row := buildJobProductionPlanSegmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan segment %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanSegmentsCreateOptions(cmd *cobra.Command) (doJobProductionPlanSegmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	jobProductionPlanSegmentSetID, _ := cmd.Flags().GetString("job-production-plan-segment-set")
	materialSiteID, _ := cmd.Flags().GetString("material-site")
	materialTypeID, _ := cmd.Flags().GetString("material-type")
	costCodeID, _ := cmd.Flags().GetString("cost-code")
	explicitMTMSILID, _ := cmd.Flags().GetString("explicit-material-type-material-site-inventory-location")
	description, _ := cmd.Flags().GetString("description")
	nonProductionMinutes, _ := cmd.Flags().GetString("non-production-minutes")
	isExpectingWeighedTransactions, _ := cmd.Flags().GetString("is-expecting-weighed-transactions")
	explicitStartSiteKind, _ := cmd.Flags().GetString("explicit-start-site-kind")
	observedPossibleCycleMinutes, _ := cmd.Flags().GetString("observed-possible-cycle-minutes")
	lockObservedPossibleCycleMinutes, _ := cmd.Flags().GetString("lock-observed-possible-cycle-minutes")
	quantity, _ := cmd.Flags().GetString("quantity")
	quantityPerHour, _ := cmd.Flags().GetString("quantity-per-hour")
	selectedGoogleRoute, _ := cmd.Flags().GetString("selected-google-route")
	sequencePosition, _ := cmd.Flags().GetString("sequence-position")
	plannedUnproductiveMinutesPerHour, _ := cmd.Flags().GetString("planned-unproductive-minutes-per-hour")
	drivingMinutesPerCycle, _ := cmd.Flags().GetString("driving-minutes-per-cycle")
	materialSiteMinutesPerCycle, _ := cmd.Flags().GetString("material-site-minutes-per-cycle")
	tonsPerCycle, _ := cmd.Flags().GetString("tons-per-cycle")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanSegmentsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		JobProductionPlanID:           jobProductionPlanID,
		JobProductionPlanSegmentSetID: jobProductionPlanSegmentSetID,
		MaterialSiteID:                materialSiteID,
		MaterialTypeID:                materialTypeID,
		CostCodeID:                    costCodeID,
		ExplicitMaterialTypeMaterialSiteInventoryLocationID: explicitMTMSILID,
		Description:                       description,
		NonProductionMinutes:              nonProductionMinutes,
		IsExpectingWeighedTransactions:    isExpectingWeighedTransactions,
		ExplicitStartSiteKind:             explicitStartSiteKind,
		ObservedPossibleCycleMinutes:      observedPossibleCycleMinutes,
		LockObservedPossibleCycleMinutes:  lockObservedPossibleCycleMinutes,
		Quantity:                          quantity,
		QuantityPerHour:                   quantityPerHour,
		SelectedGoogleRoute:               selectedGoogleRoute,
		SequencePosition:                  sequencePosition,
		PlannedUnproductiveMinutesPerHour: plannedUnproductiveMinutesPerHour,
		DrivingMinutesPerCycle:            drivingMinutesPerCycle,
		MaterialSiteMinutesPerCycle:       materialSiteMinutesPerCycle,
		TonsPerCycle:                      tonsPerCycle,
	}, nil
}

func buildJobProductionPlanSegmentRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanSegmentRow {
	resource := resp.Data
	row := jobProductionPlanSegmentRow{
		ID:          resource.ID,
		Description: stringAttr(resource.Attributes, "description"),
		Sequence:    stringAttr(resource.Attributes, "sequence"),
		Quantity:    stringAttr(resource.Attributes, "quantity"),
	}

	row.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")
	row.JobProductionPlanSegmentSetID = relationshipIDFromMap(resource.Relationships, "job-production-plan-segment-set")
	row.MaterialSiteID = relationshipIDFromMap(resource.Relationships, "material-site")
	row.MaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-type")
	row.CostCodeID = relationshipIDFromMap(resource.Relationships, "cost-code")

	return row
}

func parseRawJSON(raw string) (any, error) {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("invalid JSON for selected-google-route: %w", err)
	}
	return value, nil
}
