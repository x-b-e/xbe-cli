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

type doCustomerRetainersCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	Customer                           string
	Broker                             string
	Status                             string
	TerminatedOn                       string
	MaximumExpectedDailyHours          string
	MaximumTravelMinutes               string
	BillableTravelMinutesPerTravelMile string
	FileAttachmentIDs                  []string
}

func newDoCustomerRetainersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer retainer",
		Long: `Create a customer retainer.

Required flags:
  --customer   Customer ID (buyer) (required)
  --broker     Broker ID (seller) (required)

Optional flags:
  --status                             Status (editing, active, terminated, expired, closed)
  --terminated-on                      Termination date (YYYY-MM-DD, required when status is terminated)
  --maximum-expected-daily-hours       Maximum expected daily hours
  --maximum-travel-minutes             Maximum travel minutes
  --billable-travel-minutes-per-travel-mile  Billable travel minutes per travel mile
  --file-attachment-ids                File attachment IDs (comma-separated or repeated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a customer retainer
  xbe do customer-retainers create --customer 123 --broker 456 --status editing

  # Create with travel settings
  xbe do customer-retainers create --customer 123 --broker 456 \
    --maximum-expected-daily-hours 10 --maximum-travel-minutes 90 \
    --billable-travel-minutes-per-travel-mile 2.5

  # JSON output
  xbe do customer-retainers create --customer 123 --broker 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCustomerRetainersCreate,
	}
	initDoCustomerRetainersCreateFlags(cmd)
	return cmd
}

func init() {
	doCustomerRetainersCmd.AddCommand(newDoCustomerRetainersCreateCmd())
}

func initDoCustomerRetainersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("customer", "", "Customer ID (buyer) (required)")
	cmd.Flags().String("broker", "", "Broker ID (seller) (required)")
	cmd.Flags().String("status", "", "Status (editing, active, terminated, expired, closed)")
	cmd.Flags().String("terminated-on", "", "Termination date (YYYY-MM-DD)")
	cmd.Flags().String("maximum-expected-daily-hours", "", "Maximum expected daily hours")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("billable-travel-minutes-per-travel-mile", "", "Billable travel minutes per travel mile")
	cmd.Flags().StringSlice("file-attachment-ids", nil, "File attachment IDs (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("customer")
	_ = cmd.MarkFlagRequired("broker")
}

func runDoCustomerRetainersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCustomerRetainersCreateOptions(cmd)
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
	setStringAttrIfPresent(attributes, "status", opts.Status)
	setStringAttrIfPresent(attributes, "terminated-on", opts.TerminatedOn)
	setStringAttrIfPresent(attributes, "maximum-expected-daily-hours", opts.MaximumExpectedDailyHours)
	setStringAttrIfPresent(attributes, "maximum-travel-minutes", opts.MaximumTravelMinutes)
	setStringAttrIfPresent(attributes, "billable-travel-minutes-per-travel-mile", opts.BillableTravelMinutesPerTravelMile)

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

	fileAttachmentData := buildRelationshipData("file-attachments", opts.FileAttachmentIDs)
	if len(fileAttachmentData) > 0 {
		relationships["file-attachments"] = map[string]any{"data": fileAttachmentData}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "customer-retainers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/customer-retainers", jsonBody)
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

	row := buildCustomerRetainerRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created customer retainer %s\n", row.ID)
	return nil
}

func parseDoCustomerRetainersCreateOptions(cmd *cobra.Command) (doCustomerRetainersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	status, _ := cmd.Flags().GetString("status")
	terminatedOn, _ := cmd.Flags().GetString("terminated-on")
	maxExpectedDailyHours, _ := cmd.Flags().GetString("maximum-expected-daily-hours")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	billableTravelMinutesPerTravelMile, _ := cmd.Flags().GetString("billable-travel-minutes-per-travel-mile")
	fileAttachmentIDs, _ := cmd.Flags().GetStringSlice("file-attachment-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerRetainersCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		Customer:                           customer,
		Broker:                             broker,
		Status:                             status,
		TerminatedOn:                       terminatedOn,
		MaximumExpectedDailyHours:          maxExpectedDailyHours,
		MaximumTravelMinutes:               maximumTravelMinutes,
		BillableTravelMinutesPerTravelMile: billableTravelMinutesPerTravelMile,
		FileAttachmentIDs:                  fileAttachmentIDs,
	}, nil
}
