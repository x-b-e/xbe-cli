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

type doBrokerRetainersUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	Buyer                              string
	Seller                             string
	Broker                             string
	Trucker                            string
	Status                             string
	TerminatedOn                       string
	MaximumExpectedDailyHours          string
	MaximumTravelMinutes               string
	BillableTravelMinutesPerTravelMile string
}

func newDoBrokerRetainersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a broker retainer",
		Long: `Update a broker retainer.

Optional flags:
  --buyer                                Buyer (Type|ID, e.g., Broker|123)
  --seller                               Seller (Type|ID, e.g., Trucker|456)
  --broker                               Broker ID (buyer)
  --trucker                              Trucker ID (seller)
  --status                               Retainer status (editing/active/terminated/expired/closed)
  --terminated-on                        Termination date (YYYY-MM-DD)
  --maximum-expected-daily-hours         Maximum expected daily hours
  --maximum-travel-minutes               Maximum travel minutes
  --billable-travel-minutes-per-travel-mile   Billable travel minutes per travel mile`,
		Example: `  # Update status
  xbe do broker-retainers update 123 --status active

  # Update travel settings
  xbe do broker-retainers update 123 --maximum-travel-minutes 90`,
		Args: cobra.ExactArgs(1),
		RunE: runDoBrokerRetainersUpdate,
	}
	initDoBrokerRetainersUpdateFlags(cmd)
	return cmd
}

func init() {
	doBrokerRetainersCmd.AddCommand(newDoBrokerRetainersUpdateCmd())
}

func initDoBrokerRetainersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("buyer", "", "Buyer (Type|ID, e.g., Broker|123)")
	cmd.Flags().String("seller", "", "Seller (Type|ID, e.g., Trucker|456)")
	cmd.Flags().String("broker", "", "Broker ID (buyer)")
	cmd.Flags().String("trucker", "", "Trucker ID (seller)")
	cmd.Flags().String("status", "", "Retainer status (editing/active/terminated/expired/closed)")
	cmd.Flags().String("terminated-on", "", "Termination date (YYYY-MM-DD)")
	cmd.Flags().String("maximum-expected-daily-hours", "", "Maximum expected daily hours")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("billable-travel-minutes-per-travel-mile", "", "Billable travel minutes per travel mile")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBrokerRetainersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoBrokerRetainersUpdateOptions(cmd, args)
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
		attributes["billable-travel-minutes-per-travel-mile"] = opts.BillableTravelMinutesPerTravelMile
	}

	buyerType := ""
	buyerID := ""
	if cmd.Flags().Changed("buyer") {
		parsedType, parsedID, err := parseOrganization(opts.Buyer)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		buyerType = parsedType
		buyerID = parsedID
	} else if cmd.Flags().Changed("broker") {
		buyerType = "brokers"
		buyerID = strings.TrimSpace(opts.Broker)
	}

	sellerType := ""
	sellerID := ""
	if cmd.Flags().Changed("seller") {
		parsedType, parsedID, err := parseOrganization(opts.Seller)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		sellerType = parsedType
		sellerID = parsedID
	} else if cmd.Flags().Changed("trucker") {
		sellerType = "truckers"
		sellerID = strings.TrimSpace(opts.Trucker)
	}

	if buyerID != "" {
		relationships["buyer"] = map[string]any{
			"data": map[string]any{
				"type": buyerType,
				"id":   buyerID,
			},
		}
	}
	if sellerID != "" {
		relationships["seller"] = map[string]any{
			"data": map[string]any{
				"type": sellerType,
				"id":   sellerID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "broker-retainers",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/broker-retainers/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := brokerRetainerRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated broker retainer %s\n", resp.Data.ID)
	return nil
}

func parseDoBrokerRetainersUpdateOptions(cmd *cobra.Command, args []string) (doBrokerRetainersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	status, _ := cmd.Flags().GetString("status")
	terminatedOn, _ := cmd.Flags().GetString("terminated-on")
	maximumExpectedDailyHours, _ := cmd.Flags().GetString("maximum-expected-daily-hours")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	billableTravelMinutesPerTravelMile, _ := cmd.Flags().GetString("billable-travel-minutes-per-travel-mile")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBrokerRetainersUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		Buyer:                              buyer,
		Seller:                             seller,
		Broker:                             broker,
		Trucker:                            trucker,
		Status:                             status,
		TerminatedOn:                       terminatedOn,
		MaximumExpectedDailyHours:          maximumExpectedDailyHours,
		MaximumTravelMinutes:               maximumTravelMinutes,
		BillableTravelMinutesPerTravelMile: billableTravelMinutesPerTravelMile,
	}, nil
}
