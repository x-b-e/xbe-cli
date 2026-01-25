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

type doRetainersCreateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	Buyer                            string
	Seller                           string
	Status                           string
	TerminatedOn                     string
	MaximumExpectedDailyHours        string
	MaximumTravelMinutes             string
	BillableTravelMinutesPerTravelMi string
}

func newDoRetainersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new retainer",
		Long: `Create a new retainer.

Required flags:
  --buyer   Buyer organization in Type|ID format (required)
  --seller  Seller organization in Type|ID format (required)

Buyer/Seller combinations:
  Broker|<id>  -> Trucker|<id>  (broker retainer)
  Customer|<id> -> Broker|<id>  (customer retainer)

Optional flags:
  --status                             Retainer status (editing, active, terminated, expired, closed)
  --terminated-on                      Termination date (YYYY-MM-DD)
  --maximum-expected-daily-hours        Maximum expected daily hours
  --maximum-travel-minutes              Maximum travel minutes
  --billable-travel-minutes-per-travel-mile  Billable travel minutes per travel mile`,
		Example: `  # Create a broker retainer
  xbe do retainers create --buyer Broker|123 --seller Trucker|456 --status editing

  # Create a customer retainer
  xbe do retainers create --buyer Customer|789 --seller Broker|123 --status editing

  # Create with travel limits
  xbe do retainers create --buyer Broker|123 --seller Trucker|456 \
    --maximum-travel-minutes 90 --billable-travel-minutes-per-travel-mile 2.5

  # Output as JSON
  xbe do retainers create --buyer Broker|123 --seller Trucker|456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoRetainersCreate,
	}
	initDoRetainersCreateFlags(cmd)
	return cmd
}

func init() {
	doRetainersCmd.AddCommand(newDoRetainersCreateCmd())
}

func initDoRetainersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("buyer", "", "Buyer organization in Type|ID format (required)")
	cmd.Flags().String("seller", "", "Seller organization in Type|ID format (required)")
	cmd.Flags().String("status", "", "Retainer status")
	cmd.Flags().String("terminated-on", "", "Termination date (YYYY-MM-DD)")
	cmd.Flags().String("maximum-expected-daily-hours", "", "Maximum expected daily hours")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("billable-travel-minutes-per-travel-mile", "", "Billable travel minutes per travel mile")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoRetainersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoRetainersCreateOptions(cmd)
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

	if opts.Buyer == "" {
		err := fmt.Errorf("--buyer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Seller == "" {
		err := fmt.Errorf("--seller is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	buyerType, buyerID, err := parseRetainerParty(opts.Buyer, "buyer")
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	sellerType, sellerID, err := parseRetainerParty(opts.Seller, "seller")
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	resourceType, endpoint, err := resolveRetainerCreateTarget(buyerType, sellerType)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	setStringAttrIfPresent(attributes, "status", opts.Status)
	setStringAttrIfPresent(attributes, "terminated-on", opts.TerminatedOn)
	setStringAttrIfPresent(attributes, "maximum-expected-daily-hours", opts.MaximumExpectedDailyHours)
	setStringAttrIfPresent(attributes, "maximum-travel-minutes", opts.MaximumTravelMinutes)
	setStringAttrIfPresent(attributes, "billable-travel-minutes-per-travel-mile", opts.BillableTravelMinutesPerTravelMi)

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
			"type":          resourceType,
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

	body, _, err := client.Post(cmd.Context(), endpoint, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created retainer %s\n", row.ID)
	return nil
}

func parseDoRetainersCreateOptions(cmd *cobra.Command) (doRetainersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	status, _ := cmd.Flags().GetString("status")
	terminatedOn, _ := cmd.Flags().GetString("terminated-on")
	maximumExpectedDailyHours, _ := cmd.Flags().GetString("maximum-expected-daily-hours")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	billableTravelMinutesPerTravelMi, _ := cmd.Flags().GetString("billable-travel-minutes-per-travel-mile")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doRetainersCreateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		Buyer:                            buyer,
		Seller:                           seller,
		Status:                           status,
		TerminatedOn:                     terminatedOn,
		MaximumExpectedDailyHours:        maximumExpectedDailyHours,
		MaximumTravelMinutes:             maximumTravelMinutes,
		BillableTravelMinutesPerTravelMi: billableTravelMinutesPerTravelMi,
	}, nil
}

func parseRetainerParty(value, label string) (string, string, error) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("--%s must be in Type|ID format (e.g. Broker|123)", label)
	}
	typePart := strings.TrimSpace(parts[0])
	idPart := strings.TrimSpace(parts[1])
	if typePart == "" || idPart == "" {
		return "", "", fmt.Errorf("--%s must be in Type|ID format (e.g. Broker|123)", label)
	}

	typeKey := strings.ToLower(typePart)
	switch typeKey {
	case "broker", "brokers":
		typeKey = "brokers"
	case "customer", "customers":
		typeKey = "customers"
	case "trucker", "truckers":
		typeKey = "truckers"
	default:
		return "", "", fmt.Errorf("unsupported %s type %q (expected Broker, Customer, or Trucker)", label, typePart)
	}

	return typeKey, idPart, nil
}

func resolveRetainerCreateTarget(buyerType, sellerType string) (string, string, error) {
	switch {
	case buyerType == "brokers" && sellerType == "truckers":
		return "broker-retainers", "/v1/broker-retainers", nil
	case buyerType == "customers" && sellerType == "brokers":
		return "customer-retainers", "/v1/customer-retainers", nil
	default:
		return "", "", fmt.Errorf("invalid buyer/seller combination: %s -> %s (expected Broker->Trucker or Customer->Broker)", buyerType, sellerType)
	}
}
