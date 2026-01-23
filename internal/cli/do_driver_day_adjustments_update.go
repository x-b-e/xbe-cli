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

type doDriverDayAdjustmentsUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	AmountExplicit string
}

func newDoDriverDayAdjustmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a driver day adjustment",
		Long: `Update a driver day adjustment.

Updatable fields:
  --amount-explicit  Explicit adjustment amount`,
		Example: `  # Update explicit amount
  xbe do driver-day-adjustments update 123 --amount-explicit "15.00"

  # Clear explicit amount (if supported by API)
  xbe do driver-day-adjustments update 123 --amount-explicit ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDriverDayAdjustmentsUpdate,
	}
	initDoDriverDayAdjustmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doDriverDayAdjustmentsCmd.AddCommand(newDoDriverDayAdjustmentsUpdateCmd())
}

func initDoDriverDayAdjustmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("amount-explicit", "", "Explicit adjustment amount")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDayAdjustmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverDayAdjustmentsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("amount-explicit") {
		attributes["amount-explicit"] = opts.AmountExplicit
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "driver-day-adjustments",
		"id":         opts.ID,
		"attributes": attributes,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	path := fmt.Sprintf("/v1/driver-day-adjustments/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	row := buildDriverDayAdjustmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated driver day adjustment %s\n", row.ID)
	return nil
}

func parseDoDriverDayAdjustmentsUpdateOptions(cmd *cobra.Command, args []string) (doDriverDayAdjustmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	amountExplicit, _ := cmd.Flags().GetString("amount-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverDayAdjustmentsUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		AmountExplicit: amountExplicit,
	}, nil
}
