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

type doEquipmentMovementTripsCreateOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	Broker                                     string
	JobNumber                                  string
	TrailerClassification                      string
	TrailerClassificationEquivalentIDs         []string
	ServiceTypeUnitOfMeasureIDs                []string
	ExplicitDriverDayMobilizationBeforeMinutes string
}

func newDoEquipmentMovementTripsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement trip",
		Long: `Create an equipment movement trip.

Required:
  --broker                               Broker ID

Optional:
  --job-number                           Job number
  --trailer-classification               Trailer classification ID
  --trailer-classification-equivalent-ids Trailer classification equivalent IDs (comma-separated or repeated)
  --service-type-unit-of-measure-ids     Service type unit of measure IDs (comma-separated or repeated)
  --explicit-driver-day-mobilization-before-minutes Explicit mobilization before minutes`,
		Example: `  # Create a trip with a broker
  xbe do equipment-movement-trips create --broker 123

  # Create with job number and trailer classification
  xbe do equipment-movement-trips create --broker 123 --job-number "EMT-100" \
    --trailer-classification 45`,
		RunE: runDoEquipmentMovementTripsCreate,
	}
	initDoEquipmentMovementTripsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripsCmd.AddCommand(newDoEquipmentMovementTripsCreateCmd())
}

func initDoEquipmentMovementTripsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("job-number", "", "Job number")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().StringSlice("trailer-classification-equivalent-ids", nil, "Trailer classification equivalent IDs (comma-separated or repeated)")
	cmd.Flags().StringSlice("service-type-unit-of-measure-ids", nil, "Service type unit of measure IDs (comma-separated or repeated)")
	cmd.Flags().String("explicit-driver-day-mobilization-before-minutes", "", "Explicit mobilization before minutes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("broker")
}

func runDoEquipmentMovementTripsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementTripsCreateOptions(cmd)
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
	if opts.JobNumber != "" {
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

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	if opts.TrailerClassification != "" {
		relationships["trailer-classification"] = map[string]any{
			"data": map[string]any{
				"type": "trailer-classifications",
				"id":   opts.TrailerClassification,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-trips",
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

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-trips", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement trip %s\n", resp.Data.ID)
	return nil
}

func parseDoEquipmentMovementTripsCreateOptions(cmd *cobra.Command) (doEquipmentMovementTripsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	broker, _ := cmd.Flags().GetString("broker")
	jobNumber, _ := cmd.Flags().GetString("job-number")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	trailerClassificationEquivalentIDs, _ := cmd.Flags().GetStringSlice("trailer-classification-equivalent-ids")
	serviceTypeUnitOfMeasureIDs, _ := cmd.Flags().GetStringSlice("service-type-unit-of-measure-ids")
	explicitDriverDayMobilizationBeforeMinutes, _ := cmd.Flags().GetString("explicit-driver-day-mobilization-before-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Broker:                             broker,
		JobNumber:                          jobNumber,
		TrailerClassification:              trailerClassification,
		TrailerClassificationEquivalentIDs: trailerClassificationEquivalentIDs,
		ServiceTypeUnitOfMeasureIDs:        serviceTypeUnitOfMeasureIDs,
		ExplicitDriverDayMobilizationBeforeMinutes: explicitDriverDayMobilizationBeforeMinutes,
	}, nil
}
