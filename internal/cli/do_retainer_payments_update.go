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

type doRetainerPaymentsUpdateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	ID             string
	Status         string
	Amount         string
	CreatedOn      string
	RetainerPeriod string
}

func newDoRetainerPaymentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a retainer payment",
		Long: `Update a retainer payment.

Optional flags:
  --status           Payment status (editing, approved, batched, exported)
  --amount           Payment amount
  --created-on       Payment creation date (YYYY-MM-DD)
  --retainer-period  Retainer period ID

Note: Amount, created-on, and retainer period can only be changed while the
payment status is editing.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update retainer payment status
  xbe do retainer-payments update 123 --status approved

  # Update amount and created date (editing status required)
  xbe do retainer-payments update 123 --amount 1400.00 --created-on 2025-01-24`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRetainerPaymentsUpdate,
	}
	initDoRetainerPaymentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doRetainerPaymentsCmd.AddCommand(newDoRetainerPaymentsUpdateCmd())
}

func initDoRetainerPaymentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Payment status (editing, approved, batched, exported)")
	cmd.Flags().String("amount", "", "Payment amount")
	cmd.Flags().String("created-on", "", "Payment creation date (YYYY-MM-DD)")
	cmd.Flags().String("retainer-period", "", "Retainer period ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRetainerPaymentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRetainerPaymentsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("amount") {
		attributes["amount"] = opts.Amount
	}
	if cmd.Flags().Changed("created-on") {
		attributes["created-on"] = opts.CreatedOn
	}
	if cmd.Flags().Changed("retainer-period") {
		retainerPeriodID := strings.TrimSpace(opts.RetainerPeriod)
		if retainerPeriodID == "" {
			err := fmt.Errorf("--retainer-period cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["retainer-period"] = map[string]any{
			"data": map[string]any{
				"type": "retainer-periods",
				"id":   retainerPeriodID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "retainer-payments",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/retainer-payments/"+opts.ID, jsonBody)
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

	row := buildRetainerPaymentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated retainer payment %s\n", row.ID)
	return nil
}

func parseDoRetainerPaymentsUpdateOptions(cmd *cobra.Command, args []string) (doRetainerPaymentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	amount, _ := cmd.Flags().GetString("amount")
	createdOn, _ := cmd.Flags().GetString("created-on")
	retainerPeriod, _ := cmd.Flags().GetString("retainer-period")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainerPaymentsUpdateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		ID:             args[0],
		Status:         status,
		Amount:         amount,
		CreatedOn:      createdOn,
		RetainerPeriod: retainerPeriod,
	}, nil
}
