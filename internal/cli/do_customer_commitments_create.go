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

type doCustomerCommitmentsCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	Customer            string
	Broker              string
	Status              string
	Label               string
	Notes               string
	Tons                string
	TonsPerShift        string
	TruckScope          string
	PrecedingCommitment string
}

func newDoCustomerCommitmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer commitment",
		Long: `Create a customer commitment.

Required:
  --customer    Customer ID
  --broker      Broker ID
  --status      Status (editing, active, inactive)

Optional:
  --label                 Commitment label
  --notes                 Commitment notes
  --tons                  Committed tons
  --tons-per-shift         Committed tons per shift
  --truck-scope            Truck scope ID
  --preceding-commitment   Preceding customer commitment ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a customer commitment
  xbe do customer-commitments create --customer 123 --broker 456 --status active

  # Create with tonnage and label
  xbe do customer-commitments create --customer 123 --broker 456 --status active \
    --tons 1000 --tons-per-shift 200 --label "Q1 commitment"

  # Create with truck scope
  xbe do customer-commitments create --customer 123 --broker 456 --status active \
    --truck-scope 789`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerCommitmentsCreate,
	}
	initDoCustomerCommitmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerCommitmentsCmd.AddCommand(newDoCustomerCommitmentsCreateCmd())
}

func initDoCustomerCommitmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("status", "", "Status (editing, active, inactive)")
	cmd.Flags().String("label", "", "Commitment label")
	cmd.Flags().String("notes", "", "Commitment notes")
	cmd.Flags().String("tons", "", "Committed tons")
	cmd.Flags().String("tons-per-shift", "", "Committed tons per shift")
	cmd.Flags().String("truck-scope", "", "Truck scope ID")
	cmd.Flags().String("preceding-commitment", "", "Preceding customer commitment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("customer")
	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("status")
}

func runDoCustomerCommitmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerCommitmentsCreateOptions(cmd)
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
	if opts.Tons != "" {
		attributes["tons"] = opts.Tons
	}
	if opts.TonsPerShift != "" {
		attributes["tons-per-shift"] = opts.TonsPerShift
	}

	relationships := map[string]any{
		"buyer": map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		},
		"seller": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
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
	if opts.PrecedingCommitment != "" {
		relationships["preceding-commitment"] = map[string]any{
			"data": map[string]any{
				"type": "customer-commitments",
				"id":   opts.PrecedingCommitment,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-commitments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/customer-commitments", jsonBody)
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

	row := buildCustomerCommitmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer commitment %s\n", row.ID)
	return nil
}

func parseDoCustomerCommitmentsCreateOptions(cmd *cobra.Command) (doCustomerCommitmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	status, _ := cmd.Flags().GetString("status")
	label, _ := cmd.Flags().GetString("label")
	notes, _ := cmd.Flags().GetString("notes")
	tons, _ := cmd.Flags().GetString("tons")
	tonsPerShift, _ := cmd.Flags().GetString("tons-per-shift")
	truckScope, _ := cmd.Flags().GetString("truck-scope")
	precedingCommitment, _ := cmd.Flags().GetString("preceding-commitment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerCommitmentsCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		Customer:            customer,
		Broker:              broker,
		Status:              status,
		Label:               label,
		Notes:               notes,
		Tons:                tons,
		TonsPerShift:        tonsPerShift,
		TruckScope:          truckScope,
		PrecedingCommitment: precedingCommitment,
	}, nil
}
