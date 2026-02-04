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

type retainerPaymentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type retainerPaymentDetails struct {
	ID                          string   `json:"id"`
	Status                      string   `json:"status,omitempty"`
	Amount                      string   `json:"amount,omitempty"`
	CreatedOn                   string   `json:"created_on,omitempty"`
	PayOn                       string   `json:"pay_on,omitempty"`
	Kind                        string   `json:"kind,omitempty"`
	RetainerID                  string   `json:"retainer_id,omitempty"`
	RetainerPeriodID            string   `json:"retainer_period_id,omitempty"`
	RetainerPaymentDeductionIDs []string `json:"retainer_payment_deduction_ids,omitempty"`
	RetainerType                string   `json:"retainer_type,omitempty"`
	BuyerID                     string   `json:"buyer_id,omitempty"`
	BuyerType                   string   `json:"buyer_type,omitempty"`
	SellerID                    string   `json:"seller_id,omitempty"`
	SellerType                  string   `json:"seller_type,omitempty"`
}

func newRetainerPaymentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show retainer payment details",
		Long: `Show the full details of a retainer payment.

Output Fields:
  ID
  Status
  Amount
  Created On
  Pay On
  Kind
  Retainer ID
  Retainer Period ID
  Retainer Type
  Buyer / Seller (type + ID)
  Retainer Payment Deduction IDs

Arguments:
  <id>    The retainer payment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a retainer payment
  xbe view retainer-payments show 123

  # Get JSON output
  xbe view retainer-payments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRetainerPaymentsShow,
	}
	initRetainerPaymentsShowFlags(cmd)
	return cmd
}

func init() {
	retainerPaymentsCmd.AddCommand(newRetainerPaymentsShowCmd())
}

func initRetainerPaymentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerPaymentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseRetainerPaymentsShowOptions(cmd)
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
		return fmt.Errorf("retainer payment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[retainer-payments]", "status,amount,created-on,pay-on,kind,retainer-period,retainer,retainer-payment-deductions")
	query.Set("fields[retainers]", "type,buyer,seller")
	query.Set("include", "retainer")

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-payments/"+id, query)
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

	details := buildRetainerPaymentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRetainerPaymentDetails(cmd, details)
}

func parseRetainerPaymentsShowOptions(cmd *cobra.Command) (retainerPaymentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerPaymentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRetainerPaymentDetails(resp jsonAPISingleResponse) retainerPaymentDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := retainerPaymentDetails{
		ID:                          resource.ID,
		Status:                      stringAttr(attrs, "status"),
		Amount:                      stringAttr(attrs, "amount"),
		CreatedOn:                   formatDate(stringAttr(attrs, "created-on")),
		PayOn:                       formatDate(stringAttr(attrs, "pay-on")),
		Kind:                        stringAttr(attrs, "kind"),
		RetainerID:                  relationshipIDFromMap(resource.Relationships, "retainer"),
		RetainerPeriodID:            relationshipIDFromMap(resource.Relationships, "retainer-period"),
		RetainerPaymentDeductionIDs: relationshipIDsFromMap(resource.Relationships, "retainer-payment-deductions"),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.RetainerID != "" {
		if retainer, ok := included[resourceKey("retainers", details.RetainerID)]; ok {
			details.RetainerType = resolveRetainerType(retainer.Attributes)

			if rel, ok := retainer.Relationships["buyer"]; ok && rel.Data != nil {
				details.BuyerID = rel.Data.ID
				details.BuyerType = rel.Data.Type
			}
			if rel, ok := retainer.Relationships["seller"]; ok && rel.Data != nil {
				details.SellerID = rel.Data.ID
				details.SellerType = rel.Data.Type
			}
		}
	}

	return details
}

func resolveRetainerType(attrs map[string]any) string {
	retainerType := stringAttr(attrs, "type")
	if retainerType == "" {
		retainerType = stringAttr(attrs, "retainer-type")
	}
	return retainerType
}

func renderRetainerPaymentDetails(cmd *cobra.Command, details retainerPaymentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Amount != "" {
		fmt.Fprintf(out, "Amount: %s\n", details.Amount)
	}
	if details.CreatedOn != "" {
		fmt.Fprintf(out, "Created On: %s\n", details.CreatedOn)
	}
	if details.PayOn != "" {
		fmt.Fprintf(out, "Pay On: %s\n", details.PayOn)
	}
	if details.Kind != "" {
		fmt.Fprintf(out, "Kind: %s\n", details.Kind)
	}
	if details.RetainerID != "" {
		fmt.Fprintf(out, "Retainer ID: %s\n", details.RetainerID)
	}
	if details.RetainerPeriodID != "" {
		fmt.Fprintf(out, "Retainer Period ID: %s\n", details.RetainerPeriodID)
	}
	if details.RetainerType != "" {
		fmt.Fprintf(out, "Retainer Type: %s\n", details.RetainerType)
	}
	if details.BuyerID != "" {
		buyerType := details.BuyerType
		if buyerType == "" {
			buyerType = "buyer"
		}
		fmt.Fprintf(out, "Buyer: %s %s\n", buyerType, details.BuyerID)
	}
	if details.SellerID != "" {
		sellerType := details.SellerType
		if sellerType == "" {
			sellerType = "seller"
		}
		fmt.Fprintf(out, "Seller: %s %s\n", sellerType, details.SellerID)
	}
	if len(details.RetainerPaymentDeductionIDs) > 0 {
		fmt.Fprintf(out, "Retainer Payment Deduction IDs: %s\n", strings.Join(details.RetainerPaymentDeductionIDs, ", "))
	}

	return nil
}
