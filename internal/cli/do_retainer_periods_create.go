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

type doRetainerPeriodsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Retainer            string
	StartOn             string
	EndOn               string
	WeeklyPaymentAmount string
}

func newDoRetainerPeriodsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a retainer period",
		Long: `Create a retainer period.

Required flags:
  --retainer               Retainer ID
  --start-on               Start date (YYYY-MM-DD)
  --end-on                 End date (YYYY-MM-DD)
  --weekly-payment-amount  Weekly payment amount`,
		Example: `  # Create a retainer period
  xbe do retainer-periods create --retainer 123 --start-on 2026-01-01 --end-on 2026-01-31 --weekly-payment-amount 1000

  # Output JSON
  xbe do retainer-periods create --retainer 123 --start-on 2026-01-01 --end-on 2026-01-31 --weekly-payment-amount 1000 --json`,
		Args: cobra.NoArgs,
		RunE: runDoRetainerPeriodsCreate,
	}
	initDoRetainerPeriodsCreateFlags(cmd)
	return cmd
}

func init() {
	doRetainerPeriodsCmd.AddCommand(newDoRetainerPeriodsCreateCmd())
}

func initDoRetainerPeriodsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("retainer", "", "Retainer ID (required)")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD) (required)")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD) (required)")
	cmd.Flags().String("weekly-payment-amount", "", "Weekly payment amount (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("retainer")
	_ = cmd.MarkFlagRequired("start-on")
	_ = cmd.MarkFlagRequired("end-on")
	_ = cmd.MarkFlagRequired("weekly-payment-amount")
}

func runDoRetainerPeriodsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRetainerPeriodsCreateOptions(cmd)
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

	attributes := map[string]any{
		"start-on":              opts.StartOn,
		"end-on":                opts.EndOn,
		"weekly-payment-amount": opts.WeeklyPaymentAmount,
	}

	relationships := map[string]any{
		"retainer": map[string]any{
			"data": map[string]any{
				"type": "retainers",
				"id":   opts.Retainer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "retainer-periods",
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

	body, _, err := client.Post(cmd.Context(), "/v1/retainer-periods", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created retainer period %s\n", resp.Data.ID)
	return nil
}

func parseDoRetainerPeriodsCreateOptions(cmd *cobra.Command) (doRetainerPeriodsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	retainer, _ := cmd.Flags().GetString("retainer")
	startOn, _ := cmd.Flags().GetString("start-on")
	endOn, _ := cmd.Flags().GetString("end-on")
	weeklyPaymentAmount, _ := cmd.Flags().GetString("weekly-payment-amount")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainerPeriodsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Retainer:            retainer,
		StartOn:             startOn,
		EndOn:               endOn,
		WeeklyPaymentAmount: weeklyPaymentAmount,
	}, nil
}
