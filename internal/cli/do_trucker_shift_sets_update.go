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

type doTruckerShiftSetsUpdateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	ID                                string
	ExplicitMobilizationBeforeMinutes int
	ExplicitPreTripMinutes            int
	ExplicitPostTripMinutes           int
	IsCustomerAmountConstraintEnabled bool
	IsBrokerAmountConstraintEnabled   bool
	IsTimeSheetEnabled                bool
	OdometerStartValue                float64
	OdometerEndValue                  float64
	OdometerUnitOfMeasureExplicit     string
	NewShiftIDs                       string
	Trips                             string
	ExplicitBrokerAmountConstraint    string
}

func newDoTruckerShiftSetsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trucker shift set",
		Long: `Update a trucker shift set (driver day).

Common flags:
  --new-shift-ids                     New shift IDs to add (comma-separated)
  --explicit-mobilization-before-minutes  Explicit mobilization before minutes
  --explicit-pre-trip-minutes         Explicit pre-trip minutes
  --explicit-post-trip-minutes        Explicit post-trip minutes
  --is-customer-amount-constraint-enabled  Enable customer amount constraint
  --is-broker-amount-constraint-enabled    Enable broker amount constraint
  --is-time-sheet-enabled             Enable time sheets
  --odometer-start-value              Odometer start value
  --odometer-end-value                Odometer end value
  --odometer-unit-of-measure-explicit Odometer unit of measure (mile|kilometer)
  --explicit-broker-amount-constraint Explicit broker amount constraint ID
  --trips                             Trip IDs to set (comma-separated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Add a shift to a trucker shift set
  xbe do trucker-shift-sets update 123 --new-shift-ids 456

  # Update odometer values
  xbe do trucker-shift-sets update 123 --odometer-start-value 1200 --odometer-end-value 1250

  # Enable time sheets
  xbe do trucker-shift-sets update 123 --is-time-sheet-enabled

  # JSON output
  xbe do trucker-shift-sets update 123 --is-time-sheet-enabled --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerShiftSetsUpdate,
	}
	initDoTruckerShiftSetsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerShiftSetsCmd.AddCommand(newDoTruckerShiftSetsUpdateCmd())
}

func initDoTruckerShiftSetsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Int("explicit-mobilization-before-minutes", 0, "Explicit mobilization before minutes")
	cmd.Flags().Int("explicit-pre-trip-minutes", 0, "Explicit pre-trip minutes")
	cmd.Flags().Int("explicit-post-trip-minutes", 0, "Explicit post-trip minutes")
	cmd.Flags().Bool("is-customer-amount-constraint-enabled", false, "Enable customer amount constraint")
	cmd.Flags().Bool("is-broker-amount-constraint-enabled", false, "Enable broker amount constraint")
	cmd.Flags().Bool("is-time-sheet-enabled", false, "Enable time sheets")
	cmd.Flags().Float64("odometer-start-value", 0, "Odometer start value")
	cmd.Flags().Float64("odometer-end-value", 0, "Odometer end value")
	cmd.Flags().String("odometer-unit-of-measure-explicit", "", "Odometer unit of measure (mile|kilometer)")
	cmd.Flags().String("new-shift-ids", "", "New shift IDs to add (comma-separated)")
	cmd.Flags().String("trips", "", "Trip IDs to set (comma-separated)")
	cmd.Flags().String("explicit-broker-amount-constraint", "", "Explicit broker amount constraint ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerShiftSetsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerShiftSetsUpdateOptions(cmd, args)
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
	hasChanges := false

	if cmd.Flags().Changed("explicit-mobilization-before-minutes") {
		attributes["explicit-mobilization-before-minutes"] = opts.ExplicitMobilizationBeforeMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-pre-trip-minutes") {
		attributes["explicit-pre-trip-minutes"] = opts.ExplicitPreTripMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-post-trip-minutes") {
		attributes["explicit-post-trip-minutes"] = opts.ExplicitPostTripMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("is-customer-amount-constraint-enabled") {
		attributes["is-customer-amount-constraint-enabled"] = opts.IsCustomerAmountConstraintEnabled
		hasChanges = true
	}
	if cmd.Flags().Changed("is-broker-amount-constraint-enabled") {
		attributes["is-broker-amount-constraint-enabled"] = opts.IsBrokerAmountConstraintEnabled
		hasChanges = true
	}
	if cmd.Flags().Changed("is-time-sheet-enabled") {
		attributes["is-time-sheet-enabled"] = opts.IsTimeSheetEnabled
		hasChanges = true
	}
	if cmd.Flags().Changed("odometer-start-value") {
		attributes["odometer-start-value"] = opts.OdometerStartValue
		hasChanges = true
	}
	if cmd.Flags().Changed("odometer-end-value") {
		attributes["odometer-end-value"] = opts.OdometerEndValue
		hasChanges = true
	}
	if cmd.Flags().Changed("odometer-unit-of-measure-explicit") {
		if strings.TrimSpace(opts.OdometerUnitOfMeasureExplicit) == "" {
			attributes["odometer-unit-of-measure-explicit"] = nil
		} else {
			attributes["odometer-unit-of-measure-explicit"] = opts.OdometerUnitOfMeasureExplicit
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("new-shift-ids") {
		attributes["new-shift-ids"] = splitCommaList(opts.NewShiftIDs)
		hasChanges = true
	}

	if cmd.Flags().Changed("explicit-broker-amount-constraint") {
		if strings.TrimSpace(opts.ExplicitBrokerAmountConstraint) == "" {
			relationships["explicit-broker-amount-constraint"] = map[string]any{"data": nil}
		} else {
			relationships["explicit-broker-amount-constraint"] = map[string]any{
				"data": map[string]any{
					"type": "shift-set-time-card-constraints",
					"id":   opts.ExplicitBrokerAmountConstraint,
				},
			}
		}
		hasChanges = true
	}

	setToManyRelationship := func(flagName, key, resourceType, raw string) {
		if !cmd.Flags().Changed(flagName) {
			return
		}
		if strings.TrimSpace(raw) == "" {
			relationships[key] = map[string]any{"data": []any{}}
			hasChanges = true
			return
		}
		ids := splitCommaList(raw)
		data := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			data = append(data, map[string]any{
				"type": resourceType,
				"id":   id,
			})
		}
		relationships[key] = map[string]any{"data": data}
		hasChanges = true
	}

	setToManyRelationship("trips", "trips", "trips", opts.Trips)

	if !hasChanges {
		return fmt.Errorf("no fields to update")
	}

	data := map[string]any{
		"id":   opts.ID,
		"type": "trucker-shift-sets",
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-shift-sets/"+opts.ID, jsonBody)
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

	details := buildTruckerShiftSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker shift set %s\n", details.ID)
	return nil
}

func parseDoTruckerShiftSetsUpdateOptions(cmd *cobra.Command, args []string) (doTruckerShiftSetsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	explicitMobilizationBeforeMinutes, _ := cmd.Flags().GetInt("explicit-mobilization-before-minutes")
	explicitPreTripMinutes, _ := cmd.Flags().GetInt("explicit-pre-trip-minutes")
	explicitPostTripMinutes, _ := cmd.Flags().GetInt("explicit-post-trip-minutes")
	isCustomerAmountConstraintEnabled, _ := cmd.Flags().GetBool("is-customer-amount-constraint-enabled")
	isBrokerAmountConstraintEnabled, _ := cmd.Flags().GetBool("is-broker-amount-constraint-enabled")
	isTimeSheetEnabled, _ := cmd.Flags().GetBool("is-time-sheet-enabled")
	odometerStartValue, _ := cmd.Flags().GetFloat64("odometer-start-value")
	odometerEndValue, _ := cmd.Flags().GetFloat64("odometer-end-value")
	odometerUnitOfMeasureExplicit, _ := cmd.Flags().GetString("odometer-unit-of-measure-explicit")
	newShiftIDs, _ := cmd.Flags().GetString("new-shift-ids")
	trips, _ := cmd.Flags().GetString("trips")
	explicitBrokerAmountConstraint, _ := cmd.Flags().GetString("explicit-broker-amount-constraint")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerShiftSetsUpdateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		ID:                                args[0],
		ExplicitMobilizationBeforeMinutes: explicitMobilizationBeforeMinutes,
		ExplicitPreTripMinutes:            explicitPreTripMinutes,
		ExplicitPostTripMinutes:           explicitPostTripMinutes,
		IsCustomerAmountConstraintEnabled: isCustomerAmountConstraintEnabled,
		IsBrokerAmountConstraintEnabled:   isBrokerAmountConstraintEnabled,
		IsTimeSheetEnabled:                isTimeSheetEnabled,
		OdometerStartValue:                odometerStartValue,
		OdometerEndValue:                  odometerEndValue,
		OdometerUnitOfMeasureExplicit:     odometerUnitOfMeasureExplicit,
		NewShiftIDs:                       newShiftIDs,
		Trips:                             trips,
		ExplicitBrokerAmountConstraint:    explicitBrokerAmountConstraint,
	}, nil
}
