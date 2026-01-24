package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type retainerPaymentDeductionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type retainerPaymentDeductionDetails struct {
	ID                  string `json:"id"`
	RetainerPaymentID   string `json:"retainer_payment_id,omitempty"`
	RetainerDeductionID string `json:"retainer_deduction_id,omitempty"`
	AppliedAmount       string `json:"applied_amount,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

func newRetainerPaymentDeductionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show retainer payment deduction details",
		Long: `Show full details of a retainer payment deduction.

Output Fields:
  ID         Retainer payment deduction identifier
  PAYMENT    Retainer payment ID
  DEDUCTION  Retainer deduction ID
  APPLIED    Applied amount
  CREATED    When the deduction was created
  UPDATED    When the deduction was last updated

Arguments:
  <id>  Retainer payment deduction ID (required).`,
		Example: `  # Show a retainer payment deduction
  xbe view retainer-payment-deductions show 123

  # JSON output
  xbe view retainer-payment-deductions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRetainerPaymentDeductionsShow,
	}
	initRetainerPaymentDeductionsShowFlags(cmd)
	return cmd
}

func init() {
	retainerPaymentDeductionsCmd.AddCommand(newRetainerPaymentDeductionsShowCmd())
}

func initRetainerPaymentDeductionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerPaymentDeductionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRetainerPaymentDeductionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("retainer payment deduction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[retainer-payment-deductions]", "retainer-payment,retainer-deduction,applied-amount,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-payment-deductions/"+id, query)
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

	details := buildRetainerPaymentDeductionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRetainerPaymentDeductionDetails(cmd, details)
}

func parseRetainerPaymentDeductionsShowOptions(cmd *cobra.Command) (retainerPaymentDeductionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerPaymentDeductionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRetainerPaymentDeductionDetails(resp jsonAPISingleResponse) retainerPaymentDeductionDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := retainerPaymentDeductionDetails{
		ID:            resource.ID,
		AppliedAmount: stringAttr(attrs, "applied-amount"),
		CreatedAt:     formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:     formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["retainer-payment"]; ok && rel.Data != nil {
		details.RetainerPaymentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["retainer-deduction"]; ok && rel.Data != nil {
		details.RetainerDeductionID = rel.Data.ID
	}

	return details
}

func renderRetainerPaymentDeductionDetails(cmd *cobra.Command, details retainerPaymentDeductionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RetainerPaymentID != "" {
		fmt.Fprintf(out, "Retainer Payment: %s\n", details.RetainerPaymentID)
	}
	if details.RetainerDeductionID != "" {
		fmt.Fprintf(out, "Retainer Deduction: %s\n", details.RetainerDeductionID)
	}
	if details.AppliedAmount != "" {
		fmt.Fprintf(out, "Applied Amount: %s\n", details.AppliedAmount)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
