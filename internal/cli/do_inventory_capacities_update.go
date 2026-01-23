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

type doInventoryCapacitiesUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	MaxCapacityTons string
	MinCapacityTons string
	ThresholdTons   string
}

func newDoInventoryCapacitiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an inventory capacity",
		Long: `Update an inventory capacity.

All flags are optional. Only provided flags will be updated.

Optional flags:
  --max-capacity-tons  Maximum capacity in tons
  --min-capacity-tons  Minimum capacity in tons
  --threshold-tons     Alert threshold in tons`,
		Example: `  # Update max capacity
  xbe do inventory-capacities update 123 --max-capacity-tons 750

  # Update threshold
  xbe do inventory-capacities update 123 --threshold-tons 120`,
		Args: cobra.ExactArgs(1),
		RunE: runDoInventoryCapacitiesUpdate,
	}
	initDoInventoryCapacitiesUpdateFlags(cmd)
	return cmd
}

func init() {
	doInventoryCapacitiesCmd.AddCommand(newDoInventoryCapacitiesUpdateCmd())
}

func initDoInventoryCapacitiesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("max-capacity-tons", "", "Maximum capacity in tons")
	cmd.Flags().String("min-capacity-tons", "", "Minimum capacity in tons")
	cmd.Flags().String("threshold-tons", "", "Alert threshold in tons")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoInventoryCapacitiesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoInventoryCapacitiesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("max-capacity-tons") {
		attributes["max-capacity-tons"] = opts.MaxCapacityTons
	}
	if cmd.Flags().Changed("min-capacity-tons") {
		attributes["min-capacity-tons"] = opts.MinCapacityTons
	}
	if cmd.Flags().Changed("threshold-tons") {
		attributes["threshold-tons"] = opts.ThresholdTons
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "inventory-capacities",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/inventory-capacities/"+opts.ID, jsonBody)
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

	result := buildInventoryCapacityResult(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), result)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated inventory capacity %s\n", result["id"])
	return nil
}

func parseDoInventoryCapacitiesUpdateOptions(cmd *cobra.Command, args []string) (doInventoryCapacitiesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	maxCapacityTons, _ := cmd.Flags().GetString("max-capacity-tons")
	minCapacityTons, _ := cmd.Flags().GetString("min-capacity-tons")
	thresholdTons, _ := cmd.Flags().GetString("threshold-tons")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInventoryCapacitiesUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		MaxCapacityTons: maxCapacityTons,
		MinCapacityTons: minCapacityTons,
		ThresholdTons:   thresholdTons,
	}, nil
}
