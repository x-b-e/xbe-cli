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

type doDriverDayTripsAdjustmentsUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	Description        string
	Status             string
	NewTripsAttributes string
}

func newDoDriverDayTripsAdjustmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a driver day trips adjustment",
		Long: `Update a driver day trips adjustment.

Updatable fields:
  --description           Update the description
  --status                Update the status (e.g., editing)
  --new-trips-attributes  JSON array of updated trip attributes

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update an adjustment
  xbe do driver-day-trips-adjustments update 123 \
    --description \"Updated trips\" \
    --new-trips-attributes '[{\"note\":\"Updated trip\"}]'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDriverDayTripsAdjustmentsUpdate,
	}
	initDoDriverDayTripsAdjustmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayTripsAdjustmentsCmd.AddCommand(newDoDriverDayTripsAdjustmentsUpdateCmd())
}

func initDoDriverDayTripsAdjustmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("description", "", "Description of the adjustment")
	cmd.Flags().String("status", "", "Adjustment status (e.g., editing)")
	cmd.Flags().String("new-trips-attributes", "", "JSON array of updated trip attributes")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayTripsAdjustmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverDayTripsAdjustmentsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("new-trips-attributes") {
		if strings.TrimSpace(opts.NewTripsAttributes) == "" {
			return fmt.Errorf("--new-trips-attributes cannot be empty when provided")
		}
		var newTrips any
		if err := json.Unmarshal([]byte(opts.NewTripsAttributes), &newTrips); err != nil {
			return fmt.Errorf("invalid new-trips-attributes JSON: %w", err)
		}
		if _, ok := newTrips.([]any); !ok {
			return fmt.Errorf("--new-trips-attributes must be a JSON array")
		}
		attributes["new-trips-attributes"] = newTrips
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "driver-day-trips-adjustments",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/driver-day-trips-adjustments/"+opts.ID, jsonBody)
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

	row := driverDayTripsAdjustmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated driver day trips adjustment %s\n", row.ID)
	return nil
}

func parseDoDriverDayTripsAdjustmentsUpdateOptions(cmd *cobra.Command, args []string) (doDriverDayTripsAdjustmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	description, _ := cmd.Flags().GetString("description")
	status, _ := cmd.Flags().GetString("status")
	newTripsAttributes, _ := cmd.Flags().GetString("new-trips-attributes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayTripsAdjustmentsUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		Description:        description,
		Status:             status,
		NewTripsAttributes: newTripsAttributes,
	}, nil
}
