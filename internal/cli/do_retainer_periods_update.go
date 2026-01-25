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

type doRetainerPeriodsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	Retainer            string
	StartOn             string
	EndOn               string
	WeeklyPaymentAmount string
}

func newDoRetainerPeriodsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a retainer period",
		Long: `Update a retainer period.

Optional flags:
  --start-on               Start date (YYYY-MM-DD)
  --end-on                 End date (YYYY-MM-DD)
  --weekly-payment-amount  Weekly payment amount
  --retainer               Retainer ID`,
		Example: `  # Update weekly payment amount
  xbe do retainer-periods update 123 --weekly-payment-amount 1250

  # Update start/end dates
  xbe do retainer-periods update 123 --start-on 2026-02-01 --end-on 2026-02-28

  # Update retainer relationship
  xbe do retainer-periods update 123 --retainer 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRetainerPeriodsUpdate,
	}
	initDoRetainerPeriodsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRetainerPeriodsCmd.AddCommand(newDoRetainerPeriodsUpdateCmd())
}

func initDoRetainerPeriodsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD)")
	cmd.Flags().String("weekly-payment-amount", "", "Weekly payment amount")
	cmd.Flags().String("retainer", "", "Retainer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRetainerPeriodsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRetainerPeriodsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("start-on") {
		attributes["start-on"] = opts.StartOn
	}
	if cmd.Flags().Changed("end-on") {
		attributes["end-on"] = opts.EndOn
	}
	if cmd.Flags().Changed("weekly-payment-amount") {
		attributes["weekly-payment-amount"] = opts.WeeklyPaymentAmount
	}
	if cmd.Flags().Changed("retainer") {
		relationships["retainer"] = map[string]any{
			"data": map[string]any{
				"type": "retainers",
				"id":   opts.Retainer,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "retainer-periods",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/retainer-periods/"+opts.ID, jsonBody)
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
		row := buildRetainerPeriodRow(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated retainer period %s\n", resp.Data.ID)
	return nil
}

func parseDoRetainerPeriodsUpdateOptions(cmd *cobra.Command, args []string) (doRetainerPeriodsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	weeklyPaymentAmount, _ := cmd.Flags().GetString("weekly-payment-amount")
	retainer, _ := cmd.Flags().GetString("retainer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainerPeriodsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		Retainer:            retainer,
		StartOn:             startOn,
		EndOn:               endOn,
		WeeklyPaymentAmount: weeklyPaymentAmount,
	}, nil
}
