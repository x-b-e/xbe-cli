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

type retainerDeductionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type retainerDeductionDetails struct {
	ID                          string   `json:"id"`
	Amount                      any      `json:"amount,omitempty"`
	Note                        string   `json:"note,omitempty"`
	RetainerID                  string   `json:"retainer_id,omitempty"`
	RetainerPaymentDeductionIDs []string `json:"retainer_payment_deduction_ids,omitempty"`
}

func newRetainerDeductionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show retainer deduction details",
		Long: `Show the full details of a retainer deduction.

Output Fields:
  ID         Retainer deduction identifier
  AMOUNT     Deduction amount
  NOTE       Deduction note
  RETAINER   Retainer ID
  PAYMENTS   Retainer payment deduction IDs

Arguments:
  <id>  Retainer deduction ID (required).`,
		Example: `  # Show a retainer deduction
  xbe view retainer-deductions show 123

  # Output as JSON
  xbe view retainer-deductions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRetainerDeductionsShow,
	}
	initRetainerDeductionsShowFlags(cmd)
	return cmd
}

func init() {
	retainerDeductionsCmd.AddCommand(newRetainerDeductionsShowCmd())
}

func initRetainerDeductionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerDeductionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseRetainerDeductionsShowOptions(cmd)
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
		return fmt.Errorf("retainer deduction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[retainer-deductions]", "amount,note,retainer,retainer-payment-deductions")

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-deductions/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildRetainerDeductionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRetainerDeductionDetails(cmd, details)
}

func parseRetainerDeductionsShowOptions(cmd *cobra.Command) (retainerDeductionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerDeductionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRetainerDeductionDetails(resp jsonAPISingleResponse) retainerDeductionDetails {
	attrs := resp.Data.Attributes

	details := retainerDeductionDetails{
		ID:     resp.Data.ID,
		Amount: attrs["amount"],
		Note:   stringAttr(attrs, "note"),
	}

	if rel, ok := resp.Data.Relationships["retainer"]; ok && rel.Data != nil {
		details.RetainerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["retainer-payment-deductions"]; ok {
		details.RetainerPaymentDeductionIDs = relationshipIDList(rel)
	}

	return details
}

func renderRetainerDeductionDetails(cmd *cobra.Command, details retainerDeductionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Amount != nil {
		fmt.Fprintf(out, "Amount: %s\n", formatAnyValue(details.Amount))
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.RetainerID != "" {
		fmt.Fprintf(out, "Retainer: %s\n", details.RetainerID)
	}
	if len(details.RetainerPaymentDeductionIDs) > 0 {
		fmt.Fprintf(out, "Retainer Payment Deductions: %s\n", strings.Join(details.RetainerPaymentDeductionIDs, ", "))
	}

	return nil
}
