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

type doBrokerRetainersCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
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

func newDoBrokerRetainersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a broker retainer",
		Long: `Create a broker retainer.

Required flags:
  --broker or --buyer   Broker (Type|ID) to set as buyer
  --trucker or --seller Trucker (Type|ID) to set as seller

Optional flags:
  --status                              Retainer status (editing/active/terminated/expired/closed)
  --terminated-on                       Termination date (YYYY-MM-DD, required when status is terminated)
  --maximum-expected-daily-hours        Maximum expected daily hours
  --maximum-travel-minutes              Maximum travel minutes
  --billable-travel-minutes-per-travel-mile  Billable travel minutes per travel mile`,
		Example: `  # Create a broker retainer
  xbe do broker-retainers create --broker 123 --trucker 456

  # Create with status and travel settings
  xbe do broker-retainers create --broker 123 --trucker 456 \
    --status active \
    --maximum-travel-minutes 90 \
    --billable-travel-minutes-per-travel-mile 2

  # Output JSON
  xbe do broker-retainers create --broker 123 --trucker 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoBrokerRetainersCreate,
	}
	initDoBrokerRetainersCreateFlags(cmd)
	return cmd
}

func init() {
	doBrokerRetainersCmd.AddCommand(newDoBrokerRetainersCreateCmd())
}

func initDoBrokerRetainersCreateFlags(cmd *cobra.Command) {
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

func runDoBrokerRetainersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBrokerRetainersCreateOptions(cmd)
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

	buyerType := ""
	buyerID := ""
	if strings.TrimSpace(opts.Buyer) != "" {
		parsedType, parsedID, err := parseOrganization(opts.Buyer)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		buyerType = parsedType
		buyerID = parsedID
	} else if strings.TrimSpace(opts.Broker) != "" {
		buyerType = "brokers"
		buyerID = strings.TrimSpace(opts.Broker)
	}

	sellerType := ""
	sellerID := ""
	if strings.TrimSpace(opts.Seller) != "" {
		parsedType, parsedID, err := parseOrganization(opts.Seller)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		sellerType = parsedType
		sellerID = parsedID
	} else if strings.TrimSpace(opts.Trucker) != "" {
		sellerType = "truckers"
		sellerID = strings.TrimSpace(opts.Trucker)
	}

	if buyerID == "" {
		err := fmt.Errorf("--broker or --buyer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if sellerID == "" {
		err := fmt.Errorf("--trucker or --seller is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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
		attributes["billable-travel-minutes-per-travel-mile"] = opts.BillableTravelMinutesPerTravelMile
	}

	relationships := map[string]any{
		"buyer": map[string]any{
			"data": map[string]any{
				"type": buyerType,
				"id":   buyerID,
			},
		},
		"seller": map[string]any{
			"data": map[string]any{
				"type": sellerType,
				"id":   sellerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "broker-retainers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/broker-retainers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created broker retainer %s\n", resp.Data.ID)
	return nil
}

func parseDoBrokerRetainersCreateOptions(cmd *cobra.Command) (doBrokerRetainersCreateOptions, error) {
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

	return doBrokerRetainersCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
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
