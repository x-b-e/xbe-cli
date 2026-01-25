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

type lineupDispatchesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupDispatchDetails struct {
	ID                       string `json:"id"`
	LineupID                 string `json:"lineup_id,omitempty"`
	CreatedByID              string `json:"created_by_id,omitempty"`
	Comment                  string `json:"comment,omitempty"`
	AutoOfferCustomerTenders bool   `json:"auto_offer_customer_tenders"`
	AutoOfferTruckerTenders  bool   `json:"auto_offer_trucker_tenders"`
	AutoAcceptTruckerTenders bool   `json:"auto_accept_trucker_tenders"`
	IsFulfilled              bool   `json:"is_fulfilled"`
	IsFulfilling             bool   `json:"is_fulfilling"`
	FulfillmentCount         int    `json:"fulfillment_count"`
	FulfilledWithAutoOffer   bool   `json:"fulfilled_with_auto_offer"`
	FulfillmentResult        any    `json:"fulfillment_result,omitempty"`
}

func newLineupDispatchesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup dispatch details",
		Long: `Show the full details of a lineup dispatch.

Output Fields:
  ID                          Lineup dispatch identifier
  Lineup                      Lineup ID
  Created By                  Creator user ID
  Comment                     Dispatch comment
  Auto Offer Customer Tenders Auto-offer customer tenders
  Auto Offer Trucker Tenders  Auto-offer trucker tenders
  Auto Accept Trucker Tenders Auto-accept trucker tenders
  Is Fulfilled                Fulfillment status
  Is Fulfilling               Fulfillment in progress
  Fulfillment Count           Number of fulfillment attempts
  Fulfilled With Auto Offer   Whether auto-offer fulfilled the dispatch
  Fulfillment Result          Last fulfillment result (if present)

Arguments:
  <id>  The lineup dispatch ID (required).`,
		Example: `  # Show a lineup dispatch
  xbe view lineup-dispatches show 123

  # JSON output
  xbe view lineup-dispatches show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupDispatchesShow,
	}
	initLineupDispatchesShowFlags(cmd)
	return cmd
}

func init() {
	lineupDispatchesCmd.AddCommand(newLineupDispatchesShowCmd())
}

func initLineupDispatchesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupDispatchesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupDispatchesShowOptions(cmd)
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
		return fmt.Errorf("lineup dispatch id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-dispatches/"+id, nil)
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

	details := buildLineupDispatchDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupDispatchDetails(cmd, details)
}

func parseLineupDispatchesShowOptions(cmd *cobra.Command) (lineupDispatchesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupDispatchesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupDispatchDetails(resp jsonAPISingleResponse) lineupDispatchDetails {
	attrs := resp.Data.Attributes
	details := lineupDispatchDetails{
		ID:                       resp.Data.ID,
		Comment:                  strings.TrimSpace(stringAttr(attrs, "comment")),
		AutoOfferCustomerTenders: boolAttr(attrs, "auto-offer-customer-tenders"),
		AutoOfferTruckerTenders:  boolAttr(attrs, "auto-offer-trucker-tenders"),
		AutoAcceptTruckerTenders: boolAttr(attrs, "auto-accept-trucker-tenders"),
		IsFulfilled:              boolAttr(attrs, "is-fulfilled"),
		IsFulfilling:             boolAttr(attrs, "is-fulfilling"),
		FulfillmentCount:         intAttr(attrs, "fulfillment-count"),
		FulfilledWithAutoOffer:   boolAttr(attrs, "fulfilled-with-auto-offer"),
		FulfillmentResult:        anyAttr(attrs, "fulfillment-result"),
	}

	if rel, ok := resp.Data.Relationships["lineup"]; ok && rel.Data != nil {
		details.LineupID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderLineupDispatchDetails(cmd *cobra.Command, details lineupDispatchDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupID != "" {
		fmt.Fprintf(out, "Lineup: %s\n", details.LineupID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.Comment)
	}

	fmt.Fprintf(out, "Auto Offer Customer Tenders: %s\n", formatBool(details.AutoOfferCustomerTenders))
	fmt.Fprintf(out, "Auto Offer Trucker Tenders: %s\n", formatBool(details.AutoOfferTruckerTenders))
	fmt.Fprintf(out, "Auto Accept Trucker Tenders: %s\n", formatBool(details.AutoAcceptTruckerTenders))
	fmt.Fprintf(out, "Is Fulfilled: %s\n", formatBool(details.IsFulfilled))
	fmt.Fprintf(out, "Is Fulfilling: %s\n", formatBool(details.IsFulfilling))
	fmt.Fprintf(out, "Fulfillment Count: %d\n", details.FulfillmentCount)
	fmt.Fprintf(out, "Fulfilled With Auto Offer: %s\n", formatBool(details.FulfilledWithAutoOffer))
	if details.FulfillmentResult != nil {
		fmt.Fprintf(out, "Fulfillment Result: %s\n", formatAny(details.FulfillmentResult))
	}

	return nil
}
