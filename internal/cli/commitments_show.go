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

type commitmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type commitmentDetails struct {
	ID                      string   `json:"id"`
	Status                  string   `json:"status,omitempty"`
	Label                   string   `json:"label,omitempty"`
	Notes                   string   `json:"notes,omitempty"`
	BuyerType               string   `json:"buyer_type,omitempty"`
	BuyerID                 string   `json:"buyer_id,omitempty"`
	SellerType              string   `json:"seller_type,omitempty"`
	SellerID                string   `json:"seller_id,omitempty"`
	TruckScopeType          string   `json:"truck_scope_type,omitempty"`
	TruckScopeID            string   `json:"truck_scope_id,omitempty"`
	CommitmentItemIDs       []string `json:"commitment_item_ids,omitempty"`
	CommitmentSimulationIDs []string `json:"commitment_simulation_ids,omitempty"`
}

func newCommitmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show commitment details",
		Long: `Show the full details of a commitment.

Output Fields:
  ID
  Status
  Label
  Notes
  Buyer (type/id)
  Seller (type/id)
  Truck Scope (type/id)
  Commitment Items
  Commitment Simulations

Arguments:
  <id>    The commitment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a commitment
  xbe view commitments show 123

  # Output as JSON
  xbe view commitments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCommitmentsShow,
	}
	initCommitmentsShowFlags(cmd)
	return cmd
}

func init() {
	commitmentsCmd.AddCommand(newCommitmentsShowCmd())
}

func initCommitmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommitmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCommitmentsShowOptions(cmd)
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
		return fmt.Errorf("commitment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[commitments]", "status,label,notes,buyer,seller,truck-scope,commitment-items,commitment-simulations")

	body, _, err := client.Get(cmd.Context(), "/v1/commitments/"+id, query)
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

	details := buildCommitmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCommitmentDetails(cmd, details)
}

func parseCommitmentsShowOptions(cmd *cobra.Command) (commitmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commitmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCommitmentDetails(resp jsonAPISingleResponse) commitmentDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := commitmentDetails{
		ID:     resource.ID,
		Status: stringAttr(attrs, "status"),
		Label:  strings.TrimSpace(stringAttr(attrs, "label")),
		Notes:  strings.TrimSpace(stringAttr(attrs, "notes")),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerType = rel.Data.Type
		details.BuyerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerType = rel.Data.Type
		details.SellerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["truck-scope"]; ok && rel.Data != nil {
		details.TruckScopeType = rel.Data.Type
		details.TruckScopeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["commitment-items"]; ok {
		details.CommitmentItemIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["commitment-simulations"]; ok {
		details.CommitmentSimulationIDs = relationshipIDList(rel)
	}

	return details
}

func renderCommitmentDetails(cmd *cobra.Command, details commitmentDetails) error {
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
	if details.BuyerType != "" || details.BuyerID != "" {
		fmt.Fprintf(out, "Buyer: %s\n", formatTypeID(details.BuyerType, details.BuyerID))
	}
	if details.SellerType != "" || details.SellerID != "" {
		fmt.Fprintf(out, "Seller: %s\n", formatTypeID(details.SellerType, details.SellerID))
	}
	if details.TruckScopeType != "" || details.TruckScopeID != "" {
		fmt.Fprintf(out, "Truck Scope: %s\n", formatTypeID(details.TruckScopeType, details.TruckScopeID))
	}

	if len(details.CommitmentItemIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Commitment Items (%d):\n", len(details.CommitmentItemIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.CommitmentItemIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if len(details.CommitmentSimulationIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Commitment Simulations (%d):\n", len(details.CommitmentSimulationIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.CommitmentSimulationIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}
