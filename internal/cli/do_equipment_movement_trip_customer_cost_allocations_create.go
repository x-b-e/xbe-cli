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

type doEquipmentMovementTripCustomerCostAllocationsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	TripID     string
	IsExplicit bool
	Allocation string
}

func newDoEquipmentMovementTripCustomerCostAllocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer cost allocation",
		Long: `Create a customer cost allocation for an equipment movement trip.

Required flags:
  --trip    Equipment movement trip ID

Optional flags:
  --is-explicit   Mark allocation as explicit
  --allocation    Allocation JSON string (details with customer_id and percentage)

Notes:
  Allocation details must reference customers from the trip's requirements and
  percentages must sum to 1.`,
		Example: `  # Create an allocation using trip defaults
  xbe do equipment-movement-trip-customer-cost-allocations create --trip 123

  # Create an explicit allocation
  xbe do equipment-movement-trip-customer-cost-allocations create \
    --trip 123 \
    --is-explicit \
    --allocation '{\"details\":[{\"customer_id\":456,\"percentage\":\"1\"}]}'`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementTripCustomerCostAllocationsCreate,
	}
	initDoEquipmentMovementTripCustomerCostAllocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripCustomerCostAllocationsCmd.AddCommand(newDoEquipmentMovementTripCustomerCostAllocationsCreateCmd())
}

func initDoEquipmentMovementTripCustomerCostAllocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trip", "", "Equipment movement trip ID (required)")
	cmd.Flags().Bool("is-explicit", false, "Mark allocation as explicit")
	cmd.Flags().String("allocation", "", "Allocation JSON string")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripCustomerCostAllocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementTripCustomerCostAllocationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.TripID == "" {
		err := fmt.Errorf("--trip is required")
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

	relationships := map[string]any{
		"trip": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-trips",
				"id":   opts.TripID,
			},
		},
	}

	data := map[string]any{
		"type": "equipment-movement-trip-customer-cost-allocations",
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	data["relationships"] = relationships

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-trip-customer-cost-allocations", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement trip customer cost allocation %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementTripCustomerCostAllocationsCreateOptions(cmd *cobra.Command) (doEquipmentMovementTripCustomerCostAllocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tripID, _ := cmd.Flags().GetString("trip")
	isExplicit, _ := cmd.Flags().GetBool("is-explicit")
	allocation, _ := cmd.Flags().GetString("allocation")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripCustomerCostAllocationsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		TripID:     tripID,
		IsExplicit: isExplicit,
		Allocation: allocation,
	}, nil
}
