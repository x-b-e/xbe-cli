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

type retainerPeriodsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type retainerPeriodDetails struct {
	ID                  string   `json:"id"`
	StartOn             string   `json:"start_on,omitempty"`
	EndOn               string   `json:"end_on,omitempty"`
	WeeklyPaymentAmount any      `json:"weekly_payment_amount,omitempty"`
	RetainerID          string   `json:"retainer_id,omitempty"`
	RetainerPaymentIDs  []string `json:"retainer_payment_ids,omitempty"`
}

func newRetainerPeriodsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show retainer period details",
		Long: `Show the full details of a retainer period.

Output Fields:
  ID              Retainer period identifier
  START ON        Start date
  END ON          End date
  WEEKLY PAYMENT  Weekly payment amount
  RETAINER        Retainer ID
  PAYMENTS        Retainer payment IDs

Arguments:
  <id>  Retainer period ID (required).`,
		Example: `  # Show a retainer period
  xbe view retainer-periods show 123

  # Output as JSON
  xbe view retainer-periods show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRetainerPeriodsShow,
	}
	initRetainerPeriodsShowFlags(cmd)
	return cmd
}

func init() {
	retainerPeriodsCmd.AddCommand(newRetainerPeriodsShowCmd())
}

func initRetainerPeriodsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerPeriodsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRetainerPeriodsShowOptions(cmd)
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
		return fmt.Errorf("retainer period id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[retainer-periods]", "start-on,end-on,weekly-payment-amount,retainer,retainer-payments")

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-periods/"+id, query)
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

	details := buildRetainerPeriodDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRetainerPeriodDetails(cmd, details)
}

func parseRetainerPeriodsShowOptions(cmd *cobra.Command) (retainerPeriodsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerPeriodsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRetainerPeriodDetails(resp jsonAPISingleResponse) retainerPeriodDetails {
	attrs := resp.Data.Attributes

	details := retainerPeriodDetails{
		ID:                  resp.Data.ID,
		StartOn:             stringAttr(attrs, "start-on"),
		EndOn:               stringAttr(attrs, "end-on"),
		WeeklyPaymentAmount: attrs["weekly-payment-amount"],
	}

	if rel, ok := resp.Data.Relationships["retainer"]; ok && rel.Data != nil {
		details.RetainerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["retainer-payments"]; ok {
		details.RetainerPaymentIDs = relationshipIDList(rel)
	}

	return details
}

func renderRetainerPeriodDetails(cmd *cobra.Command, details retainerPeriodDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EndOn != "" {
		fmt.Fprintf(out, "End On: %s\n", details.EndOn)
	}
	if details.WeeklyPaymentAmount != nil {
		fmt.Fprintf(out, "Weekly Payment: %s\n", formatAnyValue(details.WeeklyPaymentAmount))
	}
	if details.RetainerID != "" {
		fmt.Fprintf(out, "Retainer: %s\n", details.RetainerID)
	}
	if len(details.RetainerPaymentIDs) > 0 {
		fmt.Fprintf(out, "Retainer Payments: %s\n", strings.Join(details.RetainerPaymentIDs, ", "))
	}

	return nil
}
