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

type doJobProductionPlanDriverMovementsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	JobProductionPlanID          string
	TenderJobScheduleShiftID     string
	DriverID                     string
	ExplicitLocationEventSources []string
	BustCache                    bool
}

type jobProductionPlanDriverMovementRow struct {
	ID                                   string   `json:"id"`
	JobProductionPlanID                  string   `json:"job_production_plan_id,omitempty"`
	TenderJobScheduleShiftID             string   `json:"tender_job_schedule_shift_id,omitempty"`
	DriverID                             string   `json:"driver_id,omitempty"`
	ExplicitLocationEventSources         []string `json:"explicit_location_event_sources,omitempty"`
	BustCache                            bool     `json:"bust_cache"`
	RelativeLocationAttributes           any      `json:"relative_location_attributes,omitempty"`
	RelativeLocationValues               any      `json:"relative_location_values,omitempty"`
	RelativeLocationFragmentStartIndexes any      `json:"relative_location_fragment_start_indexes,omitempty"`
	SegmentMovementFragments             any      `json:"segment_movement_fragments,omitempty"`
	MaxTrackedMinutes                    any      `json:"max_tracked_minutes,omitempty"`
	TrackedMinutes                       any      `json:"tracked_minutes,omitempty"`
	MinEventAt                           string   `json:"min_event_at,omitempty"`
	MaxEventAt                           string   `json:"max_event_at,omitempty"`
}

func newDoJobProductionPlanDriverMovementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate driver movement details",
		Long: `Generate driver movement details for a job production plan.

Required flags:
  --job-production-plan          Job production plan ID
  --driver or --tender-job-schedule-shift

Optional flags:
  --explicit-location-event-sources  Restrict location event sources (comma-separated or repeated)
                                      Allowed values: user_device, vehicle_location_event
  --bust-cache                        Bust cached movement data`,
		Example: `  # Generate movement for a driver
  xbe do job-production-plan-driver-movements create \
    --job-production-plan 123 \
    --driver 456

  # Generate movement for a shift
  xbe do job-production-plan-driver-movements create \
    --job-production-plan 123 \
    --tender-job-schedule-shift 789

  # Restrict location event sources and bust cache
  xbe do job-production-plan-driver-movements create \
    --job-production-plan 123 \
    --tender-job-schedule-shift 789 \
    --explicit-location-event-sources user_device,vehicle_location_event \
    --bust-cache

  # JSON output
  xbe do job-production-plan-driver-movements create --job-production-plan 123 --driver 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanDriverMovementsCreate,
	}
	initDoJobProductionPlanDriverMovementsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanDriverMovementsCmd.AddCommand(newDoJobProductionPlanDriverMovementsCreateCmd())
}

func initDoJobProductionPlanDriverMovementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID")
	cmd.Flags().String("driver", "", "Driver (user) ID")
	cmd.Flags().StringSlice("explicit-location-event-sources", nil, "Location event sources (comma-separated or repeated)")
	cmd.Flags().Bool("bust-cache", false, "Bust cached movement data")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("job-production-plan")
}

func runDoJobProductionPlanDriverMovementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanDriverMovementsCreateOptions(cmd)
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
	tenderJobScheduleShiftID := strings.TrimSpace(opts.TenderJobScheduleShiftID)
	driverID := strings.TrimSpace(opts.DriverID)

	if jobProductionPlanID == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if tenderJobScheduleShiftID == "" && driverID == "" {
		err := fmt.Errorf("--driver or --tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	explicitSources := compactStringSlice(opts.ExplicitLocationEventSources)

	attributes := map[string]any{}
	if len(explicitSources) > 0 {
		attributes["explicit-location-event-sources"] = explicitSources
	}
	if cmd.Flags().Changed("bust-cache") {
		attributes["bust-cache"] = opts.BustCache
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   jobProductionPlanID,
			},
		},
	}

	if tenderJobScheduleShiftID != "" {
		relationships["tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   tenderJobScheduleShiftID,
			},
		}
	}

	if driverID != "" {
		relationships["driver"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   driverID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-driver-movements",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-driver-movements", jsonBody)
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

	row := buildJobProductionPlanDriverMovementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan driver movement %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanDriverMovementsCreateOptions(cmd *cobra.Command) (doJobProductionPlanDriverMovementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	driverID, _ := cmd.Flags().GetString("driver")
	explicitSources, _ := cmd.Flags().GetStringSlice("explicit-location-event-sources")
	bustCache, _ := cmd.Flags().GetBool("bust-cache")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanDriverMovementsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		JobProductionPlanID:          jobProductionPlanID,
		TenderJobScheduleShiftID:     tenderJobScheduleShiftID,
		DriverID:                     driverID,
		ExplicitLocationEventSources: explicitSources,
		BustCache:                    bustCache,
	}, nil
}

func buildJobProductionPlanDriverMovementRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanDriverMovementRow {
	resource := resp.Data
	row := jobProductionPlanDriverMovementRow{
		ID:                           resource.ID,
		ExplicitLocationEventSources: stringSliceAttr(resource.Attributes, "explicit-location-event-sources"),
		BustCache:                    boolAttr(resource.Attributes, "bust-cache"),
		MinEventAt:                   stringAttr(resource.Attributes, "min-event-at"),
		MaxEventAt:                   stringAttr(resource.Attributes, "max-event-at"),
	}

	row.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")
	row.TenderJobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "tender-job-schedule-shift")
	row.DriverID = relationshipIDFromMap(resource.Relationships, "driver")

	if attrs := resource.Attributes; attrs != nil {
		if value, ok := attrs["relative-location-attributes"]; ok {
			row.RelativeLocationAttributes = value
		}
		if value, ok := attrs["relative-location-values"]; ok {
			row.RelativeLocationValues = value
		}
		if value, ok := attrs["relative-location-fragment-start-indexes"]; ok {
			row.RelativeLocationFragmentStartIndexes = value
		}
		if value, ok := attrs["segment-movement-fragments"]; ok {
			row.SegmentMovementFragments = value
		}
		if value, ok := attrs["max-tracked-minutes"]; ok {
			row.MaxTrackedMinutes = value
		}
		if value, ok := attrs["tracked-minutes"]; ok {
			row.TrackedMinutes = value
		}
	}

	return row
}
