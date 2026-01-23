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

type doEquipmentMovementTripsUpdateOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	ID                                         string
	JobNumber                                  string
	TrailerClassification                      string
	TrailerClassificationEquivalentIDs         []string
	ServiceTypeUnitOfMeasureIDs                []string
	ExplicitDriverDayMobilizationBeforeMinutes string
}

func newDoEquipmentMovementTripsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment movement trip",
		Long: `Update an equipment movement trip.

Provide the trip ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --job-number                           Job number
  --trailer-classification               Trailer classification ID (use empty string to clear)
  --trailer-classification-equivalent-ids Trailer classification equivalent IDs (comma-separated or repeated)
  --service-type-unit-of-measure-ids     Service type unit of measure IDs (comma-separated or repeated)
  --explicit-driver-day-mobilization-before-minutes Explicit mobilization before minutes`,
		Example: `  # Update job number
  xbe do equipment-movement-trips update 123 --job-number "EMT-200"

  # Update trailer classification equivalents
  xbe do equipment-movement-trips update 123 --trailer-classification-equivalent-ids 10,12`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementTripsUpdate,
	}
	initDoEquipmentMovementTripsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripsCmd.AddCommand(newDoEquipmentMovementTripsUpdateCmd())
}

func initDoEquipmentMovementTripsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-number", "", "Job number")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().StringSlice("trailer-classification-equivalent-ids", nil, "Trailer classification equivalent IDs (comma-separated or repeated)")
	cmd.Flags().StringSlice("service-type-unit-of-measure-ids", nil, "Service type unit of measure IDs (comma-separated or repeated)")
	cmd.Flags().String("explicit-driver-day-mobilization-before-minutes", "", "Explicit mobilization before minutes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementTripsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("job-number") {
		attributes["job-number"] = opts.JobNumber
	}
	if cmd.Flags().Changed("trailer-classification-equivalent-ids") {
		attributes["trailer-classification-equivalent-ids"] = opts.TrailerClassificationEquivalentIDs
	}
	if cmd.Flags().Changed("service-type-unit-of-measure-ids") {
		attributes["service-type-unit-of-measure-ids"] = opts.ServiceTypeUnitOfMeasureIDs
	}
	if cmd.Flags().Changed("explicit-driver-day-mobilization-before-minutes") {
		attributes["explicit-driver-day-mobilization-before-minutes"] = opts.ExplicitDriverDayMobilizationBeforeMinutes
	}

	if cmd.Flags().Changed("trailer-classification") {
		if opts.TrailerClassification == "" {
			relationships["trailer-classification"] = map[string]any{"data": nil}
		} else {
			relationships["trailer-classification"] = map[string]any{
				"data": map[string]any{
					"type": "trailer-classifications",
					"id":   opts.TrailerClassification,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "equipment-movement-trips",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-movement-trips/"+opts.ID, jsonBody)
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
		row := buildEquipmentMovementTripRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment movement trip %s\n", resp.Data.ID)
	return nil
}

func parseDoEquipmentMovementTripsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentMovementTripsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	trailerClassificationEquivalentIDs, _ := cmd.Flags().GetStringSlice("trailer-classification-equivalent-ids")
	serviceTypeUnitOfMeasureIDs, _ := cmd.Flags().GetStringSlice("service-type-unit-of-measure-ids")
	explicitDriverDayMobilizationBeforeMinutes, _ := cmd.Flags().GetString("explicit-driver-day-mobilization-before-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripsUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		JobNumber:                          jobNumber,
		TrailerClassification:              trailerClassification,
		TrailerClassificationEquivalentIDs: trailerClassificationEquivalentIDs,
		ServiceTypeUnitOfMeasureIDs:        serviceTypeUnitOfMeasureIDs,
		ExplicitDriverDayMobilizationBeforeMinutes: explicitDriverDayMobilizationBeforeMinutes,
	}, nil
}
