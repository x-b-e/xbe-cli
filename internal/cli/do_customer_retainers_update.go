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

type doCustomerRetainersUpdateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	ID                                 string
	Customer                           string
	Broker                             string
	Status                             string
	TerminatedOn                       string
	MaximumExpectedDailyHours          string
	MaximumTravelMinutes               string
	BillableTravelMinutesPerTravelMile string
	FileAttachmentIDs                  []string
}

func newDoCustomerRetainersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a customer retainer",
		Long: `Update a customer retainer.

Provide the retainer ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --status                             Status (editing, active, terminated, expired, closed)
  --terminated-on                      Termination date (YYYY-MM-DD)
  --maximum-expected-daily-hours       Maximum expected daily hours
  --maximum-travel-minutes             Maximum travel minutes
  --billable-travel-minutes-per-travel-mile  Billable travel minutes per travel mile
  --customer                           Customer ID (buyer)
  --broker                             Broker ID (seller)
  --file-attachment-ids                File attachment IDs (comma-separated or repeated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update retainer settings
  xbe do customer-retainers update 123 --maximum-expected-daily-hours 10 --maximum-travel-minutes 90

  # Update customer and broker
  xbe do customer-retainers update 123 --customer 456 --broker 789

  # JSON output
  xbe do customer-retainers update 123 --status active --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoCustomerRetainersUpdate,
	}
	initDoCustomerRetainersUpdateFlags(cmd)
	return cmd
}

func init() {
	doCustomerRetainersCmd.AddCommand(newDoCustomerRetainersUpdateCmd())
}

func initDoCustomerRetainersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Status (editing, active, terminated, expired, closed)")
	cmd.Flags().String("terminated-on", "", "Termination date (YYYY-MM-DD)")
	cmd.Flags().String("maximum-expected-daily-hours", "", "Maximum expected daily hours")
	cmd.Flags().String("maximum-travel-minutes", "", "Maximum travel minutes")
	cmd.Flags().String("billable-travel-minutes-per-travel-mile", "", "Billable travel minutes per travel mile")
	cmd.Flags().String("customer", "", "Customer ID (buyer)")
	cmd.Flags().String("broker", "", "Broker ID (seller)")
	cmd.Flags().StringSlice("file-attachment-ids", nil, "File attachment IDs (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoCustomerRetainersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoCustomerRetainersUpdateOptions(cmd, args)
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

	relationships := map[string]any{}
	if cmd.Flags().Changed("customer") {
		if strings.TrimSpace(opts.Customer) == "" {
			return fmt.Errorf("--customer cannot be empty")
		}
		relationships["buyer"] = map[string]any{
			"data": map[string]any{
				"type": "customers",
				"id":   opts.Customer,
			},
		}
	}
	if cmd.Flags().Changed("broker") {
		if strings.TrimSpace(opts.Broker) == "" {
			return fmt.Errorf("--broker cannot be empty")
		}
		relationships["seller"] = map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		}
	}
	if cmd.Flags().Changed("file-attachment-ids") {
		fileAttachmentData := buildRelationshipData("file-attachments", opts.FileAttachmentIDs)
		if len(fileAttachmentData) == 0 {
			relationships["file-attachments"] = map[string]any{"data": []any{}}
		} else {
			relationships["file-attachments"] = map[string]any{"data": fileAttachmentData}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "customer-retainers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/customer-retainers/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated customer retainer %s\n", row.ID)
	return nil
}

func parseDoCustomerRetainersUpdateOptions(cmd *cobra.Command, args []string) (doCustomerRetainersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	terminatedOn, _ := cmd.Flags().GetString("terminated-on")
	maxExpectedDailyHours, _ := cmd.Flags().GetString("maximum-expected-daily-hours")
	maximumTravelMinutes, _ := cmd.Flags().GetString("maximum-travel-minutes")
	billableTravelMinutesPerTravelMile, _ := cmd.Flags().GetString("billable-travel-minutes-per-travel-mile")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	fileAttachmentIDs, _ := cmd.Flags().GetStringSlice("file-attachment-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCustomerRetainersUpdateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		ID:                                 args[0],
		Status:                             status,
		TerminatedOn:                       terminatedOn,
		MaximumExpectedDailyHours:          maxExpectedDailyHours,
		MaximumTravelMinutes:               maximumTravelMinutes,
		BillableTravelMinutesPerTravelMile: billableTravelMinutesPerTravelMile,
		Customer:                           customer,
		Broker:                             broker,
		FileAttachmentIDs:                  fileAttachmentIDs,
	}, nil
}
