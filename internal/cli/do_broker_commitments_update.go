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

type doBrokerCommitmentsUpdateOptions struct {
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

func newDoBrokerCommitmentsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker commitment",
		Long: `Update an existing broker commitment.

Only the fields you specify will be updated. Fields not provided remain unchanged.

Updatable fields:
  --status       Commitment status (editing, active, inactive)
  --label        Commitment label (empty to clear)
  --notes        Notes (empty to clear)
  --broker       Broker ID (buyer)
  --trucker      Trucker ID (seller)
  --truck-scope  Truck scope ID (empty to clear)

Arguments:
  <id>    The broker commitment ID (required)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update status
  xbe do broker-commitments update 123 --status inactive

  # Update label and notes
  xbe do broker-commitments update 123 --label "Q2" --notes "Revised scope"

  # Update trucker
  xbe do broker-commitments update 123 --trucker 456

  # Clear truck scope
  xbe do broker-commitments update 123 --truck-scope ""

  # Get JSON output
  xbe do broker-commitments update 123 --status active --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerCommitmentsUpdate,
	}
	initDoBrokerCommitmentsUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerCommitmentsCmd.AddCommand(newDoBrokerCommitmentsUpdateCmd())
}

func initDoBrokerCommitmentsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Commitment status (editing, active, inactive)")
	cmd.Flags().String("label", "", "Commitment label")
	cmd.Flags().String("notes", "", "Notes")
	cmd.Flags().String("broker", "", "Broker ID")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("truck-scope", "", "Truck scope ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerCommitmentsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerCommitmentsUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("broker commitment id is required")
	}

	attributes := map[string]any{}
	relationships := map[string]any{}
	hasChanges := false

	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
		hasChanges = true
	}
	if cmd.Flags().Changed("label") {
		attributes["label"] = opts.Label
		hasChanges = true
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
		hasChanges = true
	}

	if cmd.Flags().Changed("broker") {
		if opts.Broker == "" {
			relationships["buyer"] = map[string]any{"data": nil}
		} else {
			relationships["buyer"] = map[string]any{
				"data": map[string]any{
					"type": "brokers",
					"id":   opts.Broker,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("trucker") {
		if opts.Trucker == "" {
			relationships["seller"] = map[string]any{"data": nil}
		} else {
			relationships["seller"] = map[string]any{
				"data": map[string]any{
					"type": "truckers",
					"id":   opts.Trucker,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("truck-scope") {
		if opts.TruckScope == "" {
			relationships["truck-scope"] = map[string]any{"data": nil}
		} else {
			relationships["truck-scope"] = map[string]any{
				"data": map[string]any{
					"type": "truck-scopes",
					"id":   opts.TruckScope,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no fields to update; specify at least one flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "broker-commitments",
		"id":   id,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-commitments/"+id, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker commitment %s\n", row.ID)
	return nil
}

func parseDoBrokerCommitmentsUpdateOptions(cmd *cobra.Command) (doBrokerCommitmentsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	label, _ := cmd.Flags().GetString("label")
	notes, _ := cmd.Flags().GetString("notes")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	truckScope, _ := cmd.Flags().GetString("truck-scope")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerCommitmentsUpdateOptions{
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
