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

type doCustomerCommitmentsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	Customer            string
	Broker              string
	Status              string
	Label               string
	Notes               string
	Tons                string
	TonsPerShift        string
	ExternalJobNumber   string
	TruckScope          string
	PrecedingCommitment string
}

func newDoCustomerCommitmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer commitment",
		Long: `Update a customer commitment.

Optional:
  --customer               Customer ID
  --broker                 Broker ID
  --status                 Status (editing, active, inactive)
  --label                  Commitment label
  --notes                  Commitment notes
  --tons                   Committed tons
  --tons-per-shift          Committed tons per shift
  --external-job-number     External job number
  --truck-scope             Truck scope ID
  --preceding-commitment    Preceding customer commitment ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update commitment status
  xbe do customer-commitments update 123 --status inactive

  # Update tonnage
  xbe do customer-commitments update 123 --tons 1200 --tons-per-shift 240

  # Update external job number
  xbe do customer-commitments update 123 --external-job-number JOB-42`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerCommitmentsUpdate,
	}
	initDoCustomerCommitmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerCommitmentsCmd.AddCommand(newDoCustomerCommitmentsUpdateCmd())
}

func initDoCustomerCommitmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("status", "", "Status (editing, active, inactive)")
	cmd.Flags().String("label", "", "Commitment label")
	cmd.Flags().String("notes", "", "Commitment notes")
	cmd.Flags().String("tons", "", "Committed tons")
	cmd.Flags().String("tons-per-shift", "", "Committed tons per shift")
	cmd.Flags().String("external-job-number", "", "External job number")
	cmd.Flags().String("truck-scope", "", "Truck scope ID")
	cmd.Flags().String("preceding-commitment", "", "Preceding customer commitment ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerCommitmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerCommitmentsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	relationships := map[string]any{}

	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("label") {
		attributes["label"] = opts.Label
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("tons") {
		attributes["tons"] = opts.Tons
	}
	if cmd.Flags().Changed("tons-per-shift") {
		attributes["tons-per-shift"] = opts.TonsPerShift
	}
	if cmd.Flags().Changed("external-job-number") {
		attributes["external-job-number"] = opts.ExternalJobNumber
	}

	if cmd.Flags().Changed("customer") {
		if opts.Customer == "" {
			relationships["buyer"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["buyer"] = map[string]any{
				"data": map[string]any{
					"type": "customers",
					"id":   opts.Customer,
				},
			}
		}
	}
	if cmd.Flags().Changed("broker") {
		if opts.Broker == "" {
			relationships["seller"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["seller"] = map[string]any{
				"data": map[string]any{
					"type": "brokers",
					"id":   opts.Broker,
				},
			}
		}
	}
	if cmd.Flags().Changed("truck-scope") {
		if opts.TruckScope == "" {
			relationships["truck-scope"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["truck-scope"] = map[string]any{
				"data": map[string]any{
					"type": "truck-scopes",
					"id":   opts.TruckScope,
				},
			}
		}
	}
	if cmd.Flags().Changed("preceding-commitment") {
		if opts.PrecedingCommitment == "" {
			relationships["preceding-commitment"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["preceding-commitment"] = map[string]any{
				"data": map[string]any{
					"type": "customer-commitments",
					"id":   opts.PrecedingCommitment,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "customer-commitments",
			"id":   opts.ID,
		},
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-commitments/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer commitment %s\n", row.ID)
	return nil
}

func parseDoCustomerCommitmentsUpdateOptions(cmd *cobra.Command, args []string) (doCustomerCommitmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	status, _ := cmd.Flags().GetString("status")
	label, _ := cmd.Flags().GetString("label")
	notes, _ := cmd.Flags().GetString("notes")
	tons, _ := cmd.Flags().GetString("tons")
	tonsPerShift, _ := cmd.Flags().GetString("tons-per-shift")
	externalJobNumber, _ := cmd.Flags().GetString("external-job-number")
	truckScope, _ := cmd.Flags().GetString("truck-scope")
	precedingCommitment, _ := cmd.Flags().GetString("preceding-commitment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerCommitmentsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		Customer:            customer,
		Broker:              broker,
		Status:              status,
		Label:               label,
		Notes:               notes,
		Tons:                tons,
		TonsPerShift:        tonsPerShift,
		ExternalJobNumber:   externalJobNumber,
		TruckScope:          truckScope,
		PrecedingCommitment: precedingCommitment,
	}, nil
}
