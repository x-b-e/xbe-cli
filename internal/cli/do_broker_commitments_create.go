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

type doBrokerCommitmentsCreateOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	Status     string
	Label      string
	Notes      string
	Broker     string
	Trucker    string
	TruckScope string
}

func newDoBrokerCommitmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker commitment",
		Long: `Create a broker commitment.

Required flags:
  --status     Commitment status (editing, active, inactive)
  --broker     Broker ID (buyer)
  --trucker    Trucker ID (seller)

Optional flags:
  --label        Commitment label
  --notes        Notes
  --truck-scope  Truck scope ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a broker commitment
  xbe do broker-commitments create --status active --broker 123 --trucker 456

  # Create with label and notes
  xbe do broker-commitments create --status editing --broker 123 --trucker 456 --label "Q1" --notes "Seasonal capacity"

  # Create with a truck scope
  xbe do broker-commitments create --status active --broker 123 --trucker 456 --truck-scope 789

  # Get JSON output
  xbe do broker-commitments create --status active --broker 123 --trucker 456 --json`,
		RunE: runDoBrokerCommitmentsCreate,
	}
	initDoBrokerCommitmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerCommitmentsCmd.AddCommand(newDoBrokerCommitmentsCreateCmd())
}

func initDoBrokerCommitmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Commitment status (editing, active, inactive) (required)")
	cmd.Flags().String("label", "", "Commitment label")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
	cmd.Flags().String("truck-scope", "", "Truck scope ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("status")
	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("trucker")
}

func runDoBrokerCommitmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerCommitmentsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	attributes := map[string]any{
		"status": opts.Status,
	}
	if opts.Label != "" {
		attributes["label"] = opts.Label
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}

	relationships := map[string]any{
		"buyer": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
		"seller": map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}
	if opts.TruckScope != "" {
		relationships["truck-scope"] = map[string]any{
			"data": map[string]any{
				"type": "truck-scopes",
				"id":   opts.TruckScope,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "broker-commitments",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/broker-commitments", jsonBody)
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

	row := buildBrokerCommitmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.Status != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created broker commitment %s (status: %s)\n", row.ID, row.Status)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created broker commitment %s\n", row.ID)
	return nil
}

func parseDoBrokerCommitmentsCreateOptions(cmd *cobra.Command) (doBrokerCommitmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	label, _ := cmd.Flags().GetString("label")
	notes, _ := cmd.Flags().GetString("notes")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	truckScope, _ := cmd.Flags().GetString("truck-scope")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerCommitmentsCreateOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		Status:     status,
		Label:      label,
		Notes:      notes,
		Broker:     broker,
		Trucker:    trucker,
		TruckScope: truckScope,
	}, nil
}

func buildBrokerCommitmentRowFromSingle(resp jsonAPISingleResponse) brokerCommitmentRow {
	resource := resp.Data
	row := brokerCommitmentRow{
		ID:           resource.ID,
		Status:       stringAttr(resource.Attributes, "status"),
		Label:        stringAttr(resource.Attributes, "label"),
		BrokerID:     "",
		TruckerID:    "",
		TruckScopeID: relationshipIDFromMap(resource.Relationships, "truck-scope"),
	}
	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil && rel.Data.Type == "brokers" {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil && rel.Data.Type == "truckers" {
		row.TruckerID = rel.Data.ID
	}
	if row.BrokerID == "" {
		row.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
	}
	if row.TruckerID == "" {
		row.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if row.BrokerID != "" {
		if broker, ok := included[resourceKey("brokers", row.BrokerID)]; ok {
			row.Broker = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if row.TruckerID != "" {
		if trucker, ok := included[resourceKey("truckers", row.TruckerID)]; ok {
			row.Trucker = firstNonEmpty(
				stringAttr(trucker.Attributes, "company-name"),
				stringAttr(trucker.Attributes, "name"),
			)
		}
	}

	return row
}
