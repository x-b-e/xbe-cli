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

type doEquipmentMovementTripCustomerCostAllocationsUpdateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	ID         string
	IsExplicit bool
	Allocation string
}

func newDoEquipmentMovementTripCustomerCostAllocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer cost allocation",
		Long: `Update a customer cost allocation.

Optional flags:
  --is-explicit   Mark allocation as explicit
  --allocation    Allocation JSON string (details with customer_id and percentage)

Notes:
  Allocation details must reference customers from the trip's requirements and
  percentages must sum to 1.`,
		Example: `  # Mark allocation as explicit
  xbe do equipment-movement-trip-customer-cost-allocations update 123 --is-explicit true

  # Update allocation details
  xbe do equipment-movement-trip-customer-cost-allocations update 123 \
    --allocation '{\"details\":[{\"customer_id\":456,\"percentage\":\"1\"}]}'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementTripCustomerCostAllocationsUpdate,
	}
	initDoEquipmentMovementTripCustomerCostAllocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripCustomerCostAllocationsCmd.AddCommand(newDoEquipmentMovementTripCustomerCostAllocationsUpdateCmd())
}

func initDoEquipmentMovementTripCustomerCostAllocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-explicit", false, "Mark allocation as explicit")
	cmd.Flags().String("allocation", "", "Allocation JSON string")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripCustomerCostAllocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementTripCustomerCostAllocationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-explicit") {
		attributes["is-explicit"] = opts.IsExplicit
	}
	if opts.Allocation != "" {
		var allocation map[string]any
		if err := json.Unmarshal([]byte(opts.Allocation), &allocation); err != nil {
			return fmt.Errorf("invalid allocation JSON: %w", err)
		}
		attributes["allocation"] = allocation
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "equipment-movement-trip-customer-cost-allocations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-movement-trip-customer-cost-allocations/"+opts.ID, jsonBody)
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

	row := equipmentMovementTripCustomerCostAllocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment movement trip customer cost allocation %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementTripCustomerCostAllocationsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentMovementTripCustomerCostAllocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isExplicit, _ := cmd.Flags().GetBool("is-explicit")
	allocation, _ := cmd.Flags().GetString("allocation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripCustomerCostAllocationsUpdateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		ID:         args[0],
		IsExplicit: isExplicit,
		Allocation: allocation,
	}, nil
}
