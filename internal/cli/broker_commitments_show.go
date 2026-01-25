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

type brokerCommitmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerCommitmentDetails struct {
	ID                      string   `json:"id"`
	Status                  string   `json:"status,omitempty"`
	Label                   string   `json:"label,omitempty"`
	Notes                   string   `json:"notes,omitempty"`
	BrokerID                string   `json:"broker_id,omitempty"`
	BrokerName              string   `json:"broker,omitempty"`
	TruckerID               string   `json:"trucker_id,omitempty"`
	TruckerName             string   `json:"trucker,omitempty"`
	TruckScopeID            string   `json:"truck_scope_id,omitempty"`
	BuyerID                 string   `json:"buyer_id,omitempty"`
	BuyerType               string   `json:"buyer_type,omitempty"`
	SellerID                string   `json:"seller_id,omitempty"`
	SellerType              string   `json:"seller_type,omitempty"`
	CommitmentItemIDs       []string `json:"commitment_item_ids,omitempty"`
	CommitmentSimulationIDs []string `json:"commitment_simulation_ids,omitempty"`
}

func newBrokerCommitmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker commitment details",
		Long: `Show the full details of a broker commitment, including buyer (broker) and seller (trucker).

Output Fields:
  ID
  Status
  Label
  Notes
  Broker
  Trucker
  Truck Scope ID
  Buyer / Seller (type + ID)
  Commitment Item IDs
  Commitment Simulation IDs

Arguments:
  <id>    The broker commitment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a broker commitment
  xbe view broker-commitments show 123

  # Get JSON output
  xbe view broker-commitments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerCommitmentsShow,
	}
	initBrokerCommitmentsShowFlags(cmd)
	return cmd
}

func init() {
	brokerCommitmentsCmd.AddCommand(newBrokerCommitmentsShowCmd())
}

func initBrokerCommitmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerCommitmentsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerCommitmentsShowOptions(cmd)
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
		return fmt.Errorf("broker commitment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-commitments]", "status,label,notes,buyer,seller,truck-scope,commitment-items,commitment-simulations")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("include", "buyer,seller")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-commitments/"+id, query)
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

	details := buildBrokerCommitmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerCommitmentDetails(cmd, details)
}

func parseBrokerCommitmentsShowOptions(cmd *cobra.Command) (brokerCommitmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerCommitmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerCommitmentDetails(resp jsonAPISingleResponse) brokerCommitmentDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := brokerCommitmentDetails{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		Label:                   stringAttr(attrs, "label"),
		Notes:                   stringAttr(attrs, "notes"),
		BrokerID:                "",
		TruckerID:               "",
		TruckScopeID:            relationshipIDFromMap(resource.Relationships, "truck-scope"),
		CommitmentItemIDs:       relationshipIDsFromMap(resource.Relationships, "commitment-items"),
		CommitmentSimulationIDs: relationshipIDsFromMap(resource.Relationships, "commitment-simulations"),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerID = rel.Data.ID
		details.BuyerType = rel.Data.Type
	}

	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerID = rel.Data.ID
		details.SellerType = rel.Data.Type
	}
	if details.BuyerType == "brokers" {
		details.BrokerID = details.BuyerID
	}
	if details.SellerType == "truckers" {
		details.TruckerID = details.SellerID
	}
	if details.BrokerID == "" {
		details.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
	}
	if details.TruckerID == "" {
		details.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.BrokerID != "" {
		if broker, ok := included[resourceKey("brokers", details.BrokerID)]; ok {
			details.BrokerName = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if details.TruckerID != "" {
		if trucker, ok := included[resourceKey("truckers", details.TruckerID)]; ok {
			details.TruckerName = firstNonEmpty(
				stringAttr(trucker.Attributes, "company-name"),
				stringAttr(trucker.Attributes, "name"),
			)
		}
	}

	return details
}

func renderBrokerCommitmentDetails(cmd *cobra.Command, details brokerCommitmentDetails) error {
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

	if details.BrokerID != "" || details.BrokerName != "" {
		name := details.BrokerName
		if name == "" {
			name = details.BrokerID
			fmt.Fprintf(out, "Broker: %s\n", name)
		} else if details.BrokerID != "" {
			fmt.Fprintf(out, "Broker: %s (%s)\n", name, details.BrokerID)
		} else {
			fmt.Fprintf(out, "Broker: %s\n", name)
		}
	}

	if details.TruckerID != "" || details.TruckerName != "" {
		name := details.TruckerName
		if name == "" {
			name = details.TruckerID
			fmt.Fprintf(out, "Trucker: %s\n", name)
		} else if details.TruckerID != "" {
			fmt.Fprintf(out, "Trucker: %s (%s)\n", name, details.TruckerID)
		} else {
			fmt.Fprintf(out, "Trucker: %s\n", name)
		}
	}

	if details.TruckScopeID != "" {
		fmt.Fprintf(out, "Truck Scope ID: %s\n", details.TruckScopeID)
	}

	if details.BuyerID != "" {
		fmt.Fprintf(out, "Buyer: %s %s\n", details.BuyerType, details.BuyerID)
	}

	if details.SellerID != "" {
		fmt.Fprintf(out, "Seller: %s %s\n", details.SellerType, details.SellerID)
	}

	if len(details.CommitmentItemIDs) > 0 {
		fmt.Fprintf(out, "Commitment Item IDs: %s\n", strings.Join(details.CommitmentItemIDs, ", "))
	}

	if len(details.CommitmentSimulationIDs) > 0 {
		fmt.Fprintf(out, "Commitment Simulation IDs: %s\n", strings.Join(details.CommitmentSimulationIDs, ", "))
	}

	return nil
}
