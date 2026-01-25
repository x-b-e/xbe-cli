package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type tenderReRatesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderReRateDetails struct {
	ID                                    string   `json:"id"`
	TenderIDs                             []string `json:"tender_ids,omitempty"`
	ReRate                                bool     `json:"re_rate"`
	ReConstrain                           bool     `json:"re_constrain"`
	UpdateTimeCardQuantities              bool     `json:"update_time_card_quantities"`
	SkipUpdateTravelMinutes               bool     `json:"skip_update_travel_minutes"`
	SkipValidateCustomerTenderHourlyRates bool     `json:"skip_validate_customer_tender_hourly_rates"`
	InvoiceIDs                            []string `json:"invoice_ids,omitempty"`
	Results                               any      `json:"results,omitempty"`
	Messages                              any      `json:"messages,omitempty"`
	CreatedAt                             string   `json:"created_at,omitempty"`
	UpdatedAt                             string   `json:"updated_at,omitempty"`
}

func newTenderReRatesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender re-rate details",
		Long: `Show full details of a tender re-rate.

Output Fields:
  ID                                     Re-rate identifier
  Tender IDs                             Tenders re-rated
  Re Rate                                Re-rate tenders (true/false)
  Re Constrain                           Re-constrain tenders (true/false)
  Update Time Card Quantities            Update time card quantities (true/false)
  Skip Update Travel Minutes             Skip updating travel minutes (true/false)
  Skip Validate Customer Tender Hourly Rates  Skip rate validation (true/false)
  Invoice IDs                            Invoice IDs impacted
  Results                                Re-rate results
  Messages                               Re-rate messages
  Created At                             Creation timestamp
  Updated At                             Update timestamp

Arguments:
  <id>    Tender re-rate ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a tender re-rate
  xbe view tender-re-rates show 123

  # JSON output
  xbe view tender-re-rates show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderReRatesShow,
	}
	initTenderReRatesShowFlags(cmd)
	return cmd
}

func init() {
	tenderReRatesCmd.AddCommand(newTenderReRatesShowCmd())
}

func initTenderReRatesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderReRatesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTenderReRatesShowOptions(cmd)
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
		return fmt.Errorf("tender re-rate id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tender-re-rates]", "tender-ids,re-rate,re-constrain,update-time-card-quantities,skip-update-travel-minutes,skip-validate-customer-tender-hourly-rates,invoice-ids,results,messages,created-at,updated-at")

	body, status, err := client.Get(cmd.Context(), "/v1/sombreros/tender-re-rates/"+id, query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderTenderReRatesShowUnavailable(cmd, opts.JSON)
		}
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

	details := buildTenderReRateDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderReRateDetails(cmd, details)
}

func renderTenderReRatesShowUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), tenderReRateDetails{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Tender re-rates are write-only; show is not available.")
	return nil
}

func parseTenderReRatesShowOptions(cmd *cobra.Command) (tenderReRatesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderReRatesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderReRateDetails(resp jsonAPISingleResponse) tenderReRateDetails {
	attrs := resp.Data.Attributes
	details := tenderReRateDetails{
		ID:                                    resp.Data.ID,
		TenderIDs:                             stringSliceAttr(attrs, "tender-ids"),
		ReRate:                                boolAttr(attrs, "re-rate"),
		ReConstrain:                           boolAttr(attrs, "re-constrain"),
		UpdateTimeCardQuantities:              boolAttr(attrs, "update-time-card-quantities"),
		SkipUpdateTravelMinutes:               boolAttr(attrs, "skip-update-travel-minutes"),
		SkipValidateCustomerTenderHourlyRates: boolAttr(attrs, "skip-validate-customer-tender-hourly-rates"),
		InvoiceIDs:                            stringSliceAttr(attrs, "invoice-ids"),
		Results:                               anyAttr(attrs, "results"),
		Messages:                              anyAttr(attrs, "messages"),
		CreatedAt:                             formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                             formatDateTime(stringAttr(attrs, "updated-at")),
	}

	return details
}

func renderTenderReRateDetails(cmd *cobra.Command, details tenderReRateDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Tender IDs: %s\n", formatOptional(strings.Join(details.TenderIDs, ", ")))
	fmt.Fprintf(out, "Re Rate: %t\n", details.ReRate)
	fmt.Fprintf(out, "Re Constrain: %t\n", details.ReConstrain)
	fmt.Fprintf(out, "Update Time Card Quantities: %t\n", details.UpdateTimeCardQuantities)
	fmt.Fprintf(out, "Skip Update Travel Minutes: %t\n", details.SkipUpdateTravelMinutes)
	fmt.Fprintf(out, "Skip Validate Customer Tender Hourly Rates: %t\n", details.SkipValidateCustomerTenderHourlyRates)
	fmt.Fprintf(out, "Invoice IDs: %s\n", formatOptional(strings.Join(details.InvoiceIDs, ", ")))
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	writeAnySection(out, "Results", details.Results)
	writeAnySection(out, "Messages", details.Messages)

	return nil
}
