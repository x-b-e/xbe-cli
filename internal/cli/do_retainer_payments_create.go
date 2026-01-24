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

type doRetainerPaymentsCreateOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	Status         string
	Amount         string
	CreatedOn      string
	RetainerPeriod string
}

func newDoRetainerPaymentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a retainer payment",
		Long: `Create a retainer payment.

Required flags:
  --retainer-period  Retainer period ID
  --status           Payment status (editing, approved, batched, exported)
  --amount           Payment amount
  --created-on       Payment creation date (YYYY-MM-DD)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a retainer payment
  xbe do retainer-payments create \
    --retainer-period 123 \
    --status editing \
    --amount 1500.00 \
    --created-on 2025-01-17`,
		Args: cobra.NoArgs,
		RunE: runDoRetainerPaymentsCreate,
	}
	initDoRetainerPaymentsCreateFlags(cmd)
	return cmd
}

func init() {
	doRetainerPaymentsCmd.AddCommand(newDoRetainerPaymentsCreateCmd())
}

func initDoRetainerPaymentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("retainer-period", "", "Retainer period ID (required)")
	cmd.Flags().String("status", "", "Payment status (editing, approved, batched, exported)")
	cmd.Flags().String("amount", "", "Payment amount")
	cmd.Flags().String("created-on", "", "Payment creation date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRetainerPaymentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRetainerPaymentsCreateOptions(cmd)
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

	retainerPeriodID := strings.TrimSpace(opts.RetainerPeriod)
	if retainerPeriodID == "" {
		err := fmt.Errorf("--retainer-period is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Status) == "" {
		err := fmt.Errorf("--status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.Amount) == "" {
		err := fmt.Errorf("--amount is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.CreatedOn) == "" {
		err := fmt.Errorf("--created-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"status":     opts.Status,
		"amount":     opts.Amount,
		"created-on": opts.CreatedOn,
	}

	relationships := map[string]any{
		"retainer-period": map[string]any{
			"data": map[string]any{
				"type": "retainer-periods",
				"id":   retainerPeriodID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "retainer-payments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/retainer-payments", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created retainer payment %s\n", row.ID)
	return nil
}

func parseDoRetainerPaymentsCreateOptions(cmd *cobra.Command) (doRetainerPaymentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	retainerPeriod, _ := cmd.Flags().GetString("retainer-period")
	status, _ := cmd.Flags().GetString("status")
	amount, _ := cmd.Flags().GetString("amount")
	createdOn, _ := cmd.Flags().GetString("created-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainerPaymentsCreateOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		Status:         status,
		Amount:         amount,
		CreatedOn:      createdOn,
		RetainerPeriod: retainerPeriod,
	}, nil
}

func buildRetainerPaymentRowFromSingle(resp jsonAPISingleResponse) retainerPaymentRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := retainerPaymentRow{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		Amount:    stringAttr(attrs, "amount"),
		CreatedOn: formatDate(stringAttr(attrs, "created-on")),
		PayOn:     formatDate(stringAttr(attrs, "pay-on")),
		Kind:      stringAttr(attrs, "kind"),
	}

	row.RetainerID = relationshipIDFromMap(resource.Relationships, "retainer")
	row.RetainerPeriodID = relationshipIDFromMap(resource.Relationships, "retainer-period")

	return row
}
