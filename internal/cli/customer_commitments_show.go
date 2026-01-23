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

type customerCommitmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type customerCommitmentDetails struct {
	ID                        string   `json:"id"`
	Status                    string   `json:"status,omitempty"`
	Label                     string   `json:"label,omitempty"`
	Notes                     string   `json:"notes,omitempty"`
	Tons                      string   `json:"tons,omitempty"`
	TonsPerShift              string   `json:"tons_per_shift,omitempty"`
	ExternalJobNumber         string   `json:"external_job_number,omitempty"`
	CreatedAt                 string   `json:"created_at,omitempty"`
	UpdatedAt                 string   `json:"updated_at,omitempty"`
	BuyerType                 string   `json:"buyer_type,omitempty"`
	BuyerID                   string   `json:"buyer_id,omitempty"`
	SellerType                string   `json:"seller_type,omitempty"`
	SellerID                  string   `json:"seller_id,omitempty"`
	CustomerID                string   `json:"customer_id,omitempty"`
	BrokerID                  string   `json:"broker_id,omitempty"`
	TruckScopeID              string   `json:"truck_scope_id,omitempty"`
	PrecedingCommitmentID     string   `json:"preceding_commitment_id,omitempty"`
	CommitmentItemIDs         []string `json:"commitment_item_ids,omitempty"`
	CommitmentSimulationIDs   []string `json:"commitment_simulation_ids,omitempty"`
	CommitmentMaterialSiteIDs []string `json:"commitment_material_site_ids,omitempty"`
	SubsequentCommitmentIDs   []string `json:"subsequent_commitment_ids,omitempty"`
}

func newCustomerCommitmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer commitment details",
		Long: `Show the full details of a customer commitment.

Output Fields:
  ID
  Status
  Label
  Notes
  Tons
  Tons Per Shift
  External Job Number
  Created At
  Updated At
  Buyer (type and ID)
  Seller (type and ID)
  Customer ID
  Broker ID
  Truck Scope ID
  Preceding Commitment ID
  Commitment Item IDs
  Commitment Simulation IDs
  Commitment Material Site IDs
  Subsequent Commitment IDs

Arguments:
  <id>    Customer commitment ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a customer commitment
  xbe view customer-commitments show 123

  # JSON output
  xbe view customer-commitments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerCommitmentsShow,
	}
	initCustomerCommitmentsShowFlags(cmd)
	return cmd
}

func init() {
	customerCommitmentsCmd.AddCommand(newCustomerCommitmentsShowCmd())
}

func initCustomerCommitmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerCommitmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerCommitmentsShowOptions(cmd)
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
		return fmt.Errorf("customer commitment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-commitments]", "status,label,notes,tons,tons-per-shift,external-job-number,created-at,updated-at,buyer,seller,customer,broker,truck-scope,preceding-commitment,commitment-items,commitment-simulations,commitment-material-sites,subsequent-commitments")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-commitments/"+id, query)
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

	details := buildCustomerCommitmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerCommitmentDetails(cmd, details)
}

func parseCustomerCommitmentsShowOptions(cmd *cobra.Command) (customerCommitmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerCommitmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCustomerCommitmentDetails(resp jsonAPISingleResponse) customerCommitmentDetails {
	attrs := resp.Data.Attributes
	details := customerCommitmentDetails{
		ID:                resp.Data.ID,
		Status:            stringAttr(attrs, "status"),
		Label:             strings.TrimSpace(stringAttr(attrs, "label")),
		Notes:             strings.TrimSpace(stringAttr(attrs, "notes")),
		Tons:              stringAttr(attrs, "tons"),
		TonsPerShift:      stringAttr(attrs, "tons-per-shift"),
		ExternalJobNumber: strings.TrimSpace(stringAttr(attrs, "external-job-number")),
		CreatedAt:         formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:         formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerType = rel.Data.Type
		details.BuyerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerType = rel.Data.Type
		details.SellerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["truck-scope"]; ok && rel.Data != nil {
		details.TruckScopeID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["preceding-commitment"]; ok && rel.Data != nil {
		details.PrecedingCommitmentID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["commitment-items"]; ok {
		details.CommitmentItemIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["commitment-simulations"]; ok {
		details.CommitmentSimulationIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["commitment-material-sites"]; ok {
		details.CommitmentMaterialSiteIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["subsequent-commitments"]; ok {
		details.SubsequentCommitmentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderCustomerCommitmentDetails(cmd *cobra.Command, details customerCommitmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Label != "" {
		fmt.Fprintf(out, "Label: %s\n", details.Label)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.Tons != "" {
		fmt.Fprintf(out, "Tons: %s\n", details.Tons)
	}
	if details.TonsPerShift != "" {
		fmt.Fprintf(out, "Tons Per Shift: %s\n", details.TonsPerShift)
	}
	if details.ExternalJobNumber != "" {
		fmt.Fprintf(out, "External Job Number: %s\n", details.ExternalJobNumber)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.BuyerType != "" && details.BuyerID != "" {
		fmt.Fprintf(out, "Buyer: %s/%s\n", details.BuyerType, details.BuyerID)
	}
	if details.SellerType != "" && details.SellerID != "" {
		fmt.Fprintf(out, "Seller: %s/%s\n", details.SellerType, details.SellerID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer ID: %s\n", details.CustomerID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.TruckScopeID != "" {
		fmt.Fprintf(out, "Truck Scope ID: %s\n", details.TruckScopeID)
	}
	if details.PrecedingCommitmentID != "" {
		fmt.Fprintf(out, "Preceding Commitment ID: %s\n", details.PrecedingCommitmentID)
	}
	if len(details.CommitmentItemIDs) > 0 {
		fmt.Fprintf(out, "Commitment Item IDs: %s\n", strings.Join(details.CommitmentItemIDs, ", "))
	}
	if len(details.CommitmentSimulationIDs) > 0 {
		fmt.Fprintf(out, "Commitment Simulation IDs: %s\n", strings.Join(details.CommitmentSimulationIDs, ", "))
	}
	if len(details.CommitmentMaterialSiteIDs) > 0 {
		fmt.Fprintf(out, "Commitment Material Site IDs: %s\n", strings.Join(details.CommitmentMaterialSiteIDs, ", "))
	}
	if len(details.SubsequentCommitmentIDs) > 0 {
		fmt.Fprintf(out, "Subsequent Commitment IDs: %s\n", strings.Join(details.SubsequentCommitmentIDs, ", "))
	}

	return nil
}
