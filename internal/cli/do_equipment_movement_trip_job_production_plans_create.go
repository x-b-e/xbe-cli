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

type doEquipmentMovementTripJobProductionPlansCreateOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	EquipmentMovementTrip string
}

func newDoEquipmentMovementTripJobProductionPlansCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement trip job production plan link",
		Long: `Create an equipment movement trip job production plan link.

Required flags:
  --equipment-movement-trip   Equipment movement trip ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Link a trip to a job production plan
  xbe do equipment-movement-trip-job-production-plans create --equipment-movement-trip 123`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementTripJobProductionPlansCreate,
	}
	initDoEquipmentMovementTripJobProductionPlansCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripJobProductionPlansCmd.AddCommand(newDoEquipmentMovementTripJobProductionPlansCreateCmd())
}

func initDoEquipmentMovementTripJobProductionPlansCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment-movement-trip", "", "Equipment movement trip ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripJobProductionPlansCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementTripJobProductionPlansCreateOptions(cmd)
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

	if strings.TrimSpace(opts.EquipmentMovementTrip) == "" {
		err := fmt.Errorf("--equipment-movement-trip is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"equipment-movement-trip": map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-trips",
				"id":   opts.EquipmentMovementTrip,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-trip-job-production-plans",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-trip-job-production-plans", jsonBody)
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
		row := buildEquipmentMovementTripJobProductionPlanRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})
		if len(row) > 0 {
			return writeJSON(cmd.OutOrStdout(), row[0])
		}
		return writeJSON(cmd.OutOrStdout(), map[string]any{"id": resp.Data.ID})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement trip job production plan %s\n", resp.Data.ID)
	return nil
}

func parseDoEquipmentMovementTripJobProductionPlansCreateOptions(cmd *cobra.Command) (doEquipmentMovementTripJobProductionPlansCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentMovementTrip, _ := cmd.Flags().GetString("equipment-movement-trip")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripJobProductionPlansCreateOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		EquipmentMovementTrip: equipmentMovementTrip,
	}, nil
}
