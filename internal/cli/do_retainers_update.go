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

type doRetainersUpdateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	ID                               string
	Buyer                            string
	Seller                           string
	Status                           string
	TerminatedOn                     string
	MaximumExpectedDailyHours        string
	MaximumTravelMinutes             string
	BillableTravelMinutesPerTravelMi string
}

func newDoRetainersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a retainer",
		Long: `Update an existing retainer.

Only the fields you specify will be updated. Fields not provided remain unchanged.

Arguments:
  <id>    The retainer ID (required)

Flags:
  --status                             Retainer status (editing, active, terminated, expired, closed)
  --terminated-on                      Termination date (YYYY-MM-DD)
  --maximum-expected-daily-hours        Maximum expected daily hours
  --maximum-travel-minutes              Maximum travel minutes
  --billable-travel-minutes-per-travel-mile  Billable travel minutes per travel mile
  --buyer                              Buyer organization in Type|ID format
  --seller                             Seller organization in Type|ID format

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update travel limits
  xbe do retainers update 456 --maximum-travel-minutes 90

  # Update status
  xbe do retainers update 456 --status active

  # Update buyer and seller
  xbe do retainers update 456 --buyer Broker|123 --seller Trucker|456

  # Output as JSON
  xbe do retainers update 456 --status active --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoRetainersUpdate,
	}
	initDoRetainersUpdateFlags(cmd)
	return cmd
}

func init() {
	doRetainersCmd.AddCommand(newDoRetainersUpdateCmd())
}

func initDoRetainersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Retainer status")
	cmd.Flags().String("terminated-on", "", "Termination date (YYYY-MM-DD)")
	cmd.Flags().String("maximum-expected-daily-hours", "", "Maximum expected daily hours")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("billable-travel-minutes-per-travel-mile", "", "Billable travel minutes per travel mile")
	cmd.Flags().String("buyer", "", "Buyer organization in Type|ID format")
	cmd.Flags().String("seller", "", "Seller organization in Type|ID format")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRetainersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoRetainersUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("retainer id is required")
	}

	attributes := map[string]any{}

	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("terminated-on") {
		attributes["terminated-on"] = opts.TerminatedOn
	}
	if cmd.Flags().Changed("maximum-expected-daily-hours") {
		attributes["maximum-expected-daily-hours"] = opts.MaximumExpectedDailyHours
	}
	if cmd.Flags().Changed("maximum-travel-minutes") {
		attributes["maximum-travel-minutes"] = opts.MaximumTravelMinutes
	}
	if cmd.Flags().Changed("billable-travel-minutes-per-travel-mile") {
		attributes["billable-travel-minutes-per-travel-mile"] = opts.BillableTravelMinutesPerTravelMi
	}

	relationships := map[string]any{}

	if cmd.Flags().Changed("buyer") {
		buyerType, buyerID, err := parseRetainerParty(opts.Buyer, "buyer")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["buyer"] = map[string]any{
			"data": map[string]any{
				"type": buyerType,
				"id":   buyerID,
			},
		}
	}

	if cmd.Flags().Changed("seller") {
		sellerType, sellerID, err := parseRetainerParty(opts.Seller, "seller")
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["seller"] = map[string]any{
			"data": map[string]any{
				"type": sellerType,
				"id":   sellerID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "retainers",
			"id":   id,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/retainers/"+id, jsonBody)
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

	row := buildRetainerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated retainer %s\n", row.ID)
	return nil
}

func parseDoRetainersUpdateOptions(cmd *cobra.Command, args []string) (doRetainersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	terminatedOn, _ := cmd.Flags().GetString("terminated-on")
	maximumExpectedDailyHours, _ := cmd.Flags().GetString("maximum-expected-daily-hours")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	billableTravelMinutesPerTravelMi, _ := cmd.Flags().GetString("billable-travel-minutes-per-travel-mile")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainersUpdateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		ID:                               args[0],
		Buyer:                            buyer,
		Seller:                           seller,
		Status:                           status,
		TerminatedOn:                     terminatedOn,
		MaximumExpectedDailyHours:        maximumExpectedDailyHours,
		MaximumTravelMinutes:             maximumTravelMinutes,
		BillableTravelMinutesPerTravelMi: billableTravelMinutesPerTravelMi,
	}, nil
}
