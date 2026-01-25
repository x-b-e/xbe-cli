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

type doEquipmentMovementTripDispatchFulfillmentClerksCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	EquipmentMovementTripDispatch string
}

type equipmentMovementTripDispatchFulfillmentClerkRow struct {
	ID                              string `json:"id"`
	EquipmentMovementTripDispatchID string `json:"equipment_movement_trip_dispatch_id,omitempty"`
}

func newDoEquipmentMovementTripDispatchFulfillmentClerksCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement trip dispatch fulfillment clerk",
		Long: `Create an equipment movement trip dispatch fulfillment clerk.

Triggers the asynchronous fulfillment workflow for a specific equipment movement trip dispatch.

Required flags:
  --equipment-movement-trip-dispatch   Equipment movement trip dispatch ID (required)`,
		Example: `  # Run fulfillment for a dispatch
  xbe do equipment-movement-trip-dispatch-fulfillment-clerks create --equipment-movement-trip-dispatch 123`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementTripDispatchFulfillmentClerksCreate,
	}
	initDoEquipmentMovementTripDispatchFulfillmentClerksCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripDispatchFulfillmentClerksCmd.AddCommand(newDoEquipmentMovementTripDispatchFulfillmentClerksCreateCmd())
}

func initDoEquipmentMovementTripDispatchFulfillmentClerksCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment-movement-trip-dispatch", "", "Equipment movement trip dispatch ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripDispatchFulfillmentClerksCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementTripDispatchFulfillmentClerksCreateOptions(cmd)
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

	if opts.EquipmentMovementTripDispatch == "" {
		err := fmt.Errorf("--equipment-movement-trip-dispatch is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"equipment-movement-trip-dispatch": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-trip-dispatches",
				"id":   opts.EquipmentMovementTripDispatch,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-trip-dispatch-fulfillment-clerks",
			"attributes":    map[string]any{},
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-trip-dispatch-fulfillment-clerks", jsonBody)
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

	row := buildEquipmentMovementTripDispatchFulfillmentClerkRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement trip dispatch fulfillment clerk %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementTripDispatchFulfillmentClerksCreateOptions(cmd *cobra.Command) (doEquipmentMovementTripDispatchFulfillmentClerksCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentMovementTripDispatch, _ := cmd.Flags().GetString("equipment-movement-trip-dispatch")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripDispatchFulfillmentClerksCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		EquipmentMovementTripDispatch: equipmentMovementTripDispatch,
	}, nil
}

func buildEquipmentMovementTripDispatchFulfillmentClerkRowFromSingle(resp jsonAPISingleResponse) equipmentMovementTripDispatchFulfillmentClerkRow {
	row := equipmentMovementTripDispatchFulfillmentClerkRow{ID: resp.Data.ID}

	if rel, ok := resp.Data.Relationships["equipment-movement-trip-dispatch"]; ok && rel.Data != nil {
		row.EquipmentMovementTripDispatchID = rel.Data.ID
	}

	return row
}
