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

type rateAgreementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type rateAgreementDetails struct {
	ID                            string   `json:"id"`
	Name                          string   `json:"name,omitempty"`
	Status                        string   `json:"status,omitempty"`
	CanDelete                     bool     `json:"can_delete"`
	SellerType                    string   `json:"seller_type,omitempty"`
	SellerID                      string   `json:"seller_id,omitempty"`
	BuyerType                     string   `json:"buyer_type,omitempty"`
	BuyerID                       string   `json:"buyer_id,omitempty"`
	RateIDs                       []string `json:"rate_ids,omitempty"`
	ShiftSetTimeCardConstraintIDs []string `json:"shift_set_time_card_constraint_ids,omitempty"`
}

func newRateAgreementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show rate agreement details",
		Long: `Show the full details of a specific rate agreement.

Output Fields:
  ID                               Rate agreement identifier
  Name                             Rate agreement name
  Status                           Status (active/inactive)
  Can Delete                       Whether the rate agreement can be deleted
  Seller                           Seller type and ID
  Buyer                            Buyer type and ID
  Rate IDs                         Associated rate IDs
  Shift Set Time Card Constraint IDs  Associated constraint IDs

Arguments:
  <id>    The rate agreement ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a rate agreement
  xbe view rate-agreements show 123

  # JSON output
  xbe view rate-agreements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRateAgreementsShow,
	}
	initRateAgreementsShowFlags(cmd)
	return cmd
}

func init() {
	rateAgreementsCmd.AddCommand(newRateAgreementsShowCmd())
}

func initRateAgreementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAgreementsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRateAgreementsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("rate agreement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[rate-agreements]", "name,status,can-delete,seller,buyer,rates,shift-set-time-card-constraints")

	body, _, err := client.Get(cmd.Context(), "/v1/rate-agreements/"+id, query)
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

	details := buildRateAgreementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRateAgreementDetails(cmd, details)
}

func parseRateAgreementsShowOptions(cmd *cobra.Command) (rateAgreementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAgreementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRateAgreementDetails(resp jsonAPISingleResponse) rateAgreementDetails {
	attrs := resp.Data.Attributes
	details := rateAgreementDetails{
		ID:        resp.Data.ID,
		Name:      strings.TrimSpace(stringAttr(attrs, "name")),
		Status:    stringAttr(attrs, "status"),
		CanDelete: boolAttr(attrs, "can-delete"),
	}

	if rel, ok := resp.Data.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerType = rel.Data.Type
		details.SellerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerType = rel.Data.Type
		details.BuyerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["rates"]; ok {
		details.RateIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["shift-set-time-card-constraints"]; ok {
		details.ShiftSetTimeCardConstraintIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderRateAgreementDetails(cmd *cobra.Command, details rateAgreementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	fmt.Fprintf(out, "Can Delete: %s\n", formatBool(details.CanDelete))
	if details.SellerType != "" && details.SellerID != "" {
		fmt.Fprintf(out, "Seller: %s/%s\n", details.SellerType, details.SellerID)
	}
	if details.BuyerType != "" && details.BuyerID != "" {
		fmt.Fprintf(out, "Buyer: %s/%s\n", details.BuyerType, details.BuyerID)
	}
	if len(details.RateIDs) > 0 {
		fmt.Fprintf(out, "Rate IDs: %s\n", strings.Join(details.RateIDs, ", "))
	}
	if len(details.ShiftSetTimeCardConstraintIDs) > 0 {
		fmt.Fprintf(out, "Shift Set Time Card Constraint IDs: %s\n", strings.Join(details.ShiftSetTimeCardConstraintIDs, ", "))
	}

	return nil
}
